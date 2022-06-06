/*
Package sdk is the gRPC implementation of the SDK gRPC server
Copyright 2022 Portworx

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
package sdk

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/libopenstorage/openstorage/api"
	"github.com/stretchr/testify/assert"
)

func TestBucketCreateSuccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	time.Sleep(1 * time.Second)

	name := "test_bucket"
	region := "eu-central-1"
	req := &api.BucketCreateRequest{
		Name:   name,
		Region: region,
	}

	id := "test_bucket_id"
	testMockResp := &api.BucketCreateResponse{
		BucketId: id,
	}

	// Create CreateBucket response
	s.MockBucketDriver().
		EXPECT().
		CreateBucket(name, region).
		Return(id, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageBucketClient(s.Conn())

	// Get info
	resp, err := c.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, resp.BucketId, testMockResp.BucketId)
}

func TestBucketCreateRegionMissing(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	time.Sleep(1 * time.Second)

	name := "test_bucket"
	req := &api.BucketCreateRequest{
		Name: name,
	}

	// Setup client
	c := api.NewOpenStorageBucketClient(s.Conn())

	// Get info
	_, err := c.Create(context.Background(), req)

	assert.Error(t, err)
}

func TestBucketCreateFailure(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	time.Sleep(1 * time.Second)

	name := "test_bucket"
	region := "eu-central-1"
	req := &api.BucketCreateRequest{
		Name:   name,
		Region: region,
	}

	id := "test_bucket_id"
	// Create CreateBucket response
	s.MockBucketDriver().
		EXPECT().
		CreateBucket(name, region).
		Return(id, errors.New("failed")).
		Times(1)

	// Setup client
	c := api.NewOpenStorageBucketClient(s.Conn())

	// Get info
	_, err := c.Create(context.Background(), req)

	assert.Error(t, err)
}

func TestBucketDeleteSuccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	time.Sleep(1 * time.Second)

	id := "test_bucket_id"
	region := "eu-central-1"
	req := &api.BucketDeleteRequest{
		BucketId:    id,
		Region:      region,
		ClearBucket: true,
	}

	// Create DeleteBucket response
	s.MockBucketDriver().
		EXPECT().
		DeleteBucket(id, region, true).
		Return(nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageBucketClient(s.Conn())

	// Get info
	_, err := c.Delete(context.Background(), req)

	assert.NoError(t, err)
}

func TestBucketDeleteFailure(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	time.Sleep(1 * time.Second)

	id := "test_bucket_id"
	region := "eu-central-1"
	req := &api.BucketDeleteRequest{
		Region:   region,
		BucketId: id,
	}

	// Create DeleteBucket response
	s.MockBucketDriver().
		EXPECT().
		DeleteBucket(id, region, false).
		Return(errors.New("failed")).
		Times(1)

	// Setup client
	c := api.NewOpenStorageBucketClient(s.Conn())

	// Get info
	_, err := c.Delete(context.Background(), req)

	assert.Error(t, err)
}

func TestBucketDeleteFailureRegionMissing(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	time.Sleep(1 * time.Second)

	id := "test_bucket_id"
	req := &api.BucketDeleteRequest{
		BucketId: id,
	}

	// Setup client
	c := api.NewOpenStorageBucketClient(s.Conn())

	// Get info
	_, err := c.Delete(context.Background(), req)

	assert.Error(t, err)
}

func TestBucketGrantAccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	time.Sleep(1 * time.Second)

	id := "test_bucket_id"
	accountName := "test_account_name"
	accessPolicy := "access_policy"
	req := &api.BucketGrantAccessRequest{
		BucketId:     id,
		AccountName:  accountName,
		AccessPolicy: accessPolicy,
	}

	accountId := "test_bucket_id"
	cred := "account_credentials"
	testMockResp := &api.BucketGrantAccessResponse{
		AccountId:   accountId,
		Credentials: cred,
	}

	// Create CreateBucket response
	s.MockBucketDriver().
		EXPECT().
		GrantBucketAccess(id, accountName, accessPolicy).
		Return(accountId, cred, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageBucketClient(s.Conn())

	// Get info
	resp, err := c.GrantAccess(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, resp.AccountId, testMockResp.AccountId)
	assert.Equal(t, resp.Credentials, testMockResp.Credentials)
}

func TestBucketRevokeAccess(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	time.Sleep(1 * time.Second)

	id := "test_bucket_id"
	accountId := "test_bucket_id"
	req := &api.BucketRevokeAccessRequest{
		BucketId:  id,
		AccountId: accountId,
	}
	// Create CreateBucket response
	s.MockBucketDriver().
		EXPECT().
		RevokeBucketAccess(id, accountId).
		Return(nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageBucketClient(s.Conn())

	// Get info
	_, err := c.RevokeAccess(context.Background(), req)

	assert.NoError(t, err)
}
