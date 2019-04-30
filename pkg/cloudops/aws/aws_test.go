package aws

import (
	"strings"
	"testing"

	"github.com/libopenstorage/openstorage/pkg/cloudops"
	"github.com/libopenstorage/openstorage/pkg/cloudops/test"
	"github.com/stretchr/testify/require"
)

func TestAll(t *testing.T) {
	drivers := make(map[string]cloudops.Ops)
	ops, err := NewAWSOps()
	if err != nil {
		if strings.Contains(err.Error(), "is not set") {
			t.Skipf("Skipping AWS cloudops test as test env variables are not set")
		}
	}
	require.NoError(t, err, "failed ot create AWS ops instance")

	drivers["aws"] = ops
	test.TestCloudOps(t, drivers)
}
