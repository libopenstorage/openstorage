package client

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/libopenstorage/openstorage/config"
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

func newHTTPClient(u *url.URL, tlsConfig *tls.Config, timeout time.Duration) *http.Client {

	httpTransport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	switch u.Scheme {
	default:
		httpTransport.Dial = func(proto, addr string) (net.Conn, error) {
			return net.DialTimeout(proto, addr, timeout)
		}
	case "unix":
		socketPath := u.Path
		unixDial := func(proto, addr string) (net.Conn, error) {
			return net.DialTimeout("unix", socketPath, timeout)
		}
		httpTransport.Dial = unixDial
		// Override the main URL object so the HTTP lib won't complain
		u.Scheme = "http"
		u.Host = "unix.sock"
		u.Path = ""
	}
	return &http.Client{Transport: httpTransport}
}

func NewClient(host string, version string) (*Client, error) {

	baseURL, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	if baseURL.Path == "" {
		baseURL.Path = "/"
	}
	httpClient := newHTTPClient(baseURL, nil, 10*time.Second)
	c := &Client{
		base:       baseURL,
		version:    version,
		httpClient: httpClient,
	}
	return c, nil
}

func NewDriverClient(driverName string) (*Client, error) {
	sockPath := "unix://" + config.DriverApiBase + driverName
	return NewClient(sockPath, config.Version)
}
