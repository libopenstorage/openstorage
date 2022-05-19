/*
Package sdk is the gRPC implementation of the SDK gRPC server
Copyright 2019 Portworx

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
	req := &api.BucketCreateRequest{
		Name: name,
	}

	id := "test_bucket_id"
	testMockResp := &api.BucketCreateResponse{
		BucketId: id,
	}

	// Create CreateBucket response
	s.MockBucketDriver().
		EXPECT().
		CreateBucket(name).
		Return(id, nil).
		Times(1)

	// Setup client
	c := api.NewOpenStorageBucketClient(s.Conn())

	// Get info
	resp, err := c.Create(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, resp.BucketId, testMockResp.BucketId)
}

func TestBucketCreateFailure(t *testing.T) {
	// Create server and client connection
	s := newTestServer(t)
	defer s.Stop()
	time.Sleep(1 * time.Second)

	name := "test_bucket"
	req := &api.BucketCreateRequest{
		Name: name,
	}

	id := "test_bucket_id"

	// Create CreateBucket response
	s.MockBucketDriver().
		EXPECT().
		CreateBucket(name).
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
	req := &api.BucketDeleteRequest{
		BucketId: id,
	}

	// Create DeleteBucket response
	s.MockBucketDriver().
		EXPECT().
		DeleteBucket(id).
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
	req := &api.BucketDeleteRequest{
		BucketId: id,
	}

	// Create DeleteBucket response
	s.MockBucketDriver().
		EXPECT().
		DeleteBucket(id).
		Return(errors.New("failed")).
		Times(1)

	// Setup client
	c := api.NewOpenStorageBucketClient(s.Conn())

	// Get info
	_, err := c.Delete(context.Background(), req)

	assert.Error(t, err)
}
