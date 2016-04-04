package coprhd

import (
	"crypto/tls"
	"errors"
	"gopkg.in/jmcvetta/napping.v3"
	"net/http"
	"net/url"
	"strings"
)

const (
	loginUri      = "login"
	proxyTokenUri = "proxytoken"
)

var (
	ErrLoginFailed = errors.New("Login failed")
	ErrAuthFailed  = errors.New("Authorization failed")

	// httpClient is a runtime initialized client that ignores certificate errors
	httpClient *http.Client
)

type (
	Client struct {
		// session is the rest session
		s *napping.Session

		// host is the corphd api url with port (i.e, https://host:port)
		host string

		// proxyToken is the token used for authentication
		proxyToken string

		// lastError is the last API error to occur on the service
		lastError ApiError
	}

	ApiError struct {
		Code        int    `json:"code"`
		Retryable   bool   `json:"retryable"`
		Description string `json:"description"`
		Details     string `json:"details"`
	}
)

func init() {
	httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
}

// NewClient returns a new coprhd rest api client using the specified proxy token
// The proxy token can be generated programmatically using the GetProxyToken method
// or better using the `coprtop` tool
func NewClient(host string, token string) *Client {
	return &Client{
		newAuthSession(token),
		host,
		token,
		ApiError{},
	}
}

// GetProxyToken is a helper function for creating a new proxy token for the user
func GetProxyToken(host string, username string, password string) (string, error) {
	e := ApiError{}

	creds := url.UserPassword(username, password)

	s := &napping.Session{
		Userinfo: creds,
		Client:   httpClient,
	}

	path := buildUrl(host, loginUri)

	resp, err := s.Get(path, nil, nil, &e)
	if err != nil {
		return "", err
	}
	if resp.Status() != http.StatusOK {
		return "", ErrLoginFailed
	}
	if e.Code > 0 {
		return "", errors.New(e.Description)
	}

	// this is the temporary login auth token
	token := resp.HttpResponse().Header.Get("X-SDS-AUTH-TOKEN")

	path = buildUrl(host, proxyTokenUri)

	s = newAuthSession(token)

	resp, err = s.Get(path, nil, nil, &e)
	if err != nil {
		return "", err
	}
	if resp.Status() != http.StatusOK {
		return "", ErrAuthFailed
	}
	if e.Code > 0 {
		return "", errors.New(e.Description)
	}

	// this is the proxy token that should be use with the client
	token = resp.HttpResponse().Header.Get("X-SDS-AUTH-PROXY-TOKEN")

	return token, nil
}

// Copy creates a clone of the client
func (this *Client) Copy() *Client {
	return &Client{this.s, this.host, this.proxyToken, ApiError{}}
}

// LastError return the last ApiError
// If a method returns an error object, you can check this to get more specific details
func (this *Client) LastError() ApiError {
	return this.lastError
}

// PathForResource is a helper method that returns the complete url for a resource
func (this *Client) pathForResource(r string) string {
	return buildUrl(this.host, r)
}

// Get performs an http GET using the underlying http client
func (this *Client) get(r string, p *url.Values, result interface{}) error {
	this.lastError = ApiError{}
	e := ApiError{}
	path := this.pathForResource(r)
	resp, err := this.s.Get(path, p, result, &e)

	if err != nil {
		return err
	}
	if e.Code > 0 {
		this.lastError = e
		return errors.New(e.Description + ":" + e.Details)
	}
	if resp.Status() >= http.StatusBadRequest {
		return errors.New(http.StatusText(resp.Status()))
	}

	return nil
}

// Post performs an http POST using the underlying http client
func (this *Client) post(r string, p, result interface{}) error {
	this.lastError = ApiError{}
	e := ApiError{}
	path := this.pathForResource(r)
	resp, err := this.s.Post(path, p, result, &e)

	if err != nil {
		return err
	}
	if e.Code > 0 {
		this.lastError = e
		return errors.New(e.Description + ":" + e.Details)
	}
	if resp.Status() >= http.StatusBadRequest {
		return errors.New(http.StatusText(resp.Status()))
	}

	return nil
}

func buildUrl(host string, path string) string {
	if !strings.HasSuffix(host, "/") {
		host = host + "/"
	}
	if strings.HasPrefix(path, "/") {
		path = strings.TrimPrefix(path, "/")
	}
	return host + path
}

func newAuthSession(token string) *napping.Session {
	h := http.Header{}

	h.Set("X-SDS-AUTH-TOKEN", token)

	return &napping.Session{
		Client: httpClient,
		Header: &h,
	}
}

func isStorageOsUrn(name string) bool {
	return strings.HasPrefix(name, "urn:storageos:")
}
