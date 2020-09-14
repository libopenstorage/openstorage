package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/pkg/auth"
	osecrets "github.com/libopenstorage/openstorage/pkg/auth/secrets"
	"github.com/libopenstorage/openstorage/volume"
	volumedrivers "github.com/libopenstorage/openstorage/volume/drivers"
	lsecrets "github.com/libopenstorage/secrets"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// PVCNameLabelKey is used for kubernetes auth provider indicating the name of PVC
	PVCNameLabelKey = "pvc"
	// PVCNamespaceLabelKey is used for kubernetes auth provider indicating the namespace of the PVC
	PVCNamespaceLabelKey = "namespace"
)

// NewAuthMiddleware returns a negroni implementation of an http middleware
// which will intercept the management APIs
func NewAuthMiddleware() *authMiddleware {
	return &authMiddleware{}
}

type authMiddleware struct {
}

// newSecurityMiddleware based on auth configuration returns SecurityHandler or just
func newSecurityMiddleware(authenticators map[string]auth.Authenticator) func(next http.HandlerFunc) http.HandlerFunc {
	if len(authenticators) > 0 {
		return func(next http.HandlerFunc) http.HandlerFunc {
			return SecurityHandler(authenticators, next)
		}
	}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return next
	}
}

// SecurityHandler implements Authentication and Authorization check at the same time
// this functionality where not moved to separate functions because of simplicity
func SecurityHandler(authenticators map[string]auth.Authenticator, next http.HandlerFunc) http.HandlerFunc {
	if len(authenticators) == 0 {
		return next
	}

	return func(w http.ResponseWriter, r *http.Request) {
		tokenHeader := r.Header.Get("Authorization")
		tokens := strings.Split(tokenHeader, " ")

		if len(tokens) < 2 {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "Access denied token is empty")
			return
		}
		token := tokens[1]

		// Determine issuer
		issuer, err := auth.TokenIssuer(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "Access denied, %v", err)
			return
		}

		// Use http.Request context for cancellation propagation
		ctx := r.Context()

		// Authenticate user
		var claims *auth.Claims
		if authenticator, exists := authenticators[issuer]; exists {
			claims, err = authenticator.AuthenticateToken(ctx, token)

			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprintf(w, "Access denied, %v", err)
				return
			}
			if claims == nil {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprintf(w, "Access denied, wrong claims provided")
				return
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "Access denied, no authenticator for issuer %s", issuer)
			return
		}
		// Check if user has admin role to access that endpoint
		isSystemAdmin := false

		for _, role := range claims.Roles {
			if role == "system.admin" {
				isSystemAdmin = true
				break
			}
		}

		if !isSystemAdmin {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "Access denied, user must have admin access")
			return
		}

		next.ServeHTTP(w, r)
	}
}

func (a *authMiddleware) createWithAuth(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	fn := "create"
	_, authRequired := a.isTokenProcessingRequired(r)
	if !authRequired {
		next(w, r)
		return
	}

	requestBody := a.getBody(r)
	var dcReq api.VolumeCreateRequest
	var dcRes api.VolumeCreateResponse
	if err := json.NewDecoder(requestBody).Decode(&dcReq); err != nil {
		next(w, r)
		return
	}

	spec := dcReq.GetSpec()
	locator := dcReq.GetLocator()
	tokenSecretContext, err := a.parseSecret(spec.VolumeLabels, locator.VolumeLabels, true)
	if err != nil {
		a.log(locator.Name, fn).WithError(err).Error("failed to parse secret")
		dcRes.VolumeResponse = &api.VolumeResponse{Error: "failed to parse secret: " + err.Error()}
		json.NewEncoder(w).Encode(&dcRes)
		return
	}
	if tokenSecretContext.SecretName == "" {
		errorMessage := "Access denied, no secret found in the annotations of the persistent volume claim" +
			" or storage class parameters"
		a.log(locator.Name, fn).Error(errorMessage)
		dcRes.VolumeResponse = &api.VolumeResponse{Error: errorMessage}
		json.NewEncoder(w).Encode(&dcRes)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := osecrets.GetToken(tokenSecretContext)
	if err != nil {
		a.log(locator.Name, fn).WithError(err).Error("failed to get token")
		dcRes.VolumeResponse = &api.VolumeResponse{Error: "failed to get token: " + err.Error()}
		json.NewEncoder(w).Encode(&dcRes)
		return

	} else {
		a.insertToken(r, token)
	}
	next(w, r)
}

func (a *authMiddleware) setWithAuth(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	fn := "set"
	d, authRequired := a.isTokenProcessingRequired(r)
	if !authRequired {
		next(w, r)
		return
	}

	volumeID, err := a.parseID(r)
	if err != nil {
		a.log(volumeID, fn).WithError(err).Error("Failed to parse volumeID")
		next(w, r)
		return
	}

	requestBody := a.getBody(r)
	var (
		req      api.VolumeSetRequest
		resp     api.VolumeSetResponse
		isOpDone bool
	)
	err = json.NewDecoder(requestBody).Decode(&req)
	if err != nil {
		a.log(volumeID, fn).WithError(err).Error("Failed to parse the request")
		next(w, r)
		return
	}

	// Not checking tokens for the following APIs
	// - Resize
	// - Attach/Detach
	// - Mount/Unmount

	if req.Spec != nil && req.Spec.Size > 0 {
		isOpDone = true
		err = d.Set(volumeID, req.Locator, req.Spec)
	}

	for err == nil && req.Action != nil {
		if req.Action.Attach != api.VolumeActionParam_VOLUME_ACTION_PARAM_NONE {
			isOpDone = true
			if req.Action.Attach == api.VolumeActionParam_VOLUME_ACTION_PARAM_ON {
				_, err = d.Attach(volumeID, req.Options)
			} else {
				err = d.Detach(volumeID, req.Options)
			}
			if err != nil {
				break
			}
		}

		if req.Action.Mount != api.VolumeActionParam_VOLUME_ACTION_PARAM_NONE {
			isOpDone = true
			if req.Action.Mount == api.VolumeActionParam_VOLUME_ACTION_PARAM_ON {
				if req.Action.MountPath == "" {
					err = fmt.Errorf("Invalid mount path")
					break
				}
				err = d.Mount(volumeID, req.Action.MountPath, req.Options)
			} else {
				err = d.Unmount(volumeID, req.Action.MountPath, req.Options)
			}
			if err != nil {
				break
			}
		}
		break
	}

	if isOpDone {
		if err != nil {
			processErrorForVolSetResponse(req.Action, err, &resp)
		} else {
			v, err := d.Inspect([]string{volumeID})
			if err != nil {
				processErrorForVolSetResponse(req.Action, err, &resp)
			} else if v == nil || len(v) != 1 {
				processErrorForVolSetResponse(
					req.Action,
					status.Errorf(codes.NotFound, "Volume with ID: %s is not found", volumeID),
					&resp)
			} else {
				v0 := v[0]
				resp.Volume = v0
			}
		}
		json.NewEncoder(w).Encode(resp)
		// Not calling the next handler
		return
	}
	next(w, r)
}

func (a *authMiddleware) deleteWithAuth(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	fn := "delete"
	d, authRequired := a.isTokenProcessingRequired(r)
	if !authRequired {
		next(w, r)
		return
	}

	volumeID, err := a.parseID(r)
	if err != nil {
		a.log(volumeID, fn).WithError(err).Error("Failed to parse volumeID")
		next(w, r)
		return
	}

	vols, err := d.Inspect([]string{volumeID})
	if err != nil || len(vols) == 0 || vols[0] == nil {
		a.log(volumeID, fn).WithError(err).Error("Failed to get volume object")
		next(w, r)
		return
	}

	volumeResponse := &api.VolumeResponse{}
	tokenSecretContext, err := a.parseSecret(vols[0].Spec.VolumeLabels, vols[0].Locator.VolumeLabels, false)
	if err != nil {
		a.log(volumeID, fn).WithError(err).Error("failed to parse secret")
		volumeResponse.Error = "failed to parse secret: " + err.Error()
		json.NewEncoder(w).Encode(volumeResponse)
		return
	}
	if tokenSecretContext.SecretName == "" {
		errorMessage := fmt.Sprintf("Error, unable to get secret information from the volume."+
			" You may need to re-add the following keys as volume labels to point to the secret: %s and %s",
			osecrets.SecretNameKey, osecrets.SecretNamespaceKey)
		a.log(volumeID, fn).Error(errorMessage)
		volumeResponse = &api.VolumeResponse{Error: errorMessage}
		json.NewEncoder(w).Encode(volumeResponse)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := osecrets.GetToken(tokenSecretContext)
	if err != nil {
		a.log(volumeID, fn).WithError(err).Error("failed to get token")
		volumeResponse.Error = "failed to get token: " + err.Error()
		json.NewEncoder(w).Encode(volumeResponse)
		return
	} else {
		a.insertToken(r, token)
	}

	next(w, r)
}

func (a *authMiddleware) inspectWithAuth(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	fn := "inspect"
	d, authRequired := a.isTokenProcessingRequired(r)
	if !authRequired {
		next(w, r)
		return
	}

	volumeID, err := a.parseID(r)
	if err != nil {
		a.log(volumeID, fn).WithError(err).Error("Failed to parse volumeID")
		next(w, r)
		return
	}

	dk, err := d.Inspect([]string{volumeID})
	if err != nil {
		a.log(volumeID, fn).WithError(err).Error("Failed to inspect volume")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(dk)
}

func (a *authMiddleware) enumerateWithAuth(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	fn := "enumerate"

	d, authRequired := a.isTokenProcessingRequired(r)
	if !authRequired {
		next(w, r)
		return
	}

	volIDs, ok := r.URL.Query()[api.OptVolumeID]
	if !ok || len(volIDs[0]) < 1 {
		a.log("", fn).Error("Failed to parse VolumeID")
		return
	}
	volumeID := volIDs[0]

	vols, err := d.Inspect([]string{volumeID})
	if err != nil || len(vols) == 0 || vols[0] == nil {
		a.log(volumeID, fn).WithError(err).Error("Failed to get volume object")
		next(w, r)
		return
	}

	volumeResponse := &api.VolumeResponse{}
	tokenSecretContext, err := a.parseSecret(vols[0].Spec.VolumeLabels, vols[0].Locator.VolumeLabels, false)
	if err != nil {
		a.log(volumeID, fn).WithError(err).Error("failed to parse secret")
		volumeResponse.Error = "failed to parse secret: " + err.Error()
		json.NewEncoder(w).Encode(volumeResponse)
		return
	}
	if tokenSecretContext.SecretName == "" {
		errorMessage := fmt.Sprintf("Error, unable to get secret information from the volume."+
			" You may need to re-add the following keys as volume labels to point to the secret: %s and %s",
			osecrets.SecretNameKey, osecrets.SecretNamespaceKey)
		a.log(volumeID, fn).Error(errorMessage)
		volumeResponse = &api.VolumeResponse{Error: errorMessage}
		json.NewEncoder(w).Encode(volumeResponse)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := osecrets.GetToken(tokenSecretContext)
	if err != nil {
		a.log(volumeID, fn).WithError(err).Error("failed to get token")
		volumeResponse.Error = "failed to get token: " + err.Error()
		json.NewEncoder(w).Encode(volumeResponse)
		return
	} else {
		a.insertToken(r, token)
	}

	next(w, r)
}

func (a *authMiddleware) isTokenProcessingRequired(r *http.Request) (volume.VolumeDriver, bool) {
	// If a token has been passed, then return here
	if len(r.Header.Get("Authorization")) > 0 {
		return nil, false
	}

	// No token has been passed in the request. Determine
	// if the request is from Kubernetes
	userAgent := r.Header.Get("User-Agent")
	if len(userAgent) > 0 {
		// Check if the request is coming from a container orchestrator
		clientName := strings.Split(userAgent, "/")
		if len(clientName) > 0 {
			if strings.HasSuffix(clientName[0], schedDriverPostFix) {
				d, err := volumedrivers.Get(clientName[0])
				if err != nil {
					return nil, false
				}
				return d, true
			}
		}
	}
	return nil, false
}

func (a *authMiddleware) insertToken(r *http.Request, token string) {
	// Set the token in header
	if auth.IsJwtToken(token) {
		r.Header.Set("Authorization", "bearer "+token)
	} else {
		r.Header.Set("Authorization", "Basic "+token)
	}
}

func (a *authMiddleware) parseID(r *http.Request) (string, error) {
	if id, err := a.parseParam(r, "id"); err == nil {
		return id, nil
	}

	return "", fmt.Errorf("could not parse snap ID")
}

func (a *authMiddleware) parseParam(r *http.Request, param string) (string, error) {
	vars := mux.Vars(r)
	if id, ok := vars[param]; ok {
		return id, nil
	}
	return "", fmt.Errorf("could not parse %s", param)
}

func (a *authMiddleware) parseSecret(
	specLabels, locatorLabels map[string]string,
	fetchCOLabels bool,
) (*api.TokenSecretContext, error) {
	if lsecrets.Instance() != nil &&
		lsecrets.Instance().String() == lsecrets.TypeK8s && fetchCOLabels {
		// For k8s fetch the actual annotations
		pvcName, ok := locatorLabels[PVCNameLabelKey]
		if !ok {
			// best effort to fetch the secret
			return parseSecretFromLabels(specLabels, locatorLabels)
		}
		pvcNamespace, ok := locatorLabels[PVCNamespaceLabelKey]
		if !ok {
			// best effort to fetch the secret
			return parseSecretFromLabels(specLabels, locatorLabels)
		}

		pvc, err := core.Instance().GetPersistentVolumeClaim(pvcName, pvcNamespace)
		if err != nil {
			return nil, err
		}
		secretName := pvc.ObjectMeta.Annotations[osecrets.SecretNameKey]

		if len(secretName) == 0 {
			return parseSecretFromLabels(specLabels, locatorLabels)
		}
		secretNamespace := pvc.ObjectMeta.Annotations[osecrets.SecretNamespaceKey]

		return &api.TokenSecretContext{
			SecretName:      secretName,
			SecretNamespace: secretNamespace,
		}, nil
	}
	return parseSecretFromLabels(specLabels, locatorLabels)
}

func parseSecretFromLabels(specLabels, locatorLabels map[string]string) (*api.TokenSecretContext, error) {
	// Locator labels take precendence
	secretName := locatorLabels[osecrets.SecretNameKey]
	secretNamespace := locatorLabels[osecrets.SecretNamespaceKey]
	if secretName == "" {
		secretName = specLabels[osecrets.SecretNameKey]
	}
	if secretName == "" {
		return nil, fmt.Errorf("secret name is empty")
	}
	if secretNamespace == "" {
		secretNamespace = specLabels[osecrets.SecretNamespaceKey]
	}

	return &api.TokenSecretContext{
		SecretName:      secretName,
		SecretNamespace: secretNamespace,
	}, nil
}

func (a *authMiddleware) log(id, fn string) *logrus.Entry {
	return logrus.WithFields(map[string]interface{}{
		"ID":        id,
		"Component": "auth-middleware",
		"Function":  fn,
	})
}

func (a *authMiddleware) getBody(r *http.Request) io.ReadCloser {
	// Make a copy of the reader so that the next handler
	// has access to the body
	buf, _ := ioutil.ReadAll(r.Body)
	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))

	r.Body = rdr2
	return rdr1
}
