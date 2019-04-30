package gcp

import (
	"strings"
	"testing"

	"github.com/libopenstorage/openstorage/pkg/cloudops"
	"github.com/libopenstorage/openstorage/pkg/cloudops/test"
	"github.com/libopenstorage/openstorage/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestAll(t *testing.T) {
	drivers := make(map[string]cloudops.Ops)
	_, err := util.GetEnvValueStrict("GCE_INSTANCE_PROJECT")
	if err != nil {
		if strings.Contains(err.Error(), "is not set") {
			t.Skipf("Skipping GCP cloudops test as test env variables are not set")
		}
	}

	ops, err := NewGCPOps()
	require.NoError(t, err, "failed ot create GCP ops instance")

	drivers["gcp"] = ops
	test.TestCloudOps(t, drivers)
}
