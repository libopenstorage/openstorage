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
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/pkg/auth"
	"github.com/portworx/kvdb"
	"github.com/portworx/kvdb/mem"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestVolSpecToSdkStoragePolicy(t *testing.T) {
	kv, err := kvdb.New(mem.Name, "policy", []string{}, nil, logrus.Panicf)
	assert.NoError(t, err)

	m := jsonpb.Marshaler{OrigName: true}
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

	policyStr, err := m.MarshalToString(volSpec)
	assert.NoError(t, err)

	_, err = kv.Create(prefixWithName("testkey1"), policyStr, 0)
	assert.NoError(t, err)
	_, err = kv.Create(prefixWithName("testkey2"), policyStr, 0)
	assert.NoError(t, err)

	_, err = Init(kv)
	assert.NoError(t, err)

	s, err := Inst()
	assert.NoError(t, err)

	inspReq := &api.SdkOpenStoragePolicyInspectRequest{
		Name: "testkey1",
	}

	resp, err := s.Inspect(context.Background(), inspReq)
	assert.NoError(t, err)
	assert.Equal(t, resp.GetStoragePolicy().GetName(), "testkey1")
	assert.Equal(t, resp.GetStoragePolicy().GetPolicy().GetSize(), volSpec.GetSize())
}
func TestSdkStoragePolicyDeleteWithOwnership(t *testing.T) {
	kv, err := kvdb.New(mem.Name, "policy", []string{}, nil, logrus.Panicf)
	assert.NoError(t, err)
	_, err = Init(kv)

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
			Ownership: &api.Ownership{
				Owner: "testowner",
				Acls: &api.Ownership_AccessControl{
					Collaborators: map[string]api.Ownership_AccessType{
						"notmyname": api.Ownership_Write,
						"trusted":   api.Ownership_Admin,
					},
				},
			},
		},
	}

	// Create contexts
	ctxNoAuth := context.Background()
	ctxWithNotOwner := auth.ContextSaveUserInfo(context.Background(), &auth.UserInfo{
		Username: "notmyname",
	})
	ctxWithTrusted := auth.ContextSaveUserInfo(context.Background(), &auth.UserInfo{
		Username: "trusted",
	})

	_, err = s.Create(ctxWithTrusted, req)
	assert.NoError(t, err)

	inspReq := &api.SdkOpenStoragePolicyInspectRequest{
		Name: "testdelete",
	}

	resp, err := s.Inspect(ctxNoAuth, inspReq)
	assert.NoError(t, err)
	assert.Equal(t, resp.StoragePolicy.GetName(), inspReq.GetName())
	assert.True(t, reflect.DeepEqual(resp.StoragePolicy.GetPolicy(), req.StoragePolicy.GetPolicy()))
	assert.NotNil(t, resp.GetStoragePolicy().GetOwnership())

	delReq := &api.SdkOpenStoragePolicyDeleteRequest{
		Name: "testdelete",
	}

	// test: should not work, auth with collaborator
	// with non-api.Ownship_Admin rights
	_, err = s.Delete(ctxWithNotOwner, delReq)
	assert.Error(t, err)

	// -- test: should work, no auth and owned vol
	_, err = s.Delete(ctxNoAuth, delReq)
	assert.NoError(t, err)

	_, err = s.Create(ctxWithTrusted, req)
	assert.NoError(t, err)
	// test:  should work, auth with trusted admin collaborator
	_, err = s.Delete(ctxWithTrusted, delReq)
	assert.NoError(t, err)
	resp, err = s.Inspect(ctxNoAuth, inspReq)
	serverError, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.NotFound)
	assert.Contains(t, serverError.Message(), "not found")

	// Check for public storage policy
	pubSP := &api.SdkOpenStoragePolicyCreateRequest{
		StoragePolicy: &api.SdkStoragePolicy{
			Name:   "testdeletePublicStoragePolicy",
			Policy: volSpec,
		},
	}

	_, err = s.Create(ctxNoAuth, pubSP)
	assert.NoError(t, err)

	inspReq = &api.SdkOpenStoragePolicyInspectRequest{
		Name: "testdeletePublicStoragePolicy",
	}

	resp, err = s.Inspect(ctxNoAuth, inspReq)
	assert.NoError(t, err)
	assert.Equal(t, resp.StoragePolicy.GetName(), inspReq.GetName())
	assert.True(t, reflect.DeepEqual(resp.StoragePolicy.GetPolicy(), pubSP.StoragePolicy.GetPolicy()))
	// Ownership should be nil
	assert.Nil(t, resp.GetStoragePolicy().GetOwnership())

	// -- test: should work, no auth and public storage policy
	delReq = &api.SdkOpenStoragePolicyDeleteRequest{
		Name: "testdeletePublicStoragePolicy",
	}
	_, err = s.Delete(ctxNoAuth, delReq)
	assert.NoError(t, err)
	resp, err = s.Inspect(ctxNoAuth, inspReq)
	serverError, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, serverError.Code(), codes.NotFound)
	assert.Contains(t, serverError.Message(), "not found")

	// -- test: should work , public storage policy and auth user
	_, err = s.Create(ctxNoAuth, pubSP)
	assert.NoError(t, err)

	_, err = s.Delete(ctxWithNotOwner, delReq)
	assert.NoError(t, err)
}

func TestSdkStoragePolicyUpdateWithOwnership(t *testing.T) {
	s, err := Inst()

	volSpec := &api.VolumeSpecPolicy{
		SizeOpt: &api.VolumeSpecPolicy_Size{
			Size: 1234,
		},
	}

	req := &api.SdkOpenStoragePolicyCreateRequest{
		StoragePolicy: &api.SdkStoragePolicy{
			Name:   "testUpdate",
			Policy: volSpec,
			Ownership: &api.Ownership{
				Owner: "testownerupdate",
				Acls: &api.Ownership_AccessControl{
					Collaborators: map[string]api.Ownership_AccessType{
						"readuser":  api.Ownership_Read,
						"notmyname": api.Ownership_Write,
						"trusted":   api.Ownership_Admin,
					},
				},
			},
		},
	}

	// Create contexts
	ctxNoAuth := context.Background()
	ctxWithRead := auth.ContextSaveUserInfo(context.Background(), &auth.UserInfo{
		Username: "readuser",
	})
	ctxWithNotOwner := auth.ContextSaveUserInfo(context.Background(), &auth.UserInfo{
		Username: "notmyname",
	})
	ctxWithTrusted := auth.ContextSaveUserInfo(context.Background(), &auth.UserInfo{
		Username: "trusted",
	})

	_, err = s.Create(ctxWithTrusted, req)
	assert.NoError(t, err)

	inspReq := &api.SdkOpenStoragePolicyInspectRequest{
		Name: "testUpdate",
	}

	resp, err := s.Inspect(ctxNoAuth, inspReq)
	assert.NoError(t, err)
	assert.Equal(t, resp.StoragePolicy.GetName(), inspReq.GetName())
	assert.True(t, reflect.DeepEqual(resp.StoragePolicy.GetPolicy(), req.StoragePolicy.GetPolicy()))
	assert.NotNil(t, resp.GetStoragePolicy().GetOwnership())

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
			Name:   "testUpdate",
			Policy: updateSpec,
		},
	}

	// -- test: should not work, collaborator with no write access
	_, err = s.Update(ctxWithRead, updateReq)
	assert.Error(t, err)
	// should allow to update with noauth
	_, err = s.Update(ctxNoAuth, updateReq)
	assert.NoError(t, err)

	updatedResp, err := s.Inspect(ctxWithRead, inspReq)
	assert.NoError(t, err)
	assert.Equal(t, updatedResp.GetStoragePolicy().GetPolicy().GetHaLevel(), updateSpec.GetHaLevel())
	// test: should work, auth with collaborator
	// with non-api.Ownship_Write rights
	_, err = s.Update(ctxWithNotOwner, updateReq)
	assert.NoError(t, err)

	// -- test: should work, no auth and owned vol
	_, err = s.Update(ctxWithTrusted, updateReq)
	assert.NoError(t, err)
}

func TestSdkStoragePolicyPolicyRestrictionWithOwnership(t *testing.T) {
	s, err := Inst()

	volSpec := &api.VolumeSpecPolicy{
		SizeOpt: &api.VolumeSpecPolicy_Size{
			Size: 1234,
		},
	}

	req := &api.SdkOpenStoragePolicyCreateRequest{
		StoragePolicy: &api.SdkStoragePolicy{
			Name:   "test$etDefault",
			Policy: volSpec,
			Ownership: &api.Ownership{
				Owner: "testownerupdate",
				Acls: &api.Ownership_AccessControl{
					Collaborators: map[string]api.Ownership_AccessType{
						"readuser":  api.Ownership_Read,
						"notmyname": api.Ownership_Write,
						"trusted":   api.Ownership_Admin,
					},
				},
			},
		},
	}

	// Create contexts
	ctxNoAuth := context.Background()
	ctxWithRead := auth.ContextSaveUserInfo(context.Background(), &auth.UserInfo{
		Username: "readuser",
	})
	ctxWithNotOwner := auth.ContextSaveUserInfo(context.Background(), &auth.UserInfo{
		Username: "notmyname",
	})
	ctxWithTrusted := auth.ContextSaveUserInfo(context.Background(), &auth.UserInfo{
		Username: "trusted",
		Claims: auth.Claims{
			Groups: []string{"*"},
		},
	})
	_, err = s.Create(ctxWithTrusted, req)
	assert.NoError(t, err)

	inspReq := &api.SdkOpenStoragePolicyInspectRequest{
		Name: "test$etDefault",
	}

	resp, err := s.Inspect(ctxWithRead, inspReq)
	assert.NoError(t, err)
	assert.Equal(t, resp.StoragePolicy.GetName(), inspReq.GetName())
	assert.True(t, reflect.DeepEqual(resp.StoragePolicy.GetPolicy(), req.StoragePolicy.GetPolicy()))
	assert.NotNil(t, resp.GetStoragePolicy().GetOwnership())

	defaultReq := &api.SdkOpenStoragePolicySetDefaultRequest{
		Name: inspReq.GetName(),
	}

	// test should error ,read access user can;t set policy restriction
	_, err = s.SetDefault(ctxWithRead, defaultReq)
	assert.Error(t, err)

	// test should error ,write access user can;t set policy restriction
	_, err = s.SetDefault(ctxWithNotOwner, defaultReq)
	assert.Error(t, err)

	// test should work ,admin access user can set policy restriction
	_, err = s.SetDefault(ctxWithTrusted, defaultReq)
	assert.NoError(t, err)

	policy, err := s.DefaultInspect(ctxNoAuth, &api.SdkOpenStoragePolicyDefaultInspectRequest{})
	assert.NoError(t, err)
	assert.Equal(t, policy.GetStoragePolicy().GetName(), inspReq.GetName())

	// release default storage policy restriction by read access user
	releaseReq := &api.SdkOpenStoragePolicyReleaseRequest{}
	_, err = s.Release(ctxWithRead, releaseReq)
	assert.Error(t, err)
	// release by user with write access
	_, err = s.Release(ctxWithNotOwner, releaseReq)
	assert.Error(t, err)

	// release by user with Admin access
	_, err = s.Release(ctxWithTrusted, releaseReq)
	assert.NoError(t, err)

	// test should work ,owner can set policy restriction
	_, err = s.SetDefault(ctxWithTrusted, defaultReq)
	assert.NoError(t, err)

	// release default storage policy restriction
	releaseReq = &api.SdkOpenStoragePolicyReleaseRequest{}
	_, err = s.Release(ctxNoAuth, releaseReq)
	assert.NoError(t, err)

}
