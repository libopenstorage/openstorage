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
	"github.com/libopenstorage/openstorage/pkg/auth/secrets"
	"github.com/libopenstorage/openstorage/volume"
	volumedrivers "github.com/libopenstorage/openstorage/volume/drivers"
	osecrets "github.com/libopenstorage/secrets"
	"github.com/portworx/sched-ops/k8s"
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
func NewAuthMiddleware(
	s osecrets.Secrets,
	authType secrets.AuthTokenProviders,
) (*authMiddleware, error) {
	provider, err := secrets.NewAuth(
		authType,
		s,
	)
	if err != nil {
		return nil, err
	}
	return &authMiddleware{provider}, nil
}

type authMiddleware struct {
	provider secrets.Auth
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
	secretName, secretContext, err := a.parseSecret(spec.VolumeLabels, locator.VolumeLabels, true)
	if err != nil {
		a.log(locator.Name, fn).WithError(err).Error("failed to parse secret")
		dcRes.VolumeResponse = &api.VolumeResponse{Error: "failed to parse secret: " + err.Error()}
		json.NewEncoder(w).Encode(&dcRes)
		return
	}
	if secretName == "" {
		errorMessage := "Access denied, no secret found in the annotations of the persistent volume claim" +
			" or storage class parameters"
		a.log(locator.Name, fn).Error(errorMessage)
		dcRes.VolumeResponse = &api.VolumeResponse{Error: errorMessage}
		json.NewEncoder(w).Encode(&dcRes)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := a.provider.GetToken(secretName, secretContext)
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
	secretName, secretContext, err := a.parseSecret(vols[0].Spec.VolumeLabels, vols[0].Locator.VolumeLabels, false)
	if err != nil {
		a.log(volumeID, fn).WithError(err).Error("failed to parse secret")
		volumeResponse.Error = "failed to parse secret: " + err.Error()
		json.NewEncoder(w).Encode(volumeResponse)
		return
	}
	if secretName == "" {
		errorMessage := fmt.Sprintf("Error, unable to get secret information from the volume."+
			" You may need to re-add the following keys as volume labels to point to the secret: %s and %s",
			secrets.SecretNameKey, secrets.SecretNamespaceKey)
		a.log(volumeID, fn).Error(errorMessage)
		volumeResponse = &api.VolumeResponse{Error: errorMessage}
		json.NewEncoder(w).Encode(volumeResponse)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := a.provider.GetToken(secretName, secretContext)
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

func (a *authMiddleware) isTokenProcessingRequired(r *http.Request) (volume.VolumeDriver, bool) {
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
) (string, string, error) {
	if a.provider.Type() == secrets.TypeK8s && fetchCOLabels {
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

		pvc, err := k8s.Instance().GetPersistentVolumeClaim(pvcName, pvcNamespace)
		if err != nil {
			return "", "", err
		}
		secretName := pvc.ObjectMeta.Annotations[secrets.SecretNameKey]
		secretNamespace := pvc.ObjectMeta.Annotations[secrets.SecretNamespaceKey]
		if len(secretName) == 0 {
			return parseSecretFromLabels(specLabels, locatorLabels)
		}

		return secretName, secretNamespace, nil
	}
	return parseSecretFromLabels(specLabels, locatorLabels)
}

func parseSecretFromLabels(specLabels, locatorLabels map[string]string) (string, string, error) {
	// Locator labels take precendence
	secretName := locatorLabels[secrets.SecretNameKey]
	secretNamespace := locatorLabels[secrets.SecretNamespaceKey]
	if secretName == "" {
		secretName = specLabels[secrets.SecretNameKey]
	}
	if secretName == "" {
		return "", "", fmt.Errorf("secret name is empty")
	}
	if secretNamespace == "" {
		secretNamespace = specLabels[secrets.SecretNamespaceKey]
	}
	return secretName, secretNamespace, nil
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
