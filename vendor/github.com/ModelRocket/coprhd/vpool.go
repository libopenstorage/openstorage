package coprhd

import (
	"fmt"
)

const (
	queryVPoolUriTpl = "block/vpools/%s.json"
	searchVPoolUri   = "block/vpools/search.json?"
)

type (
	VPoolService struct {
		*Client

		id   string
		name string
	}

	VPool struct {
		StorageObject `json:",inline"`
		Protocols     []InitiatorType `json:"protocols"`
	}
)

func (this *Client) VPool() *VPoolService {
	return &VPoolService{
		Client: this,
	}
}

func (this *VPoolService) Id(id string) *VPoolService {
	this.id = id
	return this
}

func (this *VPoolService) Name(name string) *VPoolService {
	this.name = name
	return this
}

func (this *VPoolService) Query() (*VPool, error) {
	if !isStorageOsUrn(this.id) {
		return this.Search("name=" + this.name)
	}

	path := fmt.Sprintf(queryVPoolUriTpl, this.id)
	v := VPool{}

	err := this.get(path, nil, &v)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

func (this *VPoolService) Search(query string) (*VPool, error) {
	path := searchVPoolUri + query

	res, err := this.Client.Search(path)
	if err != nil {
		return nil, err
	}

	this.id = res[0].Id

	return this.Query()
}

func (this *VPool) IsBlock() bool {
	return this.Type == "block"
}

func (this *VPool) HasProtocol(p InitiatorType) bool {
	for _, t := range this.Protocols {
		if t == p {
			return true
		}
	}
	return false
}
