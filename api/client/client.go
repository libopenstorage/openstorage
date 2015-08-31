package client

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/libopenstorage/openstorage/cluster"
	"github.com/libopenstorage/openstorage/config"
	"github.com/libopenstorage/openstorage/volume"
)

// Client is an HTTP REST wrapper. Use one of Get/Porst/Put/Delete to get a request
// object.
type Client struct {
	base       *url.URL
	version    string
	httpClient *http.Client
}

// VolumeDriver returns a REST wrapper for the VolumeDriver interface.
func (c *Client) VolumeDriver() volume.VolumeDriver {
	return newVolumeClient(c)
}

// ClusterManager returns a REST wrapper for the Cluster interface.
func (c *Client) ClusterManager() cluster.Cluster {
	return newClusterClient(c)
}

// Status sends a Status request at the /status REST endpoint.
func (c *Client) Status() (*Status, error) {
	var status Status
	err := c.Get().UsePath("/status").Do().Unmarshal(&status)
	return &status, err
}

// Get returns a Request object setup for GET call.
func (c *Client) Get() *Request {
	return NewRequest(c.httpClient, c.base, "GET", c.version)
}

// Post returns a Request object setup for POST call.
func (c *Client) Post() *Request {
	return NewRequest(c.httpClient, c.base, "POST", c.version)
}

// Put returns a Request object setup for PUT call.
func (c *Client) Put() *Request {
	return NewRequest(c.httpClient, c.base, "PUT", c.version)
}

// Put returns a Request object setup for DELETE call.
func (c *Client) Delete() *Request {
	return NewRequest(c.httpClient, c.base, "DELETE", c.version)
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
			ret, err := net.DialTimeout("unix", socketPath, timeout)
			return ret, err
		}
		httpTransport.Dial = unixDial
		// Override the main URL object so the HTTP lib won't complain
		u.Scheme = "http"
		u.Host = "unix.sock"
		u.Path = ""
	}
	return &http.Client{Transport: httpTransport}
}

// NewClient returns a new REST client for specified server.
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

// NewDriver returns a new REST client for specified driver.
func NewDriverClient(driverName string) (*Client, error) {
	sockPath := "unix://" + config.DriverAPIBase + driverName + ".sock"
	return NewClient(sockPath, config.Version)
}
