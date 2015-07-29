package client

import (
	"net/http"
	"net/url"

	"github.com/libopenstorage/openstorage/volume"
)

type Client struct {
	base       *url.URL
	version    string
	httpClient *http.Client
}

func (c *Client) VolumeDriver() volume.VolumeDriver {
	return newVolumeClient(c)
}

func (c *Client) Status() (*Status, error) {
	var status Status
	err := c.Get().UsePath("/status").Do().Unmarshal(&status)
	return &status, err
}

func (c *Client) Get() *Request {
	return NewRequest(&http.Client{}, c.base, "GET", c.version)
}

func (c *Client) Post() *Request {
	return NewRequest(&http.Client{}, c.base, "POST", c.version)
}

func (c *Client) Put() *Request {
	return NewRequest(&http.Client{}, c.base, "PUT", c.version)
}

func (c *Client) Delete() *Request {
	return NewRequest(&http.Client{}, c.base, "DELETE", c.version)
}

func New(host string, version string) (*Client, error) {
	baseURL, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	if baseURL.Path == "" {
		baseURL.Path = "/"
	}
	c := &Client{
		base:       baseURL,
		version:    version,
		httpClient: &http.Client{},
	}
	return c, nil
}
