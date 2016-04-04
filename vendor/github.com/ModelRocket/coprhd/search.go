package coprhd

import (
	"errors"
)

var (
	ErrResourceNotFound = errors.New("resource not found")
)

type (
	SearchResult struct {
		Resources []Resource `json:"resource"`
	}
)

func (this *Client) Search(path string) ([]Resource, error) {
	result := SearchResult{}

	err := this.get(path, nil, &result)
	if err != nil {
		return nil, err
	}
	if len(result.Resources) < 1 {
		return nil, ErrResourceNotFound
	}

	return result.Resources, nil
}
