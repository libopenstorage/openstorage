package coprhd

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"


	"github.com/Sirupsen/logrus"
	"github.com/portworx/kvdb"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"

	"gopkg.in/jmcvetta/napping.v3"
)

const (
	Name = "coprhd"
	Type = api.Block

	// LoginUri path to create a authentication token
	loginUri = "login.json"
	// LoginUri path to create volume
	createVolumeUri = "block/volumes.json"
)

type (
	driver struct {
		*volume.IoNotSupported
		*volume.DefaultEnumerator
		consistency_group string
		project           string
		varray            string
		vpool             string
		url               string
		httpClient        *http.Client
		creds             *url.Userinfo
	}

	// ApiError represents the default api error code
	ApiError struct {
		Code        string `json:"code"`
		Retryable   string `json:"retryable"`
		Description string `json:"description"`
		Details     string `json:"details"`
	}

	// CreateVolumeArgs represents the json parameters for the create volume REST call
	CreateVolumeArgs struct {
		ConsistencyGroup string `json:"consistency_group"`
		Count            int    `json:"count"`
		Name             string `json:"name"`
		Project          string `json:"project"`
		Size             string `json:"size"`
		VArray           string `json:"varray"`
		VPool            string `json:"vpool"`
	}

	// CreateVolumeReply is the reply from the create volume REST call
	CreateVolumeReply struct {
		Task []struct {
			Resource struct {
				Name string       `json:"name"`
				Id   api.VolumeID `json:"id"`
			} `json:"resource"`
		} `json:"task"`
	}
)

func init() {
	volume.Register(Name, Init)
}

func Init(params volume.DriverParams) (volume.VolumeDriver, error) {
	restUrl, ok := params["restUrl"]
	if !ok {
		return nil, fmt.Errorf("rest api 'url' configuration parameter must be set")
	}

	user, ok := params["user"]
	if !ok {
		return nil, fmt.Errorf("rest auth 'user' must be set")
	}

	pass, ok := params["password"]
	if !ok {
		return nil, fmt.Errorf("rest auth 'password' must be set")
	}

	consistency_group, ok := params["consistency_group"]
	if !ok {
		return nil, fmt.Errorf("'consistency_group' configuration parameter must be set")
	}

	project, ok := params["project"]
	if !ok {
		return nil, fmt.Errorf("'project' configuration parameter must be set")
	}

	varray, ok := params["varray"]
	if !ok {
		return nil, fmt.Errorf("'varray' configuration parameter must be set")
	}

	vpool, ok := params["vpool"]
	if !ok {
		return nil, fmt.Errorf("'vpool' configuration parameter must be set")
	}

	d := &driver{
		DefaultEnumerator: volume.NewDefaultEnumerator(Name, kvdb.Instance()),
		consistency_group: consistency_group,
		project:           project,
		varray:            varray,
		vpool:             vpool,
		url:               restUrl,
		creds:             url.UserPassword(user, pass),
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}

	return d, nil
}

func (d *driver) String() string {
	return Name
}

func (d *driver) Type() api.DriverType {
	return Type
}

func (d *driver) Create(
	locator api.VolumeLocator,
	source *api.Source,
	spec *api.VolumeSpec) (api.VolumeID, error) {

	s, err := d.getAuthSession()

	if err != nil {
		logrus.Errorf("Failed to create session: %s", err.Error())
		return api.BadVolumeID, err
	}

	e := ApiError{}

	res := &CreateVolumeReply{}

	sz := int64(spec.Size / (1024 * 1024 * 1000))

	payload := CreateVolumeArgs{
		d.consistency_group,       // ConsistencyGroup
		1,                         // Count
		locator.Name,              // Name
		d.project,                 // Project
		fmt.Sprintf("%.6fGB", sz), // Volume Size
		d.varray,                  // Virtual Block Array
		d.vpool,                   // Virtual Block Pool
	}

	url := d.url + createVolumeUri

	resp, err := s.Post(url, &payload, res, &e)

	if resp.Status() != http.StatusAccepted {

		return api.BadVolumeID, fmt.Errorf("Failed to create volume: %s", resp.Status())
	}

	return res.Task[0].Resource.Id, err
}

func (d *driver) Delete(volumeID api.VolumeID) error {
	return nil
}

func (d *driver) Stats(volumeID api.VolumeID) (api.Stats, error) {
	return api.Stats{}, volume.ErrNotSupported
}

func (d *driver) Alerts(volumeID api.VolumeID) (api.Alerts, error) {
	return api.Alerts{}, volume.ErrNotSupported
}

func (d *driver) Attach(volumeID api.VolumeID) (path string, err error) {
	return "", nil
}

func (d *driver) Detach(volumeID api.VolumeID) error {
	return nil
}

func (d *driver) Mount(volumeID api.VolumeID, mountpath string) error {
	return nil
}

func (d *driver) Unmount(volumeID api.VolumeID, mountpath string) error {

	return nil
}

func (d *driver) Set(volumeID api.VolumeID, locator *api.VolumeLocator, spec *api.VolumeSpec) error {
	return volume.ErrNotSupported
}

func (d *driver) Shutdown() {
	logrus.Infof("%s Shutting down", Name)
}

func (d *driver) Snapshot(volumeID api.VolumeID, readonly bool, locator api.VolumeLocator) (api.VolumeID, error) {
	return "", nil
}

func (v *driver) Status() [][2]string {
	return [][2]string{}
}

// getAuthSession returns an authenticated API Session
func (d *driver) getAuthSession() (session *napping.Session, err error) {
	e := ApiError{}

	s := napping.Session{
		Userinfo: d.creds,
		Client:   d.httpClient,
	}

	url := d.url + loginUri

	resp, err := s.Get(url, nil, nil, &e)

	if err != nil {
		return
	}

	token := resp.HttpResponse().Header.Get("X-SDS-AUTH-TOKEN")

	h := http.Header{}

	h.Set("X-SDS-AUTH-TOKEN", token)

	session = &napping.Session{
		Client: d.httpClient,
		Header: &h,
	}

	return
}
