/*
Package storagepolicy manages storage policy and apply/validate storage policy restriction
volume operations.
Copyright 2018 Portworx

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package storagepolicy

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"github.com/libopenstorage/openstorage/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPrefixWithName(t *testing.T) {
	assert.Equal(t, prefixWithName("H$ll0_123$"), policyPrefix+"/policies/"+"H$ll0_123$")
}

func TestSdkStoragePolicyCreate(t *testing.T) {
	kv, err := kvdb.New(mem.Name, "policy", []string{}, nil, logrus.Panicf)
	assert.NoError(t, err)
	_, err = Init(kv)

	s, err := Inst()
	assert.NoError(t, err)
	volSpec := &api.VolumeSpecPolicy{
		SizeOpt: &api.VolumeSpecPolicy_Size{
			Size: 1234,
		},
		SharedOpt: &api.VolumeSpecPolicy_Shared{
			Shared: false,
		},
		Sharedv4Opt: &api.VolumeSpecPolicy_Sharedv4{
			Sharedv4: false,
		},
		JournalOpt: &api.VolumeSpecPolicy_Journal{
			Journal: true,
		},
		HaLevelOpt: &api.VolumeSpecPolicy_HaLevel{
			HaLevel: 3,
		},
		SnapshotScheduleOpt: &api.VolumeSpecPolicy_SnapshotSchedule{
			SnapshotSchedule: "freq:periodic\nperiod:120000\n",
		},
	}

	req := &api.SdkOpenStoragePolicyCreateRequest{
		StoragePolicy: &api.SdkStoragePolicy{
			Name:   "testbasicpolicy",
			Policy: volSpec,
		},
	}

	_, err = s.Create(context.Background(), req)
	assert.NoError(t, err)
}

func TestSdkStoragePolicyCreateBadArguments(t *testing.T) {

	s, err := Inst()
	// empty params
	req := &api.SdkOpenStoragePolicyCreateRequest{}
	_, err = s.Create(context.Background(), req)
	assert.Error(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)

	// empty vol specs
	req = &api.SdkOpenStoragePolicyCreateRequest{
		StoragePolicy: &api.SdkStoragePolicy{
			Name: "testbasicpolicy",
		},
	}
	_, err = s.Create(context.Background(), req)
	assert.Error(t, err)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply Volume Specs")
}

func TestSdkStoragePolicyInspect(t *testing.T) {

	s, err := Inst()
	volSpec := &api.VolumeSpecPolicy{
		SizeOpt: &api.VolumeSpecPolicy_Size{
			Size: 2000,
		},
	}

	req := &api.SdkOpenStoragePolicyCreateRequest{
		StoragePolicy: &api.SdkStoragePolicy{
			Name:   "Test_In$pect-123",
			Policy: volSpec,
		},
	}

	_, err = s.Create(context.Background(), req)
	assert.NoError(t, err)

	inspReq := &api.SdkOpenStoragePolicyInspectRequest{
		Name: "Test_In$pect-123",
	}

	resp, err := s.Inspect(context.Background(), inspReq)
	assert.NoError(t, err)
	assert.Equal(t, resp.StoragePolicy.GetName(), inspReq.GetName())
	assert.True(t, reflect.DeepEqual(resp.StoragePolicy.GetPolicy(), req.StoragePolicy.GetPolicy()))
}

func TestSdkStoragePolicyInspectBadArgument(t *testing.T) {

	s, err := Inst()

	volSpec := &api.VolumeSpecPolicy{
		SizeOpt: &api.VolumeSpecPolicy_Size{
			Size: 1234,
		},
	}

	req := &api.SdkOpenStoragePolicyCreateRequest{
		StoragePolicy: &api.SdkStoragePolicy{
			Name:   "testinspect",
			Policy: volSpec,
		},
	}

	_, err = s.Create(context.Background(), req)
	assert.NoError(t, err)

	inspReq := &api.SdkOpenStoragePolicyInspectRequest{}
	_, err = s.Inspect(context.Background(), inspReq)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply a Storage Policy Name")

	inspReq = &api.SdkOpenStoragePolicyInspectRequest{
		Name: "non-existent-name",
	}
	_, err = s.Inspect(context.Background(), inspReq)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.NotFound)
	assert.Contains(t, serverError.Message(), "not found")
}

func TestSdkStoragePolicyUpdate(t *testing.T) {
	s, err := Inst()

	volSpec := &api.VolumeSpecPolicy{
		SizeOpt: &api.VolumeSpecPolicy_Size{
			Size: 1234,
		},
		SharedOpt: &api.VolumeSpecPolicy_Shared{
			Shared: true,
		},
		SizeOperator: api.VolumeSpecPolicy_Maximum,
	}

	req := &api.SdkOpenStoragePolicyCreateRequest{
		StoragePolicy: &api.SdkStoragePolicy{
			Name:   "testupdate",
			Policy: volSpec,
		},
	}

	_, err = s.Create(context.Background(), req)
	assert.NoError(t, err)

	inspReq := &api.SdkOpenStoragePolicyInspectRequest{
		Name: "testupdate",
	}

	oldResp, err := s.Inspect(context.Background(), inspReq)
	assert.NoError(t, err)
	assert.Equal(t, oldResp.StoragePolicy.GetName(), inspReq.GetName())
	assert.True(t, reflect.DeepEqual(oldResp.StoragePolicy.GetPolicy(), req.StoragePolicy.GetPolicy()))

	// update volume
	updateSpec := &api.VolumeSpecPolicy{
		SharedOpt: &api.VolumeSpecPolicy_Shared{
			Shared: false,
		},
		JournalOpt: &api.VolumeSpecPolicy_Journal{
			Journal: false,
		},
		StickyOpt: &api.VolumeSpecPolicy_Sticky{
			Sticky: false,
		},
		HaLevelOperator: api.VolumeSpecPolicy_Minimum,
		HaLevelOpt: &api.VolumeSpecPolicy_HaLevel{
			HaLevel: 2,
		},
		IoProfileOpt: &api.VolumeSpecPolicy_IoProfile{
			IoProfile: api.IoProfile_IO_PROFILE_DB,
		},
		SnapshotScheduleOpt: &api.VolumeSpecPolicy_SnapshotSchedule{
			SnapshotSchedule: "testschedule",
		},
	}

	updateReq := &api.SdkOpenStoragePolicyUpdateRequest{
		StoragePolicy: &api.SdkStoragePolicy{
			Name:   "testupdate",
			Policy: updateSpec,
		},
	}

	_, err = s.Update(context.Background(), updateReq)
	assert.NoError(t, err)

	updatedResp, err := s.Inspect(context.Background(), inspReq)
	assert.NoError(t, err)
	assert.Equal(t, updatedResp.StoragePolicy.GetName(), inspReq.GetName())

	// check indivisual params
	//assert.Equal(t, updatedResp.StoragePolicy.GetPolicy().GetSize(), oldResp.StoragePolicy.GetPolicy().GetSize())
	// check old param updated to new params
	assert.Equal(t, updatedResp.StoragePolicy.GetPolicy().GetShared(), false)
	// check new params
	assert.Equal(t, updatedResp.StoragePolicy.GetPolicy().GetHaLevel(), updateSpec.GetHaLevel())
	assert.Equal(t, updatedResp.StoragePolicy.GetPolicy().GetHaLevelOperator(), api.VolumeSpecPolicy_Minimum)
	assert.Equal(t, updatedResp.StoragePolicy.GetPolicy().GetSticky(), false)
	assert.Equal(t, updatedResp.StoragePolicy.GetPolicy().GetIoProfile(), api.IoProfile_IO_PROFILE_DB)
	// Check not specified operators are nil
	assert.Nil(t, updatedResp.StoragePolicy.GetPolicy().GetEncryptedOpt())
}

func TestSdkStoragePolicyUpdateBadArgument(t *testing.T) {

	s, err := Inst()

	volSpec := &api.VolumeSpecPolicy{
		SizeOpt: &api.VolumeSpecPolicy_Size{
			Size: 1234,
		},
	}

	req := &api.SdkOpenStoragePolicyCreateRequest{
		StoragePolicy: &api.SdkStoragePolicy{
			Name:   "testupdatebadargs",
			Policy: volSpec,
		},
	}

	_, err = s.Create(context.Background(), req)
	assert.NoError(t, err)

	inspReq := &api.SdkOpenStoragePolicyInspectRequest{}
	_, err = s.Inspect(context.Background(), inspReq)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply a Storage Policy Name")

	updateReq := &api.SdkOpenStoragePolicyUpdateRequest{
		StoragePolicy: &api.SdkStoragePolicy{
			Name: "testinspect",
		},
	}
	_, err = s.Update(context.Background(), updateReq)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply Volume Specs")

	updateReq = &api.SdkOpenStoragePolicyUpdateRequest{
		StoragePolicy: &api.SdkStoragePolicy{
			Name:   "non-existant-key",
			Policy: volSpec,
		},
	}
	_, err = s.Update(context.Background(), updateReq)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.NotFound)
	assert.Contains(t, serverError.Message(), "not found")
}
func TestSdkStoragePolicyDelete(t *testing.T) {

	s, err := Inst()

	volSpec := &api.VolumeSpecPolicy{
		SizeOpt: &api.VolumeSpecPolicy_Size{
			Size: 1234,
		},
	}

	req := &api.SdkOpenStoragePolicyCreateRequest{
		StoragePolicy: &api.SdkStoragePolicy{
			Name:   "testdelete",
			Policy: volSpec,
		},
	}

	_, err = s.Create(context.Background(), req)
	assert.NoError(t, err)

	inspReq := &api.SdkOpenStoragePolicyInspectRequest{
		Name: "testdelete",
	}

	resp, err := s.Inspect(context.Background(), inspReq)
	assert.NoError(t, err)
	assert.Equal(t, resp.StoragePolicy.GetName(), inspReq.GetName())
	assert.True(t, reflect.DeepEqual(resp.StoragePolicy.GetPolicy(), req.StoragePolicy.GetPolicy()))
	assert.Nil(t, resp.GetStoragePolicy().GetOwnership())

	delReq := &api.SdkOpenStoragePolicyDeleteRequest{
		Name: "testdelete",
	}
	_, err = s.Delete(context.Background(), delReq)
	assert.NoError(t, err)

	resp, err = s.Inspect(context.Background(), inspReq)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.NotFound)
	assert.Contains(t, serverError.Message(), "not found")
}

func TestSdkStoragePolicyDeleteBadArgument(t *testing.T) {

	s, err := Inst()

	// Create Policy
	volSpec := &api.VolumeSpecPolicy{
		SizeOpt: &api.VolumeSpecPolicy_Size{
			Size: 1234,
		},
	}

	req := &api.SdkOpenStoragePolicyCreateRequest{
		StoragePolicy: &api.SdkStoragePolicy{
			Name:   "testdeleteBad",
			Policy: volSpec,
		},
	}

	_, err = s.Create(context.Background(), req)
	assert.NoError(t, err)

	inspReq := &api.SdkOpenStoragePolicyInspectRequest{
		Name: "testdeleteBad",
	}
	_, err = s.Inspect(context.Background(), inspReq)
	assert.NoError(t, err)

	// empty policy name
	delReq := &api.SdkOpenStoragePolicyDeleteRequest{}
	_, err = s.Delete(context.Background(), delReq)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply a Storage Policy Name")
}

func TestSdkStoragePolicyEnumerate(t *testing.T) {

	s, err := Inst()

	volSpec := &api.VolumeSpecPolicy{
		SizeOpt: &api.VolumeSpecPolicy_Size{
			Size: 8000,
		},
	}
	// Delete any existing policies
	policies, err := s.Enumerate(
		context.Background(),
		&api.SdkOpenStoragePolicyEnumerateRequest{},
	)
	for _, pol := range policies.GetStoragePolicies() {
		_, err := s.Delete(context.Background(), &api.SdkOpenStoragePolicyDeleteRequest{
			Name: pol.GetName(),
		})
		assert.NoError(t, err)
	}

	for i := 1; i <= 10; i++ {
		req := &api.SdkOpenStoragePolicyCreateRequest{
			StoragePolicy: &api.SdkStoragePolicy{
				Name:   "Te$t-Enum_" + strconv.Itoa(i),
				Policy: volSpec,
			},
		}

		_, err = s.Create(context.Background(), req)
		assert.NoError(t, err)
	}

	policies, err = s.Enumerate(
		context.Background(),
		&api.SdkOpenStoragePolicyEnumerateRequest{},
	)

	assert.NoError(t, err)
	assert.Equal(t, 10, len(policies.GetStoragePolicies()))
}

func TestSdkStoragePolicyRestrictions(t *testing.T) {

	s, err := Inst()

	volSpec := &api.VolumeSpecPolicy{
		SizeOpt: &api.VolumeSpecPolicy_Size{
			Size: 8989,
		},
	}

	req := &api.SdkOpenStoragePolicyCreateRequest{
		StoragePolicy: &api.SdkStoragePolicy{
			Name:   "testdefault1",
			Policy: volSpec,
		},
	}

	_, err = s.Create(context.Background(), req)
	assert.NoError(t, err)

	inspReq := &api.SdkOpenStoragePolicyInspectRequest{
		Name: "testdefault1",
	}

	resp, err := s.Inspect(context.Background(), inspReq)
	assert.NoError(t, err)
	assert.Equal(t, resp.StoragePolicy.GetName(), inspReq.GetName())
	assert.True(t, reflect.DeepEqual(resp.StoragePolicy.GetPolicy(), req.StoragePolicy.GetPolicy()))

	setDefaultReq := &api.SdkOpenStoragePolicySetDefaultRequest{
		Name: inspReq.GetName(),
	}
	_, err = s.SetDefault(context.Background(), setDefaultReq)
	assert.NoError(t, err)

	policy, err := s.DefaultInspect(context.Background(), &api.SdkOpenStoragePolicyDefaultInspectRequest{})
	assert.NoError(t, err)
	assert.Equal(t, policy.GetStoragePolicy().GetName(), inspReq.GetName())

	// replace exisiting policy with new one
	updateReq := &api.SdkOpenStoragePolicyCreateRequest{
		StoragePolicy: &api.SdkStoragePolicy{
			Name:   "testdefault2",
			Policy: volSpec,
		},
	}

	_, err = s.Create(context.Background(), updateReq)
	assert.NoError(t, err)

	inspReq = &api.SdkOpenStoragePolicyInspectRequest{
		Name: "testdefault2",
	}

	resp, err = s.Inspect(context.Background(), inspReq)
	assert.NoError(t, err)
	assert.Equal(t, resp.StoragePolicy.GetName(), inspReq.GetName())
	assert.True(t, reflect.DeepEqual(resp.StoragePolicy.GetPolicy(), req.StoragePolicy.GetPolicy()))

	setdefaultReq := &api.SdkOpenStoragePolicySetDefaultRequest{
		Name: inspReq.GetName(),
	}
	_, err = s.SetDefault(context.Background(), setdefaultReq)
	assert.NoError(t, err)

	policy, err = s.DefaultInspect(context.Background(), &api.SdkOpenStoragePolicyDefaultInspectRequest{})
	assert.NoError(t, err)
	assert.Equal(t, policy.StoragePolicy.GetName(), inspReq.GetName())

	// remove default storage policy restriction
	releaseReq := &api.SdkOpenStoragePolicyReleaseRequest{}
	_, err = s.Release(context.Background(), releaseReq)
	assert.NoError(t, err)

	policy, err = s.DefaultInspect(context.Background(), &api.SdkOpenStoragePolicyDefaultInspectRequest{})
	assert.NoError(t, err)
	assert.Nil(t, policy.GetStoragePolicy())
}

func TestSdkStoragePolicyDefaultBadArgument(t *testing.T) {

	s, err := Inst()

	volSpec := &api.VolumeSpecPolicy{
		SizeOpt: &api.VolumeSpecPolicy_Size{
			Size: 8989,
		},
	}

	req := &api.SdkOpenStoragePolicyCreateRequest{
		StoragePolicy: &api.SdkStoragePolicy{
			Name:   "test-default",
			Policy: volSpec,
		},
	}

	_, err = s.Create(context.Background(), req)
	assert.NoError(t, err)

	inspReq := &api.SdkOpenStoragePolicyInspectRequest{
		Name: "test-default",
	}

	resp, err := s.Inspect(context.Background(), inspReq)
	assert.NoError(t, err)
	assert.Equal(t, resp.StoragePolicy.GetName(), inspReq.GetName())
	assert.True(t, reflect.DeepEqual(resp.StoragePolicy.GetPolicy(), req.StoragePolicy.GetPolicy()))

	defaultReq := &api.SdkOpenStoragePolicySetDefaultRequest{}
	_, err = s.SetDefault(context.Background(), defaultReq)
	assert.Error(t, err)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.InvalidArgument)
	assert.Contains(t, serverError.Message(), "Must supply a Storage Policy Name")

	defaultReq = &api.SdkOpenStoragePolicySetDefaultRequest{
		Name: "non-exist-key",
	}
	_, err = s.SetDefault(context.Background(), defaultReq)
	assert.Error(t, err)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.NotFound)
	assert.Contains(t, serverError.Message(), "not found")
}

func TestSdkStoragePolicyDefaultInspect(t *testing.T) {

	s, err := Inst()

	policy, err := s.DefaultInspect(context.Background(), &api.SdkOpenStoragePolicyDefaultInspectRequest{})
	assert.NoError(t, err)
	assert.Nil(t, policy.GetStoragePolicy())

	volSpec := &api.VolumeSpecPolicy{
		SizeOpt: &api.VolumeSpecPolicy_Size{
			Size: 8989,
		},
	}

	req := &api.SdkOpenStoragePolicyCreateRequest{
		StoragePolicy: &api.SdkStoragePolicy{
			Name:   "defaultInspect",
			Policy: volSpec,
		},
	}

	_, err = s.Create(context.Background(), req)
	assert.NoError(t, err)

	inspReq := &api.SdkOpenStoragePolicyInspectRequest{
		Name: "defaultInspect",
	}

	resp, err := s.Inspect(context.Background(), inspReq)
	assert.NoError(t, err)
	assert.Equal(t, resp.StoragePolicy.GetName(), inspReq.GetName())
	assert.True(t, reflect.DeepEqual(resp.StoragePolicy.GetPolicy(), req.StoragePolicy.GetPolicy()))

	defaultReq := &api.SdkOpenStoragePolicySetDefaultRequest{
		Name: inspReq.GetName(),
	}
	_, err = s.SetDefault(context.Background(), defaultReq)
	assert.NoError(t, err)

	policy, err = s.DefaultInspect(context.Background(), &api.SdkOpenStoragePolicyDefaultInspectRequest{})
	assert.NoError(t, err)
	assert.Equal(t, policy.GetStoragePolicy().GetName(), inspReq.GetName())

	// release default storage policy restriction
	releaseReq := &api.SdkOpenStoragePolicyReleaseRequest{}
	_, err = s.Release(context.Background(), releaseReq)
	assert.NoError(t, err)

	policy, err = s.DefaultInspect(context.Background(), &api.SdkOpenStoragePolicyDefaultInspectRequest{})
	assert.NoError(t, err)
	assert.Nil(t, policy.GetStoragePolicy())
}
