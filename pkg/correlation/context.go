package correlation

import (
	"context"

	"github.com/pborman/uuid"
)

const (
	// ContextKey represents the key for storing and retrieving
	// the correlation context in a context.Context object.
	ContextKey = "correlation-context"
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

// NewContext returns a new correlation context object
func NewContext(ctx context.Context, origin Component) context.Context {
	requestContext := &RequestContext{
		ID:     uuid.New(),
		Origin: origin,
	}

	return context.WithValue(ctx, ContextKey, requestContext)
}
