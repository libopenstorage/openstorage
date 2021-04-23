/*
Package grpcutil is a package for gRPC utilities
Copyright 2021 Portworx

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
package grpcutil

import (
	"context"
	"testing"
	"time"

	"github.com/libopenstorage/openstorage/pkg/defaultcontext"
	"github.com/stretchr/testify/assert"
)

func TestWithDefaultTimeout(t *testing.T) {
	timeout := defaultcontext.Inst().GetDefaultTimeout()

	ctx, _ := WithDefaultTimeout(context.Background())
	deadline, ok := ctx.Deadline()
	assert.True(t, ok)
	time.Sleep(10 * time.Millisecond)
	assert.True(t, deadline.Before(time.Now().Add(timeout)))

	timeout = time.Hour
	defaultcontext.Inst().SetDefaultTimeout(time.Hour)
	ctx, _ = WithDefaultTimeout(context.Background())
	deadline, ok = ctx.Deadline()
	assert.True(t, ok)
	time.Sleep(10 * time.Millisecond)
	assert.True(t, deadline.Before(time.Now().Add(timeout)))
}
