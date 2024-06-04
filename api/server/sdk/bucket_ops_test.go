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
	"github.com/libopenstorage/openstorage/bucket"
	"github.com/stretchr/testify/assert"
)

func TestBucketCreateSuccess(t *testing.T) {
	t.Skip()
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	time.Sleep(1 * time.Second)

	name := "test_bucket"
	region := "eu-central-1"
	req := &api.BucketCreateRequest{
		Name:                      name,
		Region:                    region,
		AnonymousBucketAccessMode: api.AnonymousBucketAccessMode_Private,
	}

	id := "test_bucket_id"
	testMockResp := &api.BucketCreateResponse{
		BucketId: id,
	}

	// Create CreateBucket response
	s.MockBucketDriver().
		EXPECT().
		CreateBucket(name, region, "", api.AnonymousBucketAccessMode_Private).
		Return(id, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageBucketClient(s.Conn())

	// Get info
	resp, err := c.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, resp.BucketId, testMockResp.BucketId)
}

func TestBucketCreateSuccessRegionMissingPureFB(t *testing.T) {
	t.Skip()
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	time.Sleep(1 * time.Second)

	name := "test_bucket"
	req := &api.BucketCreateRequest{
		Name:                      name,
		AnonymousBucketAccessMode: api.AnonymousBucketAccessMode_Private,
	}

	id := "test_bucket_id"
	testMockResp := &api.BucketCreateResponse{
		BucketId: id,
	}

	// Return PureFBDriver drive string
	s.MockBucketDriver().
		EXPECT().
		String().
		Return(PureFBDriver).
		Times(1)

	// Create CreateBucket response
	s.MockBucketDriver().
		EXPECT().
		CreateBucket(name, "default", "", api.AnonymousBucketAccessMode_Private).
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
	t.Skip()
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	time.Sleep(1 * time.Second)

	name := "test_bucket"
	req := &api.BucketCreateRequest{
		Name: name,
	}

	// Return Non PureFBDriver name as String() response
	s.MockBucketDriver().
		EXPECT().
		String().
		Return("S3Driver").
		Times(1)
	// Setup client
	c := api.NewOpenStorageBucketClient(s.Conn())

	// Get info
	_, err := c.Create(context.Background(), req)
	assert.Error(t, err)
}

func TestBucketCreateFailure(t *testing.T) {
	t.Skip()
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	time.Sleep(1 * time.Second)

	name := "test_bucket"
	region := "eu-central-1"
	req := &api.BucketCreateRequest{
		Name:                      name,
		Region:                    region,
		AnonymousBucketAccessMode: api.AnonymousBucketAccessMode_Private,
	}

	id := "test_bucket_id"
	// Create CreateBucket response
	s.MockBucketDriver().
		EXPECT().
		CreateBucket(name, region, "", api.AnonymousBucketAccessMode_Private).
		Return(id, errors.New("failed")).
		Times(1)

	// Setup client
	c := api.NewOpenStorageBucketClient(s.Conn())

	// Get info
	_, err := c.Create(context.Background(), req)

	assert.Error(t, err)
}

func TestBucketDeleteSuccessRegionMissingPureFB(t *testing.T) {
	t.Skip()
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	time.Sleep(1 * time.Second)

	id := "test_bucket_id"
	req := &api.BucketDeleteRequest{
		BucketId:    id,
		ClearBucket: true,
	}

	// Return PureFBDriver drive string
	s.MockBucketDriver().
		EXPECT().
		String().
		Return("PureFBDriver").
		Times(1)

	// Create DeleteBucket response
	s.MockBucketDriver().
		EXPECT().
		DeleteBucket(id, "default", "", true).
		Return(nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageBucketClient(s.Conn())

	// Get info
	_, err := c.Delete(context.Background(), req)

	assert.NoError(t, err)
}

func TestBucketDeleteSuccess(t *testing.T) {
	t.Skip()
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
		DeleteBucket(id, region, "", true).
		Return(nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageBucketClient(s.Conn())

	// Get info
	_, err := c.Delete(context.Background(), req)

	assert.NoError(t, err)
}

func TestBucketDeleteFailure(t *testing.T) {
	t.Skip()
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
		DeleteBucket(id, region, "", false).
		Return(errors.New("failed")).
		Times(1)

	// Setup client
	c := api.NewOpenStorageBucketClient(s.Conn())

	// Get info
	_, err := c.Delete(context.Background(), req)

	assert.Error(t, err)
}

func TestBucketDeleteFailureRegionMissing(t *testing.T) {
	t.Skip()
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	time.Sleep(1 * time.Second)

	id := "test_bucket_id"
	req := &api.BucketDeleteRequest{
		BucketId: id,
	}

	// Return Non PureFBDriver name as String() response
	s.MockBucketDriver().
		EXPECT().
		String().
		Return("S3Driver").
		Times(1)
	// Setup client
	c := api.NewOpenStorageBucketClient(s.Conn())

	// Get info
	_, err := c.Delete(context.Background(), req)

	assert.Error(t, err)
}

func TestBucketGrantAccess(t *testing.T) {
	t.Skip()
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	time.Sleep(1 * time.Second)

	id := "test_bucket_id"
	accountName := "test_account_name"
	accessPolicy := "{\"Statement\":{\"Effect\":\"Allow\",\"Action\":\"*\",\"Resource\":\"*\"}}"
	req := &api.BucketGrantAccessRequest{
		BucketId:     id,
		AccountName:  accountName,
		AccessPolicy: accessPolicy,
	}

	bucketCredentials := &bucket.BucketAccessCredentials{
		AccessKeyId:     "YOUR-ACCESSKEYID",
		SecretAccessKey: "YOUR-SECRETACCESSKEY",
	}

	// Create CreateBucket response
	s.MockBucketDriver().
		EXPECT().
		GrantBucketAccess(id, accountName, accessPolicy).
		Return(accountName, bucketCredentials, nil).
		Times(1)

	testMockResp := &api.BucketGrantAccessResponse{
		AccountId: accountName,
		Credentials: &api.BucketAccessCredentials{
			AccessKeyId:     "YOUR-ACCESSKEYID",
			SecretAccessKey: "YOUR-SECRETACCESSKEY",
		},
	}

	// Setup client
	c := api.NewOpenStorageBucketClient(s.Conn())

	// Get info
	resp, err := c.GrantAccess(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, resp.AccountId, testMockResp.AccountId)
	assert.Equal(t, resp.Credentials.AccessKeyId, testMockResp.Credentials.AccessKeyId)
	assert.Equal(t, resp.Credentials.SecretAccessKey, testMockResp.Credentials.SecretAccessKey)
}

func TestBucketRevokeAccess(t *testing.T) {
	t.Skip()
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
