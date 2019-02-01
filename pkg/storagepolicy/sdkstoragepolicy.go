/*
Package storagepolicy manages storage policy and enforce policy for
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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/portworx/kvdb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/pkg/jsonpb"
)

// SdkPolicyManager is an implementation of the
// Storage Policy Manager for the SDK
type SdkPolicyManager struct {
	kv kvdb.Kvdb
}

const (
	policyPrefix = "/storage/policy"
	policyPath   = "/policies"
	enforcePath  = "/storage/policy/enforce"
)

var (
	// Check interface
	_ PolicyManager = &SdkPolicyManager{}

	inst *SdkPolicyManager
	Inst = func() (PolicyManager, error) {
		return policyInst()
	}
)

func Init(kv kvdb.Kvdb) (PolicyManager,error) {
	if inst != nil {
		return nil,fmt.Errorf("Policy Manager is already initialized")
	}
	if kv == nil {
		return nil, fmt.Errorf("KVDB is not yet initialized.  " +
			"A valid KVDB instance required for the Storage Policy.")
	}

	inst = &SdkPolicyManager{
		kv: kv,
	}

	return inst, nil
}

func policyInst() (PolicyManager, error) {
	if inst == nil {
		return nil, fmt.Errorf("Policy Manager is not initialized")
	}
	return inst, nil
}

// Simple function which creates key for Kvdb
func prefixWithName(name string) string {
	return policyPrefix + policyPath + "/" + name
}

// Create Storage policy
func (p *SdkPolicyManager) Create(
	ctx context.Context,
	req *api.SdkOpenStoragePolicyCreateRequest,
) (*api.SdkOpenStoragePolicyCreateResponse, error) {
	if req.StoragePolicy.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "Must supply a Storage Policy Name")
	} else if req.StoragePolicy.GetPolicy() == nil {
		return nil, status.Error(codes.InvalidArgument, "Must supply Volume Specs")
	}

	// Since VolumeSpecPolicy has oneof method of proto,
	// we need to marshal it into string using protobuf jsonpb
	m := jsonpb.Marshaler{}
	policyStr, err := m.MarshalToString(req.StoragePolicy.GetPolicy())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Json Marshal failed for policy %s: %v", req.StoragePolicy.GetName(), err)
	}

	_, err = p.kv.Create(prefixWithName(req.StoragePolicy.GetName()), policyStr, 0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to save storage policy: %v", err)
	}

	return &api.SdkOpenStoragePolicyCreateResponse{}, nil
}

// Update Storage policy
func (p *SdkPolicyManager) Update(
	ctx context.Context,
	req *api.SdkOpenStoragePolicyUpdateRequest,
) (*api.SdkOpenStoragePolicyUpdateResponse, error) {
	if req.StoragePolicy.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "Must supply a Storage Policy Name")
	}

	if req.StoragePolicy.GetPolicy() == nil {
		return nil, status.Error(codes.InvalidArgument, "Must supply Volume Specs")
	}

	m := jsonpb.Marshaler{}
	policyStr, err := m.MarshalToString(req.StoragePolicy.GetPolicy())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Json Marshal failed for policy %s: %v", req.StoragePolicy.GetName(), err)
	}

	_, err = p.kv.Update(prefixWithName(req.StoragePolicy.GetName()), policyStr, 0)
	if err == kvdb.ErrNotFound {
		return nil, status.Errorf(codes.NotFound, "Storage Policy %s not found", req.StoragePolicy.GetPolicy())
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update storage policy: %v", err)
	}

	return &api.SdkOpenStoragePolicyUpdateResponse{}, nil
}

// Delete storage policy specified by name
func (p *SdkPolicyManager) Delete(
	ctx context.Context,
	req *api.SdkOpenStoragePolicyDeleteRequest,
) (*api.SdkOpenStoragePolicyDeleteResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "Must supply a Storage Policy Name")
	}

	// release enforcement before deleting policy
	policy, err := p.GetEnforcement()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to retrive enforcement details %v", err)
	}

	if policy != nil && policy.GetName() == req.GetName() {
		_, err := p.Release(ctx, &api.SdkOpenStoragePolicyReleaseRequest{})
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Disable enforcement failed with: %v", err)
		}
	}
	_, err = p.kv.Delete(prefixWithName(req.GetName()))
	if err != kvdb.ErrNotFound && err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete Storage Policy %s: %v", req.GetName(), err)
	}

	return &api.SdkOpenStoragePolicyDeleteResponse{}, nil
}

// Inspect storage policy specifed by name
func (p *SdkPolicyManager) Inspect(
	ctx context.Context,
	req *api.SdkOpenStoragePolicyInspectRequest,
) (*api.SdkOpenStoragePolicyInspectResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "Must supply a Storage Policy Name")
	}

	var volSpecs *api.VolumeSpecPolicy
	kvp, err := p.kv.GetVal(prefixWithName(req.GetName()), &volSpecs)
	if err == kvdb.ErrNotFound {
		return nil, status.Errorf(codes.NotFound, "Policy %s not found", req.GetName())
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get policy %s information: %v", req.GetName(), err)
	}

	err = jsonpb.Unmarshal(strings.NewReader(string(kvp.Value)), volSpecs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Json Unmarshal failed for policy %s: %v", req.GetName(), err)
	}

	return &api.SdkOpenStoragePolicyInspectResponse{
		StoragePolicy: &api.SdkStoragePolicy{
			Name:   req.GetName(),
			Policy: volSpecs,
		},
	}, nil
}

// Enumerate all of storage policies
func (p *SdkPolicyManager) Enumerate(
	ctx context.Context,
	req *api.SdkOpenStoragePolicyEnumerateRequest,
) (*api.SdkOpenStoragePolicyEnumerateResponse, error) {
	// get all keyValue pair at /storage/policy/policies
	kvp, err := p.kv.Enumerate(policyPrefix + policyPath)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get policies from database: %v", err)
	}

	policies := make([]*api.SdkStoragePolicy, 0)
	for _, policy := range kvp {
		volSpecs := &api.VolumeSpecPolicy{}
		err = jsonpb.Unmarshal(strings.NewReader(string(policy.Value)), volSpecs)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Json Unmarshal failed for policy %s: %v", policy.Key, err)
		}
		storagePolicy := &api.SdkStoragePolicy{
			Name:   strings.TrimPrefix(policy.Key, policyPrefix+policyPath+"/"),
			Policy: volSpecs,
		}
		policies = append(policies, storagePolicy)
	}

	return &api.SdkOpenStoragePolicyEnumerateResponse{
		StoragePolicies: policies,
	}, nil
}

// Enforce given storage policy
func (p *SdkPolicyManager) Enforce(
	ctx context.Context,
	req *api.SdkOpenStoragePolicyEnforceRequest,
) (*api.SdkOpenStoragePolicyEnforceResponse, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "Must supply a Storage Policy Name")
	}

	// verify policy exists, before enforcing
	policy, err := p.Inspect(ctx,
		&api.SdkOpenStoragePolicyInspectRequest{
			Name: req.GetName(),
		},
	)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Policy with name %s not found", req.GetName())
	}

	policyStr, err := json.Marshal(policy.StoragePolicy.GetName())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Json marshal failed for policy %s :%v", req.GetName(), err)
	}

	_, err = p.kv.Update(enforcePath, policyStr, 0)
	if err == kvdb.ErrNotFound {
		if _, err := p.kv.Create(enforcePath, policyStr, 0); err != nil {
			return nil, status.Errorf(codes.Internal, "Unable to save enforcement details %v", err)
		}
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to enforce policy: %v", err)
	}

	return &api.SdkOpenStoragePolicyEnforceResponse{}, nil
}

// Release storage policy if enforced
func (p *SdkPolicyManager) Release(
	ctx context.Context,
	req *api.SdkOpenStoragePolicyReleaseRequest,
) (*api.SdkOpenStoragePolicyReleaseResponse, error) {
	// empty represents no policy enforcement is enabled
	strB, _ := json.Marshal("")
	_, err := p.kv.Update(enforcePath, strB, 0)
	if err != kvdb.ErrNotFound && err != nil {
		return nil, status.Errorf(codes.Internal, "Disable enforcement failed with: %v", err)
	}

	return &api.SdkOpenStoragePolicyReleaseResponse{}, nil
}

// GetEnforcement return enforced policy details
func (p *SdkPolicyManager) GetEnforcement() (*api.SdkStoragePolicy, error) {
	var policyName string
	var defaultPolicy *api.SdkStoragePolicy

	_, err := p.kv.GetVal(enforcePath, &policyName)
	// enforcePath key is not created
	if err == kvdb.ErrNotFound {
		return defaultPolicy, nil
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to retrive Enforcement details: %v", err)
	}

	// no enforcement found
	if policyName == "" {
		return defaultPolicy, nil
	}

	// retrive enforced policy details
	inspResp, err := p.Inspect(context.Background(),
		&api.SdkOpenStoragePolicyInspectRequest{
			Name: policyName,
		},
	)

	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Policy with name %s not found", defaultPolicy)
	}

	return inspResp.GetStoragePolicy(), nil
}
