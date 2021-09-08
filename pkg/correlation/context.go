/*
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
package correlation

import (
	"context"

	"github.com/pborman/uuid"
)

// Component represents a control plane component for
// correlating requests
type Component string

// contextKeyType represents a key for interacting with the
// corrleation context request object
type contextKeyType string

const (
	// ContextKey represents the key for storing and retrieving
	// the correlation context in a context.Context object.
	ContextKey = contextKeyType("correlation-context")

	ComponentUnknown   = Component("unknown")
	ComponentCSIDriver = Component("csi-driver")
	ComponentSDK       = Component("sdk-server")
	ComponentAuth      = Component("openstorage/pkg/auth")
)

// RequestContext represents the context for a given a request.
// A request represents a single action received from an SDK
// user, container orchestrator, or any other request.
type RequestContext struct {
	// ID is a randomly generated UUID per requst
	ID string

	// Origin is the starting point for this request.
	// Examples may include any of the following:
	// pxctl, pxc, kubernetes, CSI, SDK, etc
	Origin Component
}

// WithCorrelationContext returns a new correlation context object
func WithCorrelationContext(ctx context.Context, origin Component) context.Context {
	if v := ctx.Value(ContextKey); v == nil {
		requestContext := &RequestContext{
			ID:     uuid.New(),
			Origin: origin,
		}
		ctx = context.WithValue(ctx, ContextKey, requestContext)
	}

	return ctx
}

// TODO is an alias for context.TODO(), specifically
// for keeping track of areas where we might want to add
// the correlation context.
func TODO() context.Context {
	return context.TODO()
}
