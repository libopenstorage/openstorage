package cli

import (
	"testing"

	"github.com/libopenstorage/openstorage/api"
	"github.com/stretchr/testify/require"
)

func TestCmdMarshalProto(t *testing.T) {
	volumeSpec := &api.VolumeSpec{
		Size: 64,
		Format: api.FSType_FS_TYPE_EXT4,
	}
	data := cmdMarshalProto(volumeSpec)
	require.Equal(
		t,
		`{
 "ephemeral": false,
 "size": "64",
 "format": "ext4",
 "block_size": "0",
 "ha_level": "0",
 "cos": 0,
 "dedupe": false,
 "snapshot_interval": 0
}`,
		data,
	)
}
