package coprhd

import (
	"fmt"
	"time"
)

const (
	HostTypeLinux   HostType = "Linux"
	HostTypeWindows HostType = "Windows"
	HostTypeHPUX    HostType = "HPUX"
	HostTypeEsx     HostType = "Esx"
	HostTypeOther   HostType = "Other"

	createHostUri      = "compute/hosts.json"
	queryHostItrUriTpl = "compute/hosts/%s/initiators.json"
	searchHostUri      = "compute/hosts/search.json?"
	queryHostUriTpl    = "compute/hosts/%s.json"
)

type (
	// HostService provides host creation and querying
	HostService struct {
		*Client

		id     string
		name   string
		typ    HostType
		os     string
		tenant string
	}

	// Host represents a physical host resource
	Host struct {
		StorageObject      `json:",inline"`
		Type               HostType `json:"type"`
		OSVersion          string   `json:"os_version,omitempty"`
		HostName           string   `json:"host_name"`
		Port               int      `json:"port_number,omitempty"`
		Username           string   `json:"user_name,omitempty"`
		SSL                bool     `json:"use_ssl,omitempty"`
		Discoverable       bool     `json:"discoverable"`
		RegistrationStatus string   `json:"registration_status"`
		Tenant             Resource `json:"tenant"`
		Cluster            Resource `json:"cluster,omitempty"`
	}

	// HostType is a string type for the host
	HostType string

	createHostReq struct {
		Name         string   `json:"name"`
		Type         HostType `json:"type"`
		OSVersion    string   `json:"os_version,omitempty"`
		HostName     string   `json:"host_name"`
		Port         int      `json:"port_number,omitempty"`
		Tenant       string   `json:"tenant"`
		SSL          bool     `json:"use_ssl,omitempty"`
		Discoverable bool     `json:"discoverable"`
		Username     string   `json:"user_name"`
		Password     string   `json:"password"`
	}

	queryHostItrRes struct {
		Initiators []NamedResource `json:"initiator"`
	}
)

// Host returns an instance of the HostService
func (this *Client) Host() *HostService {
	return &HostService{
		Client: this.Copy(),
	}
}

// Id sets the id for the host query
func (this *HostService) Id(id string) *HostService {
	this.id = id
	return this
}

// Name sets the name for the host creation or query
func (this *HostService) Name(name string) *HostService {
	this.name = name
	return this
}

// Tenant sets the tenant id for the creation
func (this *HostService) Tenant(id string) *HostService {
	this.tenant = id
	return this
}

// Type sets the HostType for the creation
func (this *HostService) Type(t HostType) *HostService {
	this.typ = t
	return this
}

// OSVersion sets the os version string for the creation
func (this *HostService) OSVersion(v string) *HostService {
	this.os = v
	return this
}

// Create creates a new host with the name and host
func (this *HostService) Create(host string) (*Host, error) {
	req := createHostReq{
		Name:         this.name,
		HostName:     host,
		Discoverable: false,
		Type:         this.typ,
		OSVersion:    this.os,
		Tenant:       this.tenant,
	}

	task := Task{}

	err := this.post(createHostUri, &req, &task)
	if err != nil {
		if this.LastError().IsDup() {
			return this.Query()
		}
		return nil, err
	}

	err = this.Task().WaitDone(task.Id, TaskStateReady, time.Second*180)
	if err != nil {
		return nil, err
	}

	this.id = task.Resource.Id

	return this.Query()
}

// Discover creates and attempts to discover a new host
func (this *HostService) Discover(host, username, password string, port int, ssl bool) (*Host, error) {
	req := createHostReq{
		Name:         this.name,
		HostName:     host,
		Port:         port,
		Discoverable: true,
		Username:     username,
		Password:     password,
		SSL:          ssl,
		Type:         this.typ,
		OSVersion:    this.os,
		Tenant:       this.tenant,
	}

	task := Task{}

	err := this.post(createHostUri, &req, &task)
	if err != nil {
		if this.LastError().IsDup() {
			return this.Query()
		}
		return nil, err
	}

	err = this.Task().WaitDone(task.Id, TaskStateReady, time.Second*180)
	if err != nil {
		return nil, err
	}

	this.id = task.Resource.Id

	return this.Query()
}

// Query locates a Host record by id or name
func (this *HostService) Query() (*Host, error) {
	if !isStorageOsUrn(this.id) {
		return this.Search("name=" + this.name)
	}

	path := fmt.Sprintf(queryHostUriTpl, this.id)
	host := Host{}

	err := this.get(path, nil, &host)
	if err != nil {
		return nil, err
	}

	return &host, nil
}

// Search performs a search for the host using the query string
func (this *HostService) Search(query string) (*Host, error) {
	path := searchHostUri + query

	res, err := this.Client.Search(path)
	if err != nil {
		return nil, err
	}

	this.id = res[0].Id

	return this.Query()
}

// Initiators returns a slice of Initiator objects for the host
func (this *HostService) Initiators() ([]Initiator, error) {
	if err := this.queryHostByName(); err != nil {
		return nil, err
	}

	path := fmt.Sprintf(queryHostItrUriTpl, this.id)
	res := queryHostItrRes{}
	itrs := make([]Initiator, 0)

	err := this.get(path, nil, &res)
	if err != nil {
		return nil, err
	}

	for _, i := range res.Initiators {
		itr, err := this.Initiator().
			Id(i.Id).
			Query()

		if err != nil {
			return itrs, err
		}

		itrs = append(itrs, *itr)
	}

	return itrs, nil
}

func (this *HostService) queryHostByName() error {
	if !isStorageOsUrn(this.id) {
		host, err := this.Query()
		if err != nil {
			return err
		}
		this.id = host.Id
		this.name = host.Name
	}
	return nil
}
