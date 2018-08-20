package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/libopenstorage/openstorage/pkg/storageops"
	"github.com/libopenstorage/openstorage/pkg/storageops/test"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
)

const (
	newDiskSizeInGB = 10
	newDiskPrefix   = "openstorage-test"
)

var diskName = fmt.Sprintf("%s-%s", newDiskPrefix, uuid.New())

func TestAll(t *testing.T) {
	drivers := make(map[string]storageops.Ops)
	diskTemplates := make(map[string]map[string]interface{})

	if d, err := NewEnvClient(); err == nil {
		volType := opsworks.VolumeTypeGp2
		volSize := int64(newDiskSizeInGB)
		zone, _ := storageops.GetEnvValueStrict("AWS_ZONE")
		ebsVol := &ec2.Volume{
			AvailabilityZone: &zone,
			VolumeType:       &volType,
			Size:             &volSize,
		}
		drivers[d.Name()] = d
		diskTemplates[d.Name()] = map[string]interface{}{
			diskName: ebsVol,
		}
	} else {
		t.Skipf("skipping AWS tests as environment is not set...\n")
	}

	test.RunTest(drivers, diskTemplates, t)
}

func TestAwsGetPrefixFromRootDeviceName(t *testing.T) {
	a := &ec2Ops{}
	tests := []struct {
		deviceName     string
		expectedPrefix string
		expectError    bool
	}{
		{
			deviceName:     "/dev/sdb",
			expectedPrefix: "/dev/sd",
			expectError:    false,
		},
		{
			deviceName:     "/dev/xvdd",
			expectedPrefix: "/dev/xvd",
			expectError:    false,
		},
		{
			deviceName:     "/dev/xvdda",
			expectedPrefix: "/dev/xvd",
			expectError:    false,
		},
		{
			deviceName:     "/dev/dda",
			expectedPrefix: "",
			expectError:    true,
		},
		{
			deviceName:     "/dev/hdf",
			expectedPrefix: "/dev/hd",
			expectError:    false,
		},
		{
			deviceName:     "/dev/sys/dev/asdfasdfasdf",
			expectedPrefix: "",
			expectError:    true,
		},
		{
			deviceName:     "",
			expectedPrefix: "",
			expectError:    true,
		},
	}

	for _, test := range tests {
		prefix, err := a.getPrefixFromRootDeviceName(test.deviceName)
		assert.Equal(t, err != nil, test.expectError)
		assert.Equal(t, test.expectedPrefix, prefix)
	}
}
