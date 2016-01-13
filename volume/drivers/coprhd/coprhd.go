package coprhd

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/portworx/kvdb"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"

	napping "gopkg.in/jmcvetta/napping.v3"
)

const (
	Name = "coprhd"
	Type = api.Block

	// URI_LOGIN path to create a authentication token
	URI_LOGIN = "login.json"
	// URI_LOGIN path to create volume
	URI_CREATE_VOL = "block/volumes.json"
	// URI_EXPORT_VOL path to export a volume
	URI_EXPORT_VOL = "block/export.json"
	// URI_TPL_DEL_VOL template path to delete/deactivate a volume
	URI_TPL_DEL_VOL = "block/volumes/%s/deactivate.json"
	// URL_TPL_NEW_SNAP path to create a volume snapshot
	URL_TPL_NEW_SNAP = "block/volumes/%s/protections/snapshots.json"
	// URI_TPL_UNEXP_VOL path template to remove a volume export
	URI_TPL_UNEXP_VOL = "block/export/%s/deactivate.json"
)

type (
	driver struct {
		*volume.DefaultEnumerator
		consistency_group string
		project           string
		varray            string
		vpool             string
		url               string
		creds             *url.UserInfo
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

func Init(params volume.DriverParams) (volume.VolumeDriver, error) {

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

	url, ok := params["url"]
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

	d := &driver{
		consistency_group: consistency_group,
		project:           project,
		varray:            varray,
		vpool:             vpool,
		url:               url,
		creads:            url.UserPassword(user, password),
		DefaultEnumerator: volume.NewDefaultEnumerator(Name, kvdb.Instance()),
	}

	return d, nil
}

func (d *driver) String() string {
	return Name
}

func (d *driver) Type() api.DriverType {
	return Type
}

func init() {
	volume.Register(Name, Init)
}

func (d *driver) Create(
	locator api.VolumeLocator,
	source *api.Source,
	spec *api.VolumeSpec) (api.VolumeID, error) {

	e := ApiError{}

	token, err := d.getAuthToken()

	if err != nil {
		log.Printf("%s", err.Error())
		return api.BadVolumeID, err
	}

	url := d.url + URI_CREATE_VOL

	h := http.Header{"X-SDS-AUTH-TOKEN": token}

	s := napping.Session{
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
		Header: &h,
	}

	res := &CreateVolumeReply{}

	payload := CreateVolumeArgs{
		"Default",                 // ConsistencyGroup
		1,                         // Count
		locator.Name,              // Name
		d.Project,                 // Project
		fmt.Sprintf("%.6fGB", sz), // Volume Size
		d.varray,                  // Virtual Block Array
		d.vpool,                   // Virtual Block Pool
	}

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
	log.Infof("%s Shutting down", Name)
}

func (d *driver) Snapshot(volumeID api.VolumeID, readonly bool, locator api.VolumeLocator) (api.VolumeID, error) {
	return "", nil
}

func (v *driver) Status() [][2]string {
	return [][2]string{}
}

// getAuthToken returns an API Session Token
func (d *driver) getAuthToken() (token string, err error) {

	p := []string{d.url, LoginURI}

	e := ApiError{}

	s := napping.Session{
		Userinfo: url.UserPassword(d.user, d.password),
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}

	url := strings.Join(p, "")

	resp, err := s.Get(url, nil, nil, &e)

	if err != nil {
		return "", err
	}

	token = resp.HttpResponse().Header.Get("X-SDS-AUTH-TOKEN")

	return token, nil
}
