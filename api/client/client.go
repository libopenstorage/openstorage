package client

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var (
	httpCache = make(map[string]*http.Client)
	cacheLock sync.Mutex
)

// NewClient returns a new REST client for specified server.
func NewClient(host, version, userAgent string) (*Client, error) {
	baseURL, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	if baseURL.Path == "" {
		baseURL.Path = "/"
	}
	unix2HTTP(baseURL)
	hClient := getHTTPClient(host)
	if hClient == nil {
		return nil, fmt.Errorf("Unable to parse provided url: %v", host)
	}
	c := &Client{
		host:        host,
		tlsConfig:   nil,
		base:        baseURL,
		version:     version,
		httpClient:  hClient,
		authstring:  "",
		accesstoken: "",
		userAgent:   fmt.Sprintf("%v/%v", userAgent, version),
	}
	c.transport = hClient.Transport
	hClient.Transport = c
	return c, nil
}

// NewAuthClient returns a new REST client for specified server.
func NewAuthClient(host, version, authstring, accesstoken, userAgent string) (*Client, error) {
	baseURL, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	if baseURL.Path == "" {
		baseURL.Path = "/"
	}
	unix2HTTP(baseURL)
	hClient := getHTTPClient(host)
	if hClient == nil {
		return nil, fmt.Errorf("Unable to parse provided url: %v", host)
	}
	c := &Client{
		host:        host,
		tlsConfig:   nil,
		base:        baseURL,
		version:     version,
		httpClient:  hClient,
		authstring:  authstring,
		accesstoken: accesstoken,
		userAgent:   fmt.Sprintf("%v/%v", userAgent, version),
	}
	c.transport = hClient.Transport
	hClient.Transport = c
	return c, nil
}

// GetUnixServerPath returns a unix domain socket prepended with the
// provided path.
func GetUnixServerPath(socketName string, paths ...string) string {
	serverPath := "unix://"
	for _, path := range paths {
		serverPath = serverPath + path
	}
	serverPath = serverPath + socketName + ".sock"
	return serverPath
}

// Client is an HTTP REST wrapper. Use one of Get/Post/Put/Delete to get a request
// object.
type Client struct {
	host        string
	tlsConfig   *tls.Config
	transport   http.RoundTripper
	base        *url.URL
	version     string
	httpClient  *http.Client
	authstring  string
	accesstoken string
	userAgent   string
}

func (c *Client) BaseURL() string {
	return c.base.String()
}

func (c *Client) SetTLS(tlsConfig *tls.Config) {
	transport := &http.Transport{TLSClientConfig: c.tlsConfig}

	// re-assign transport layer defaults
	c.tlsConfig = tlsConfig
	c.transport = transport

	c.httpClient = &http.Client{
		Transport: c,
	}

}

// Versions send a request at the /versions REST endpoint.
func (c *Client) Versions(endpoint string) ([]string, error) {
	versions := []string{}
	err := c.Get().Resource(endpoint + "/versions").Do().Unmarshal(&versions)
	return versions, err
}

// Get returns a Request object setup for GET call.
func (c *Client) Get() *Request {
	return NewRequest(c.httpClient, c.base, http.MethodGet, c.version, c.authstring, c.userAgent)
}

// Post returns a Request object setup for POST call.
func (c *Client) Post() *Request {
	r := NewRequest(c.httpClient, c.base, http.MethodPost, c.version, c.authstring, c.userAgent)
	return r
}

// Put returns a Request object setup for PUT call.
func (c *Client) Put() *Request {
	return NewRequest(c.httpClient, c.base, http.MethodPut, c.version, c.authstring, c.userAgent)
}

// Patch returns a Request object setup for PATCH call.
func (c *Client) Patch() *Request {
	return NewRequest(c.httpClient, c.base, http.MethodPatch, c.version, c.authstring, c.userAgent)
}

// Delete returns a Request object setup for DELETE call.
func (c *Client) Delete() *Request {
	return NewRequest(c.httpClient, c.base, http.MethodDelete, c.version, c.authstring, c.userAgent)
}

func unix2HTTP(u *url.URL) {
	if u.Scheme == "unix" {
		// Override the main URL object so the HTTP lib won't complain
		u.Scheme = "http"
		u.Host = "unix.sock"
		u.Path = ""
	}
}

// shouldRoundTripRetry
func (c *Client) shouldRoundTripRetry(res *http.Response, err error) bool {
	if http.ErrHandlerTimeout == err || res == nil {
		return true
	}
	return res.StatusCode == http.StatusRequestTimeout ||
		res.StatusCode == http.StatusGatewayTimeout
}

// RoundTrip
// When creating the http client we cache the client against the host without any expiration
// when picking the client from cache for any further request
// The cached client resolves the IP that was assigned to the node prior to DHCP update
// A custom round tripper in the transport layer which can invalidate the cache and
// build a new http client in case of a timeout.
// Rebuilding the cache on timeout and retrying will resolve the new IP for that host.
func (c *Client) RoundTrip(req *http.Request) (res *http.Response, err error) {
	res, err = c.transport.RoundTrip(req)

	if c.shouldRoundTripRetry(res, err) {
		retireHTTPClient(c.host)
		c.httpClient = getHTTPClient(c.host)
		if c.tlsConfig != nil {
			c.SetTLS(c.tlsConfig)
		} else {
			c.transport = c.httpClient.Transport
			c.httpClient.Transport = c
		}

		res, err = c.transport.RoundTrip(req)
	}

	return
}

func newHTTPClient(
	u *url.URL,
	tlsConfig *tls.Config,
	timeout time.Duration,
	responseTimeout time.Duration,
) *http.Client {
	httpTransport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	switch u.Scheme {
	case "unix":
		socketPath := u.Path
		unixDial := func(proto, addr string) (net.Conn, error) {
			ret, err := net.DialTimeout("unix", socketPath, timeout)
			return ret, err
		}
		httpTransport.Dial = unixDial
		unix2HTTP(u)
	default:
		httpTransport.Dial = func(proto, addr string) (net.Conn, error) {
			return net.DialTimeout(proto, addr, timeout)
		}
	}

	return &http.Client{Transport: httpTransport, Timeout: responseTimeout}
}

func retireHTTPClient(host string) {
	cacheLock.Lock()
	defer cacheLock.Unlock()
	if _, ok := httpCache[host]; ok {
		delete(httpCache, host)
	}
}

func getHTTPClient(host string) *http.Client {
	cacheLock.Lock()
	defer cacheLock.Unlock()
	c, ok := httpCache[host]
	if !ok {
		u, err := url.Parse(host)
		if err != nil {
			return nil
		}
		if u.Path == "" {
			u.Path = "/"
		}
		c = newHTTPClient(u, nil, 10*time.Second, 5*time.Minute)
		httpCache[host] = c
	}

	return c
}
