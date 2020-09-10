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
	osecrets "github.com/libopenstorage/openstorage/pkg/auth/secrets"
	"github.com/libopenstorage/openstorage/pkg/util"
	"github.com/libopenstorage/openstorage/volume"
	volumedrivers "github.com/libopenstorage/openstorage/volume/drivers"
	lsecrets "github.com/libopenstorage/secrets"
	"github.com/portworx/sched-ops/k8s/core"
	"github.com/sirupsen/logrus"
)

const (
	// PVCNameLabelKey is used for kubernetes auth provider indicating the name of PVC
	PVCNameLabelKey = "pvc"
	// PVCNamespaceLabelKey is used for kubernetes auth provider indicating the namespace of the PVC
	PVCNamespaceLabelKey = "namespace"
)

var (
	// OverrideSchedDriverName is set by osd program to override the schedule driver
	OverrideSchedDriverName = ""
)

// NewK8sMiddleware returns a negroni implementation of an http middleware
// which will intercept the management APIs
func NewK8sMiddleware() *k8sMiddleware {
	return &k8sMiddleware{}
}

type k8sMiddleware struct {
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

func (a *k8sMiddleware) createWithAuth(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
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
	tokenSecretContext, err := a.parseSecret(spec.VolumeLabels, locator.VolumeLabels)
	if err != nil {
		a.log(locator.Name, fn).WithError(err).Error("failed to parse secret")
		dcRes.VolumeResponse = &api.VolumeResponse{Error: "failed to parse secret: " + err.Error()}
		json.NewEncoder(w).Encode(&dcRes)
		return
	} else if tokenSecretContext == nil {
		tokenSecretContext = &api.TokenSecretContext{}
	}

	// If no secret is provided, then the caller is accessing publicly
	if tokenSecretContext.SecretName != "" {
		token, err := osecrets.GetToken(tokenSecretContext)
		if err != nil {
			a.log(locator.Name, fn).WithError(err).Error("failed to get token")
			dcRes.VolumeResponse = &api.VolumeResponse{Error: "failed to get token: " + err.Error()}
			json.NewEncoder(w).Encode(&dcRes)
			return
		}

		// Save a reference to the secret
		// These values will be stored in the header for the create() server handler
		// to take and place in the labels for the volume since we do not want to adjust
		// the body of the request in this middleware. When create() gets these values
		// from the headers, it will copy them to the labels of the volume so that
		// we can track the secret in the rest of the middleware calls.
		r.Header.Set(secrets.SecretNameKey, tokenSecretContext.SecretName)
		r.Header.Set(secrets.SecretNamespaceKey, tokenSecretContext.SecretNamespace)

		// If the source PVC was set, save it for the next layer to store on
		// the labels of the volume
		if len(tokenSecretContext.PvcName) != 0 && len(tokenSecretContext.PvcNamespace) != 0 {
			r.Header.Set(api.KubernetesPvcNameKey, tokenSecretContext.PvcName)
			r.Header.Set(api.KubernetesPvcNamespaceKey, tokenSecretContext.PvcNamespace)
		}

		a.insertToken(r, token)
	}
	next(w, r)
}

func (a *k8sMiddleware) setWithAuth(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
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

	var req api.VolumeSetRequest

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		volumeResponse := &api.VolumeResponse{
			Error: err.Error(),
		}
		json.NewEncoder(w).Encode(volumeResponse)
		return
	}
	if req.Spec != nil {
		// A volume Set request for changing spec
		// For requests from COs we need to reset certain fields
		vols, err := d.Inspect([]string{volumeID})
		if err != nil || len(vols) == 0 || vols[0] == nil {
			next(w, r)
			return
		}
		resetFields(req.Spec, vols[0].Spec)
	}

	newBody, err := json.Marshal(req)
	if err != nil {
		err = fmt.Errorf("failed to wrap set request: %v", err)
		volumeResponse := &api.VolumeResponse{
			Error: err.Error(),
		}
		json.NewEncoder(w).Encode(volumeResponse)
		return
	}

	// Reset the request's Body
	r.ContentLength = int64(len(newBody))
	r.Body = ioutil.NopCloser(bytes.NewBuffer(newBody))

	token, err := a.fetchSecretForVolume(d, volumeID)
	if err != nil {
		volumeResponse := &api.VolumeResponse{}
		a.log(volumeID, fn).WithError(err).Error("Failed to fetch secret")
		volumeResponse.Error = err.Error()
		json.NewEncoder(w).Encode(volumeResponse)
		return
	}
	if len(token) != 0 {
		a.insertToken(r, token)
	}

	next(w, r)
}

func resetFields(inputSpec *api.VolumeSpec, currentSpec *api.VolumeSpec) {
	// If the request is coming from a container orchestrator like kubernetes
	// we only allow setting the volume size in the Set request. All other fields
	// should be reset to the original values. This needs to be done since kubernetes
	// in-tree driver has older vendor'ed openstorage code, so any new boolean flags
	// added to the VolumeSpec will always be set to false overwriting the exising value.

	inputSpec.Cascaded = currentSpec.Cascaded
	inputSpec.Journal = currentSpec.Journal
	inputSpec.Sharedv4 = currentSpec.Sharedv4
	inputSpec.QueueDepth = currentSpec.QueueDepth
	inputSpec.ForceUnsupportedFsType = currentSpec.ForceUnsupportedFsType
	inputSpec.Nodiscard = currentSpec.Nodiscard
	inputSpec.StoragePolicy = currentSpec.StoragePolicy
}

func (a *k8sMiddleware) deleteWithAuth(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
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

	// Idempotency
	vols, err := d.Inspect([]string{volumeID})
	if err != nil || len(vols) == 0 || vols[0] == nil {
		next(w, r)
		return
	}

	token, err := a.fetchSecretForVolume(d, volumeID)
	if err != nil {
		volumeResponse := &api.VolumeResponse{}
		a.log(volumeID, fn).WithError(err).Error("Failed to fetch secret")
		volumeResponse.Error = err.Error()
		json.NewEncoder(w).Encode(volumeResponse)
		return
	}
	if len(token) != 0 {
		a.insertToken(r, token)
	}

	next(w, r)
}

func (a *k8sMiddleware) inspectWithAuth(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
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

func (a *k8sMiddleware) enumerateWithAuth(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
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

	token, err := a.fetchSecretForVolume(d, volumeID)
	if err != nil {
		volumeResponse := &api.VolumeResponse{}
		a.log(volumeID, fn).WithError(err).Error("Failed to fetch secret")
		volumeResponse.Error = err.Error()
		json.NewEncoder(w).Encode(volumeResponse)
		return
	}
	if len(token) != 0 {
		a.insertToken(r, token)
	}

	next(w, r)
}

func (a *k8sMiddleware) isTokenProcessingRequired(r *http.Request) (volume.VolumeDriver, bool) {
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
				driverName := clientName[0]
				if len(OverrideSchedDriverName) != 0 {
					driverName = OverrideSchedDriverName
				}
				d, err := volumedrivers.Get(driverName)
				if err != nil {
					return nil, false
				}
				return d, true
			}
		}
	}
	return nil, false
}

func (a *k8sMiddleware) insertToken(r *http.Request, token string) {
	// Set the token in header
	if auth.IsJwtToken(token) {
		r.Header.Set("Authorization", "bearer "+token)
	} else {
		r.Header.Set("Authorization", "Basic "+token)
	}
}

func (a *k8sMiddleware) parseID(r *http.Request) (string, error) {
	if id, err := a.parseParam(r, "id"); err == nil {
		return id, nil
	}

	return "", fmt.Errorf("could not parse snap ID")
}

func (a *k8sMiddleware) parseParam(r *http.Request, param string) (string, error) {
	vars := mux.Vars(r)
	if id, ok := vars[param]; ok {
		return id, nil
	}
	return "", fmt.Errorf("could not parse %s", param)
}

// This functions makes it possible to secure the model of accessing the secret by allowing
// the definition of secret access to come from the storage class, as done by CSI.
func (a *k8sMiddleware) getSecretInformationInKubernetes(
	specLabels, locatorLabels map[string]string,
) (*api.TokenSecretContext, error) {
	// Get pvc location and name
	// For k8s fetch the actual annotations
	pvcName, ok := getVolumeLabel(PVCNameLabelKey, specLabels, locatorLabels)
	if !ok {
		return nil, fmt.Errorf("Unable to authenticate request due to not able to determine name of the pvc from the volume")
	}
	pvcNamespace, ok := getVolumeLabel(PVCNamespaceLabelKey, specLabels, locatorLabels)
	if !ok {
		return nil, fmt.Errorf("Unable to authenticate request due to not able to determine namespace of the pvc from the volume")
	}

	// Get pvc object
	pvc, err := core.Instance().GetPersistentVolumeClaim(pvcName, pvcNamespace)
	if err != nil {
		return nil, fmt.Errorf("Unable to get PVC information from Kubernetes: %v", err)
	}

	// Get storageclass for pvc object
	sc, err := core.Instance().GetStorageClassForPVC(pvc)
	if err != nil {
		return nil, fmt.Errorf("Unable to get StorageClass information from Kubernetes: %v", err)
	}

	// Get secret namespace
	secretNamespaceValue := sc.Parameters[osecrets.SecretNamespaceKey]
	secretNameValue := sc.Parameters[osecrets.SecretNameKey]
	if len(secretNameValue) == 0 && len(secretNamespaceValue) == 0 {
		return &api.TokenSecretContext{}, nil
	}

	// Allow ${pvc.namespace} to be set in the storage class
	namespaceParams := map[string]string{"pvc.namespace": pvc.GetNamespace()}
	secretNamespace, err := util.ResolveTemplate(secretNamespaceValue, namespaceParams)
	if err != nil {
		return nil, err
	}

	// Get secret name
	nameParams := make(map[string]string)
	// Allow ${pvc.annotations['pvcNameKey']} to be set in the storage class
	// See pkg/auth/secrets/secrets.go for more information
	for k, v := range pvc.Annotations {
		nameParams["pvc.annotations['"+k+"']"] = v
	}
	secretName, err := util.ResolveTemplate(secretNameValue, nameParams)
	if err != nil {
		return nil, err
	}

	return &api.TokenSecretContext{
		SecretName:      secretName,
		SecretNamespace: secretNamespace,
		PvcName:         pvcName,
		PvcNamespace:    pvcNamespace,
	}, nil
}

func (a *k8sMiddleware) parseSecret(
	specLabels, locatorLabels map[string]string,
) (*api.TokenSecretContext, error) {

	// Check if it is Kubernetes
	if lsecrets.Instance() != nil &&
		lsecrets.Instance().String() == lsecrets.TypeK8s {
		return a.getSecretInformationInKubernetes(specLabels, locatorLabels)
	}

	// Not Kubernetes, try to get secret information from labels
	return parseSecretFromLabels(specLabels, locatorLabels)
}

func parseSecretFromLabels(specLabels, locatorLabels map[string]string) (*api.TokenSecretContext, error) {
	// Locator labels take precedence
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

func (a *k8sMiddleware) log(id, fn string) *logrus.Entry {
	return logrus.WithFields(map[string]interface{}{
		"ID":        id,
		"Component": "auth-middleware",
		"Function":  fn,
	})
}

func (a *k8sMiddleware) getBody(r *http.Request) io.ReadCloser {
	// Make a copy of the reader so that the next handler
	// has access to the body
	buf, _ := ioutil.ReadAll(r.Body)
	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))

	r.Body = rdr2
	return rdr1
}

func getVolumeLabel(key string, specLabels, locatorLabels map[string]string) (string, bool) {
	if v, ok := locatorLabels[key]; ok {
		return v, true
	}
	v, ok := specLabels[key]
	return v, ok
}

func (a *k8sMiddleware) fetchSecretForVolume(d volume.VolumeDriver, id string) (string, error) {
	vols, err := d.Inspect([]string{id})
	if err != nil || len(vols) == 0 || vols[0] == nil {
		return "", fmt.Errorf("Volume %s does not exist", id)
	}

	v := vols[0]
	if v.GetLocator().GetVolumeLabels() == nil {
		return "", nil
	}

	tokenSecretContext := &api.TokenSecretContext{
		SecretName:      v.GetLocator().GetVolumeLabels()[secrets.SecretNameKey],
		SecretNamespace: v.GetLocator().GetVolumeLabels()[secrets.SecretNamespaceKey],
	}

	// If no secret is provided, then the caller is accessing publicly
	if tokenSecretContext.SecretName == "" || tokenSecretContext.SecretNamespace == "" {
		return "", nil
	}

	// Retrieve secret
	token, err := osecrets.GetToken(tokenSecretContext)
	if err != nil {
		return "", fmt.Errorf("Failed to get token from secret %s/%s: %v",
			tokenSecretContext.SecretNamespace,
			tokenSecretContext.SecretName,
			err)
	}
	return token, nil
}
