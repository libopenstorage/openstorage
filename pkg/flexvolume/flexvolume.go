/*
Package flexvolume implements utility code for Kubernetes flexvolumes.

https://github.com/kubernetes/kubernetes/pull/13840
https://github.com/kubernetes/kubernetes/tree/master/examples/flexvolume
*/
package flexvolume

// Client is the client for a flexvolume implementation.
//
// It is both called from the wrapper cli tool, and implemented by a given implementation.
type Client interface {
	Init() error
	Attach(jsonOptions map[string]interface{}) error
	Detach(mountDevice string) error
	Mount(targetMountDir string, mountDevice string, jsonOptions map[string]interface{}) error
	Unmount(mountDir string) error
}

// NewClient returns a new Client for the given APIClient.
func NewClient(apiClient APIClient) Client {
	return nil
}

// NewLocalAPIClient returns a new APIClient for the given APIServer.
func NewLocalAPIClient(apiServer APIServer) APIClient {
	return nil
}

// NewAPIServer returns a new APIServer for the given Client.
func NewAPIServer(client Client) APIServer {
	return nil
}
