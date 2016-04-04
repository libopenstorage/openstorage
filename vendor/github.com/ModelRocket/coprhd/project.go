package coprhd

import (
	"fmt"
)

const (
	queryProjectUriTpl = "projects/%s.json"
	searchProjectUri   = "projects/search.json?"
)

type (
	// ProjectService provides an interface for querying projects
	ProjectService struct {
		*Client

		id   string
		name string
	}

	// Project represents a storage project object
	Project struct {
		StorageObject `json:",inline"`
	}
)

// Project returns an instances the ProjectService
func (this *Client) Project() *ProjectService {
	return &ProjectService{
		Client: this,
	}
}

// Id sets the id urn for the query
func (this *ProjectService) Id(id string) *ProjectService {
	this.id = id
	return this
}

// Name sets the name for the query
func (this *ProjectService) Name(name string) *ProjectService {
	this.name = name
	return this
}

// Query locates a Project by id or name
func (this *ProjectService) Query() (*Project, error) {
	if !isStorageOsUrn(this.id) {
		return this.Search("name=" + this.name)
	}

	path := fmt.Sprintf(queryProjectUriTpl, this.id)
	proj := Project{}

	err := this.get(path, nil, &proj)
	if err != nil {
		return nil, err
	}

	return &proj, nil
}

// Search locates a project using the specified query
func (this *ProjectService) Search(query string) (*Project, error) {
	path := searchProjectUri + query

	res, err := this.Client.Search(path)
	if err != nil {
		return nil, err
	}

	this.id = res[0].Id

	return this.Query()
}
