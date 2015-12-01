package coprhd

import (
	"fmt"
	"strings"
	"net/url"
	"log"
	"net/http"
	"crypto/tls"

	"github.com/portworx/kvdb"
	
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/volume"
	"gopkg.in/jmcvetta/napping.v3"
)

const (
	Name           = "coprhd"
	Type           = api.Block
	RESTBaseUrl    = "https://localhost:4443/"
	APIUser	       = "root"
	APIPassword    = "ChangeMe"
	LoginPath      = "login.json"
	VolumePath     = "block/volumes.json"
)

type driver struct {
	*volume.DefaultEnumerator
	consistency_group string
	project string
	varray string
	vpool string
	url string
	user string
	password string
}

type ApiError struct {
	code string
	retryable string
	description string
	details string
}

type CreateVolume struct{
	consistency_group string `json:"consistency_group"`
	count int `json:"count"`
	name string `json:"name"`
	project string `json:"project"`
	size string `json:"size"`
	varray string `json:"varray"`
	vpool string `json:"vpool"`
}

func Init(params volume.DriverParams) (volume.VolumeDriver, error) {

	consistency_group, ok := params["consistency_group"]
	if !ok {
		consistency_group = "Default"
	}

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
		project: project,
		varray: varray,
		vpool: vpool,
		url: url,
		user: user,
		password: pass,
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
		log.Fatal(err)
		return api.BadVolumeID, err
	}
	
	fmt.Printf("API auth token: %s\n\n", token)

	p := []string{d.url, VolumePath}

	url := strings.Join(p, "")

	fmt.Println(url)

	h := http.Header{}

	h.Add("X-SDS-AUTH-TOKEN", token)
	
	s := napping.Session{
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify : true},
			},
		},
		Header: &h,
	}
	
	payload := CreateVolume{
		consistency_group: "Default",
		count: 1,
		name: locator.Name,
		project: "Default",
		size: "1GB",
		varray: "Default",
		vpool: "Default",
	}

	resp, err := s.Post(url, &payload, nil, &e)

	if resp.Status() == 200 {
		return api.BadVolumeID, err
	} else {
		fmt.Println("Bad response status from API server")
		fmt.Printf("\t Status:  %v\n", resp.Status())
	}
	
	println("")
	
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

func (d *driver) getAuthToken () (token string, err error) {

	p := []string{d.url, LoginPath}

	e := ApiError{}
	
	s := napping.Session{
		Userinfo: url.UserPassword(d.user, d.password),
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify : true},
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

