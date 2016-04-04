package coprhd

import (
	"fmt"
	"time"
)

const (
	ExportTypeExclusive = "Exclusive"

	createExportUri    = "block/exports.json"
	queryExportUriTpl  = "block/exports/%s.json"
	searchExportUri    = "block/exports/search.json?"
	deleteExportUriTpl = "block/exports/%s/deactivate.json"
)

type (
	// ExportService provices a means for querying or creating volume export groups
	ExportService struct {
		*Client
		id         string
		name       string
		itrs       []string
		project    string
		exportType ExportType
		array      string
		volumes    []ExportVolume
	}

	// Export represents a storage export group
	Export struct {
		StorageObject `json:",inline"`
		Volumes       []ExportVolume  `json:"volumes"`
		Initiators    []Initiator     `json:"initiators"`
		Hosts         []NamedResource `json:"hosts"`
		Clustsers     []NamedResource `json:"clusters"`
		GeneratedName string          `json:"generated_name"`
		PathParams    []string        `json:"path_parameters"`
	}

	// ExportType is a string value of the export type
	ExportType string

	ExportVolume struct {
		ResourceId `json:",inline"`
		Lun        int `json:"lun,omitempty"`
	}

	createExportReq struct {
		Initiators []string       `json:"initiators"`
		Name       string         `json:"name"`
		Project    string         `json:"project"`
		Type       ExportType     `json:"type"`
		VArray     string         `json:"varray"`
		Volumes    []ExportVolume `json:"volumes"`
	}
)

// Export gets an instance to the ExportService
func (this *Client) Export() *ExportService {
	return &ExportService{
		Client:     this.Copy(),
		itrs:       make([]string, 0),
		volumes:    make([]ExportVolume, 0),
		exportType: ExportTypeExclusive,
	}
}

// Id sets the id urn for the export group, use for query, ignored for create
func (this *ExportService) Id(id string) *ExportService {
	this.id = id
	return this
}

// Name sets the name for the export group
func (this *ExportService) Name(name string) *ExportService {
	this.name = name
	return this
}

func (this *ExportService) Initiators(itrs ...string) *ExportService {
	this.itrs = append(this.itrs, itrs...)
	return this
}

func (this *ExportService) Volumes(vols ...string) *ExportService {
	for _, v := range vols {
		this.volumes = append(this.volumes, ExportVolume{ResourceId{v}, 0})
	}
	return this
}

func (this *ExportService) Project(project string) *ExportService {
	this.project = project
	return this
}

func (this *ExportService) Array(array string) *ExportService {
	this.array = array
	return this
}

func (this *ExportService) Type(t ExportType) *ExportService {
	this.exportType = t
	return this
}

// Create creates and export with the specfied name
func (this *ExportService) Create() (*Export, error) {

	if err := this.getExportUrns(); err != nil {
		return nil, err
	}

	req := createExportReq{
		Name:       this.name,
		Initiators: this.itrs,
		Project:    this.project,
		Type:       this.exportType,
		VArray:     this.array,
		Volumes:    this.volumes,
	}

	task := Task{}

	err := this.post(createExportUri, &req, &task)
	if err != nil {
		if this.LastError().IsDup() {
			return this.Query()
		}
		return nil, err
	}

	// wait for the task to complete
	err = this.Task().WaitDone(task.Id, TaskStateReady, time.Second*180)
	if err != nil {
		return nil, err
	}

	this.id = task.Resource.Id

	return this.Query()
}

func (this *ExportService) Query() (*Export, error) {
	if !isStorageOsUrn(this.id) {
		return this.Search("name=" + this.name)
	}

	path := fmt.Sprintf(queryExportUriTpl, this.id)
	exp := Export{}

	err := this.get(path, nil, &exp)
	if err != nil {
		return nil, err
	}

	return &exp, nil
}

func (this *ExportService) Search(query string) (*Export, error) {
	path := searchExportUri + query

	res, err := this.Client.Search(path)
	if err != nil {
		return nil, err
	}

	this.id = res[0].Id

	return this.Query()
}

func (this *ExportService) Delete() error {
	path := fmt.Sprintf(deleteExportUriTpl, this.id)

	task := Task{}

	err := this.post(path, nil, &task)
	if err != nil {
		return err
	}

	return this.Task().WaitDone(task.Id, TaskStateReady, time.Second*180)
}

func (this *ExportService) getExportUrns() error {
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

	// Lookup the intiators
	itrs := []string{}
	for _, i := range this.itrs {
		if !isStorageOsUrn(i) {
			itr, err := this.Client.Initiator().
				Port(i).
				Query()
			if err != nil {
				return err
			}
			i = itr.Id
		}
		itrs = append(itrs, i)
	}
	this.itrs = itrs

	// Look up the volumes
	vols := []ExportVolume{}
	for _, v := range this.volumes {
		if !isStorageOsUrn(v.Id) {
			vol, err := this.Client.Volume().
				Name(v.Id).
				Query()
			if err != nil {
				return err
			}
			v.Id = vol.Id
		}
		vols = append(vols, v)
	}
	this.volumes = vols

	return nil
}
