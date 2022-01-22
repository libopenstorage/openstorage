package cosi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	cosi "sigs.k8s.io/container-object-storage-interface-spec"
)

func TestProvisionerGetInfo(t *testing.T) {
	// Create test server
	testServer := newCOSITestServer(t)
	defer testServer.Stop()

	// Test provisioner get info
	cosiClient := cosi.NewIdentityClient(testServer.conn)
	resp, err := cosiClient.ProvisionerGetInfo(context.TODO(), &cosi.ProvisionerGetInfoRequest{})
	assert.NoError(t, err)
	assert.Equal(t, "osd.openstorage.org", resp.GetName())
}
