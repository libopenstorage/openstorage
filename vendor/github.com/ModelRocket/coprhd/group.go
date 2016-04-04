package coprhd

import (
	"fmt"
)

const (
	queryGroupUriTpl = "block/consistency-groups/%s.json"
	searchGroupUri   = "block/consistency-groups/search.json?"
)

type (
	GroupService struct {
		*Client

		id   string
		name string
	}

	Group struct {
		StorageObject `json:",inline"`
	}
)

func (this *Client) Group() *GroupService {
	return &GroupService{
		Client: this,
	}
}

func (this *GroupService) Id(id string) *GroupService {
	this.id = id
	return this
}

func (this *GroupService) Name(name string) *GroupService {
	this.name = name
	return this
}

func (this *GroupService) Query() (*Group, error) {
	if !isStorageOsUrn(this.id) {
		return this.Search("name=" + this.name)
	}

	path := fmt.Sprintf(queryGroupUriTpl, this.id)
	group := Group{}

	err := this.get(path, nil, &group)
	if err != nil {
		return nil, err
	}

	return &group, nil
}

func (this *GroupService) Search(query string) (*Group, error) {
	path := searchGroupUri + query

	res, err := this.Client.Search(path)
	if err != nil {
		return nil, err
	}

	this.id = res[0].Id

	return this.Query()
}
