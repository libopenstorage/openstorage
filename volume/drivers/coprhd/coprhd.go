package coprhd

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/portworx/kvdb"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
	"gopkg.in/jmcvetta/napping.v3"
)

const (
	Name = "coprhd"
	Type = api.Block

	// Place holders
	RESTBaseUrl = "https://localhost:4443/"
	APIUser     = "root"
	APIPassword = "ChangeMe"

	// Common API URIs
	LoginURI       = "login.json"
	CreateVolURI   = "block/volumes.json"
	DeleteVolURI   = "block/volumes/%s/deactivate.json"
	CreateSnapURI  = "block/volumes/%s/protections/snapshots.json"
	ExportVolURI   = "block/export.json"
	UnexportVolURI = "block/export/%s/deactivate.json"
)

type driver struct {
	*volume.DefaultEnumerator
	consistency_group string
	project           string
	varray            string
	vpool             string
	url               string
	user              string
	password          string
}

type ApiError struct {
	Code        int    `json:"code"`
	Retryable   bool   `json:"retryable"`
	Description string `json:"description"`
	Details     string `json:"details"`
}

func Init(params volume.DriverParams) (volume.VolumeDriver, error) {

	consistency_group, ok := params["consistency_group"]

	project, ok := params["project"]
	if !ok {
		project = "Default"
	}

	varray, ok := params["varray"]
	if !ok {
		varray = "Default"
	}

	vpool, ok := params["vpool"]
	if !ok {
		vpool = "Default"
	}

	url, ok := params["url"]
	if !ok {
		url = RESTBaseUrl
	}

	user, ok := params["user"]
	if !ok {
		user = APIUser
	}

	pass, ok := params["password"]
	if !ok {
		user = APIPassword
	}

	d := &driver{
		consistency_group: consistency_group,
		project:           project,
		varray:            varray,
		vpool:             vpool,
		url:               url,
		user:              user,
		password:          pass,
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

// API volume create args
type CreateVolumeArgs struct {
	Consistency_group string `json:"consistency_group,omitempty"`
	Count             int    `json:"count"`
	Name              string `json:"name"`
	Project           string `json:"project"`
	Size              string `json:"size"`
	Varray            string `json:"varray"`
	Vpool             string `json:"vpool"`
}

type CreateVolumeReply struct {
	Task []struct {
		Resource struct {
			Name string       `json:"name"`
			Id   api.VolumeID `json:"id"`
		} `json:"resource"`
	} `json:"task"`
}

func (d *driver) Create(
	locator api.VolumeLocator,
	source *api.Source,
	spec *api.VolumeSpec) (api.VolumeID, error) {

	e := ApiError{}

	token, err := d.getAuthToken()

	if err != nil {
		log.Printf("error: %s", err.Error())
		return api.BadVolumeID, err
	}

	log.Printf("API auth token: %s\n\n", token)

	p := []string{d.url, CreateVolURI}

	url := strings.Join(p, "")

	h := http.Header{}

	h.Add("X-SDS-AUTH-TOKEN", token)

	s := napping.Session{
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
		Header: &h,
	}

	res := CreateVolumeReply{}

	sz := float64(spec.Size / (1024 * 1024 * 1000))

	payload := CreateVolumeArgs{
		Consistency_group: d.consistency_group,
		Count:             1,
		Name:              locator.Name,
		Project:           d.project,
		Size:              fmt.Sprintf("%.6fGB", sz),
		Varray:            d.varray,
		Vpool:             d.vpool,
	}

	resp, err := s.Post(url, &payload, &res, &e)

	if resp.Status() == 202 {

		volId := res.Task[0].Resource.Id

		log.Printf("Coprhd volume %s created\n", volId)

		return volId, err
	} else {
		log.Printf("error: %s (%d)", e.Details, e.Code)

		err = fmt.Errorf("%s", e.Details)
	}

	return api.BadVolumeID, err
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
	log.Printf("%s Shutting down", Name)
}

func (d *driver) Snapshot(volumeID api.VolumeID, readonly bool, locator api.VolumeLocator) (api.VolumeID, error) {
	return "", nil
}

func (v *driver) Status() [][2]string {
	return [][2]string{}
}

// Retrieves an API Session Auth Token
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
