/*
Package sdk is the gRPC implementation of the SDK gRPC server
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
package util

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWaitForWithContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*9)
	defer cancel()

	tests := []struct {
		ctx            context.Context
		minTimeout     time.Duration
		maxTimeout     time.Duration
		defaultTimeout time.Duration
		period         time.Duration
		f              func() (bool, error)
		expectFailure  bool
	}{
		{
			ctx:           ctx,
			minTimeout:    1 * time.Second,
			maxTimeout:    1 * time.Hour,
			period:        1 * time.Millisecond,
			f:             func() (bool, error) { return false, nil },
			expectFailure: false,
		},
		{
			ctx:           ctx,
			minTimeout:    10 * time.Second,
			maxTimeout:    1 * time.Hour,
			period:        1 * time.Millisecond,
			f:             func() (bool, error) { return false, nil },
			expectFailure: true,
		},
		{
			ctx:           ctx,
			minTimeout:    1 * time.Second,
			maxTimeout:    2 * time.Second,
			period:        1 * time.Millisecond,
			f:             func() (bool, error) { return false, nil },
			expectFailure: true,
		},
	}

	for _, test := range tests {
		err := WaitForWithContext(
			test.ctx,
			test.minTimeout,
			test.maxTimeout,
			test.defaultTimeout,
			test.period,
			test.f,
		)
		if test.expectFailure {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}
