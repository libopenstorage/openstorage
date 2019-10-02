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
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WaitForWithContext waits for a function to complete work with a default timeout or
// a deadline from the context if provided
func WaitForWithContext(
	ctx context.Context,
	minTimeout, maxTimeout, defaultTimeout time.Duration,
	period time.Duration,
	f func() (bool, error),
) error {
	var timeout time.Duration

	// Check if the caller provided a deadline
	d, ok := ctx.Deadline()
	if !ok {
		timeout = defaultTimeout
	} else {
		timeout = d.Sub(time.Now())

		// Determine if it is too short or too long
		if timeout < minTimeout || timeout > maxTimeout {
			return status.Errorf(codes.InvalidArgument,
				"Deadline must be between %v and %v; was: %v", minTimeout, maxTimeout, timeout)
		}
	}

	return WaitFor(timeout, period, f)
}

// WaitFor() waits until f() returns false or err != nil
// f() returns <wait as bool, or err>.
func WaitFor(timeout time.Duration, period time.Duration, f func() (bool, error)) error {
	timeoutChan := time.After(timeout)
	var (
		wait bool = true
		err  error
	)
	for wait {
		select {
		case <-timeoutChan:
			return status.Errorf(codes.DeadlineExceeded, "Timed out")
		default:
			wait, err = f()
			if err != nil {
				return err
			}
			time.Sleep(period)
		}
	}

	return nil
}
