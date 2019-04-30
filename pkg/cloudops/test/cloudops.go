package test

import (
	"testing"

	"github.com/libopenstorage/openstorage/pkg/cloudops"
	"github.com/stretchr/testify/require"
)

func TestCloudOps(
	t *testing.T,
	drivers map[string]cloudops.Ops,
) {
	for _, d := range drivers {
		info, err := d.InspectSelf()
		require.NoError(t, err, "failed to inspect instance")
		require.NotNil(t, info, "got nil instance info from inspect")

		groupInfo, err := d.InspectSelfInstanceGroup()
		require.NoError(t, err, "failed to inspect instance group")
		require.NotNil(t, groupInfo, "got nil instance group info from inspect")
	}
}
