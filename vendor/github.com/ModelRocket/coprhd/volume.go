package coprhd

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	createVolumeUri    = "block/volumes.json"
	queryVolumeUriTpl  = "block/volumes/%s.json"
	searchVolumeUri    = "block/volumes/search.json?"
	listVolumesUri     = "block/volumes/bulk.json"
	deleteVolumeUriTpl = "block/volumes/%s/deactivate.json"
)

var (
	// ErrCreateResponse is returned when the api call returns an unexptected result
	ErrCreateResponse = errors.New("Invalid create response received")
)

type (
	// VolumeService is used to create, search, and query for volumes
	VolumeService struct {
		*Client
		id      string
		name    string
		wwn     string
		array   string
		pool    string
		group   string
		project string
	}

	// Volume is a complete coprhd volume object
	Volume struct {
		StorageObject       `json:",inline"`
		WWN                 string      `json:"wwn"`
		Protocols           []string    `json:"protocols"`
		Protection          interface{} `json:"protection"`
		ConsistencyGroup    string      `json:"consistency_group,omitempty"`
		StorageController   string      `json:"storage_controller"`
		DeviceLabel         string      `json:"device_label"`
		NativeId            string      `json:"native_id"`
		ProvisionedCapacity string      `json:"provisioned_capacity_gb"`
		AllocatedCapacity   string      `json:"allocated_capacity_gb"`
		RequestedCapacity   string      `json:"requested_capacity_gb"`
		PreAllocationSize   string      `json:"pre_allocation_size_gb"`
		IsComposite         bool        `json:"is_composite"`
		ThinlyProvisioned   bool        `json:"thinly_provisioned"`
		HABackingVolumes    []string    `json:"high_availability_backing_volumes"`
		AccessState         string      `json:"access_state"`
		StoragePool         Resource    `json:"storage_pool"`
	}

	// createVolumeReq represents the json parameters for the create volume REST call
	createVolumeReq struct {
		ConsistencyGroup string `json:"consistency_group,omitempty"`
		Count            int    `json:"count"`
		Name             string `json:"name"`
		Project          string `json:"project"`
		Size             string `json:"size"`
		VArray           string `json:"varray"`
		VPool            string `json:"vpool"`
	}

	// createVolumeRes is the reply from the create volume REST call
	createVolumeRes struct {
		Task []Task `json:"task"`
	}

	// listVolumesRes is the reply to geting a list of volumes
	listVolumesRes struct {
		Volumes []string `json:"id"`
	}
)

// Volume returns an instance of the VolumeService
func (this *Client) Volume() *VolumeService {
	return &VolumeService{
		Client: this.Copy(),
	}
}

// Id sets the volume id urn for the VolumeService instance
func (this *VolumeService) Id(id string) *VolumeService {
	// make sure Volume is capitalized
	this.id = strings.Replace(id, "volume", "Volume", 1)
	return this
}

// Name sets the volume name for the VolumeService instance
func (this *VolumeService) Name(name string) *VolumeService {
	this.name = name
	return this
}

func (this *VolumeService) WWN(wwn string) *VolumeService {
	this.wwn = wwn
	return this
}

// Array sets the varray urn for the VolumeService instance
func (this *VolumeService) Array(array string) *VolumeService {
	this.array = array
	return this
}

// Pool sets the vpool urn for the VolumeService instance
func (this *VolumeService) Pool(pool string) *VolumeService {
	this.pool = pool
	return this
}

// Group sets the consistency group urn for the VolumeService instance
func (this *VolumeService) Group(group string) *VolumeService {
	this.group = group
	return this
}

// Project sets the project urn for the VolumeService instance
func (this *VolumeService) Project(project string) *VolumeService {
	this.project = project
	return this
}

// Create creates a new volume with the specified name and size using the volume service
func (this *VolumeService) Create(size uint64) (*Volume, error) {
	sz := float64(size / (1024 * 1024 * 1000))

	if err := this.getVolumeUrns(); err != nil {
		return nil, err
	}

	req := createVolumeReq{
		Count:   1,
		Name:    this.name,
		Project: this.project,
		VArray:  this.array,
		VPool:   this.pool,
		Size:    fmt.Sprintf("%.6fGB", sz),
	}

	if this.group != "" {
		req.ConsistencyGroup = this.group
	}

	res := createVolumeRes{}

	err := this.post(createVolumeUri, &req, &res)
	if err != nil {
		if this.LastError().IsDup() {
			return this.Query()
		}
		return nil, err
	}

	if len(res.Task) != 1 {
		return nil, ErrCreateResponse
	}

	task := res.Task[0]

	// wait for the task to complete
	err = this.Task().WaitDone(task.Id, TaskStateReady, time.Second*180)
	if err != nil {
		return nil, err
	}

	this.id = task.Resource.Id

	return this.Query()
}

// Query returns the volume object using the specified id
func (this *VolumeService) Query() (*Volume, error) {
	if !isStorageOsUrn(this.id) {
		if this.name != "" {
			return this.Search("name=" + this.name)
		}
		return this.Search("wwn=" + this.wwn)
	}

	path := fmt.Sprintf(queryVolumeUriTpl, this.id)
	vol := Volume{}

	err := this.get(path, nil, &vol)
	if err != nil {
		return nil, err
	}

	return &vol, nil
}

// Search searches for a volume using the specified query string
// For example:
//    Search("name=foo")
//
func (this *VolumeService) Search(query string) (*Volume, error) {
	path := searchVolumeUri + query

	res, err := this.Client.Search(path)
	if err != nil {
		return nil, err
	}

	this.id = res[0].Id

	return this.Query()
}

func (this *VolumeService) List() ([]string, error) {

	res := listVolumesRes{}

	err := this.get(listVolumesUri, nil, &res)
	if err != nil {
		return nil, err
	}
	return res.Volumes, nil
}

// Delete deactivates the volume using the volume service
func (this *VolumeService) Delete(force bool) error {
	if this.id == "" {
		vol, err := this.Query()
		if err != nil {
			return err
		}
		this.id = vol.Id
	}
	path := fmt.Sprintf(deleteVolumeUriTpl, this.id)

	if force {
		path = path + "?force=true"
	}

	task := Task{}

	err := this.post(path, nil, &task)
	if err != nil {
		return err
	}

	return this.Task().WaitDone(task.Id, TaskStateReady, time.Second*180)
}

func (this *VolumeService) getVolumeUrns() error {

	// lookup the project by name
	if !isStorageOsUrn(this.project) {
		project, err := this.Client.Project().
			Name(this.project).
			Query()
		if err != nil {
			return err
		}
		this.project = project.Id
	}

	// lookup the array by name
	if !isStorageOsUrn(this.array) {
		array, err := this.Client.VArray().
			Name(this.array).
			Query()
		if err != nil {
			return err
		}
		this.array = array.Id
	}

	// lookup the pool by name
	if !isStorageOsUrn(this.pool) {
		pool, err := this.Client.VPool().
			Name(this.pool).
			Query()
		if err != nil {
			return err
		}
		this.pool = pool.Id
	}

	// lookup the group by name
	if this.group != "" && !isStorageOsUrn(this.group) {
		group, err := this.Client.Group().
			Name(this.group).
			Query()
		if err != nil {
			return err
		}
		this.group = group.Id
	}

	return nil
}

func (v Volume) UUID() string {
	tmp := strings.TrimPrefix(v.Id, "urn:storageos:Volume:")
	parts := strings.Split(tmp, ":")
	return parts[0]
}
