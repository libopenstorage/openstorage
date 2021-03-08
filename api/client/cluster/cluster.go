package cluster

import (
	"crypto/tls"
	"net/url"

	"github.com/libopenstorage/openstorage/api/client"
	"github.com/libopenstorage/openstorage/cluster"
)

const (
	// OsdSocket is the unix socket for cluster apis
	OsdSocket  = "osd"
	APIVersion = cluster.APIVersion
)

// ClusterManager returns a REST wrapper for the Cluster interface.
func ClusterManager(c *client.Client) cluster.Cluster {
	return newClusterClient(c)
}

// NewAuthClusterClient returns a new REST client.
// host: REST endpoint [http://<ip>:<port> OR unix://<path-to-unix-socket>]. default: [unix://var/lib/osd/cluster/osd.sock]
// version: Cluster API version
func NewAuthClusterClient(host, version string, authstring string, accesstoken string) (*client.Client, error) {
	if host == "" {
		host = client.GetUnixServerPath(OsdSocket, cluster.APIBase)
	}

	if version == "" {
		// Set the default version
		version = cluster.APIVersion
	}

	return client.NewAuthClient(host, version, authstring, accesstoken, "")
}

// NewInsecureTLSAuthClusterClient returns a new REST client that will skip TLS verification for https
// host: REST endpoint [http(s)://<ip>:<port>]
// version: ClusterAPI version
func NewInsecureTLSAuthClusterClient(host, version, auth string, accesstoken string) (*client.Client, error) {
	if host == "" {
		host = client.GetUnixServerPath(OsdSocket, cluster.APIBase)
	}

	if version == "" {
		// Set the default version
		version = cluster.APIVersion
	}

	var (
		skipTLSVerify bool
	)
	u, err := url.Parse(host)
	if err == nil && len(u.Scheme) != 0 {
		if u.Scheme == "https" {
			// We don't support cert validation yet
			skipTLSVerify = true
		} else if u.Scheme != "http" {
			// In certain cases like AWS ELB - ae20db68c7cb34616b16837ab395fe9c-1428320453.us-east-2.elb.amazonaws.com
			// url.Parse returns scheme as the actual endpoint
			host = "http://" + host
		} // else u.Scheme == http
	} else {
		host = "http://" + host
	}

	clnt, err := client.NewAuthClient(host, version, auth, accesstoken, "")
	if err != nil {
		return nil, err
	}
	if skipTLSVerify {
		clnt.SetTLS(&tls.Config{InsecureSkipVerify: true})
	}
	return clnt, nil
}

// NewClusterClient returns a new REST client.
// host: REST endpoint [http://<ip>:<port> OR unix://<path-to-unix-socket>]. default: [unix://var/lib/osd/cluster/osd.sock]
// version: Cluster API version
func NewClusterClient(host, version string) (*client.Client, error) {
	if host == "" {
		host = client.GetUnixServerPath(OsdSocket, cluster.APIBase)
	}

	if version == "" {
		// Set the default version
		version = cluster.APIVersion
	}

	return client.NewClient(host, version, "")
}

// GetSupportedClusterVersions returns a list of supported versions of the Cluster API
// host: REST endpoint [http://<ip>:<port> OR unix://<path-to-unix-socket>]. default: [unix://var/lib/osd/cluster/osd.sock]
func GetSupportedClusterVersions(host string) ([]string, error) {
	if host == "" {
		host = client.GetUnixServerPath(OsdSocket, cluster.APIBase)
	}
	client, err := client.NewClient(host, "", "")
	if err != nil {
		return []string{}, err
	}
	versions, err := client.Versions("cluster")
	if err != nil {
		return []string{}, err
	}
	return versions, nil
}
