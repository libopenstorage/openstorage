package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/libopenstorage/openstorage/pkg/storageops"
	"github.com/libopenstorage/openstorage/pkg/storageops/test"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

const (
	newDiskSizeInGB = 10
	newDiskPrefix   = "openstorage-test"
)

var uid string
var diskName = fmt.Sprintf("%s-%s", newDiskPrefix, uid)

func TestAll(t *testing.T) {
	drivers := make(map[string]storageops.Ops)
	diskTemplates := make(map[string]map[string]interface{})

	if d, err := NewEnvClient(); err != ErrAWSEnvNotAvailable {
		volType := opsworks.VolumeTypeGp2
		volSize := int64(newDiskSizeInGB)
		zone := os.Getenv("AWS_ZONE")
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

func init() {
	val, _ := uuid.NewV4()
	uid = val.String()
}
