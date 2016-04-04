package coprhd

import (
	"fmt"
)

const (
	queryVArrayUriTpl = "vdc/varrays/%s.json"
	searchVArrayUri   = "vdc/varrays/search.json?"
)

type (
	VArrayService struct {
		*Client

		id   string
		name string
	}

	VArrayBlockSettings struct {
		AutoSanZoning bool `json:"auto_san_zoning"`
	}

	VArrayObjectSettings struct {
		DeviceRegistered bool   `json:"device_registered"`
		ProtectionType   string `json:"protection_type"`
	}

	VArray struct {
		StorageObject  `json:",inline"`
		BlockSettings  VArrayBlockSettings  `json:"block_settings"`
		ObjectSettings VArrayObjectSettings `json:"object_settings"`
		AutoSanZoning  bool                 `json:"auto_san_zoning"`
	}
)

func (this *Client) VArray() *VArrayService {
	return &VArrayService{
		Client: this,
	}
}

func (this *VArrayService) Id(id string) *VArrayService {
	this.id = id
	return this
}

func (this *VArrayService) Name(name string) *VArrayService {
	this.name = name
	return this
}

func (this *VArrayService) Query() (*VArray, error) {
	if !isStorageOsUrn(this.id) {
		return this.Search("name=" + this.name)
	}

	path := fmt.Sprintf(queryVArrayUriTpl, this.id)
	v := VArray{}

	err := this.get(path, nil, &v)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

func (this *VArrayService) Search(query string) (*VArray, error) {
	path := searchVArrayUri + query

	res, err := this.Client.Search(path)
	if err != nil {
		return nil, err
	}

	this.id = res[0].Id

	return this.Query()
}
