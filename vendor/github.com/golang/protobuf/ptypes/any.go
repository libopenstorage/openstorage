// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ptypes

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	anypb "github.com/golang/protobuf/ptypes/any"
)

const urlPrefix = "type.googleapis.com/"

// AnyMessageName returns the message name contained in an anypb.Any message.
// Most type assertions should use the Is function instead.
func AnyMessageName(any *anypb.Any) (string, error) {
	name, err := anyMessageName(any)
	return string(name), err
}
func anyMessageName(any *anypb.Any) (protoreflect.FullName, error) {
	if any == nil {
		return "", fmt.Errorf("message is nil")
	}
	name := protoreflect.FullName(any.TypeUrl)
	if i := strings.LastIndex(any.TypeUrl, "/"); i >= 0 {
		name = name[i+len("/"):]
	}
	if !name.IsValid() {
		return "", fmt.Errorf("message type url %q is invalid", any.TypeUrl)
	}
	return name, nil
}

// MarshalAny marshals the given message m into an anypb.Any message.
func MarshalAny(m proto.Message) (*anypb.Any, error) {
	switch dm := m.(type) {
	case DynamicAny:
		m = dm.Message
	case *DynamicAny:
		if dm == nil {
			return nil, proto.ErrNil
		}
		m = dm.Message
	}
	b, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	return &anypb.Any{TypeUrl: urlPrefix + proto.MessageName(m), Value: b}, nil
}

// Empty returns a new message of the type specified in an anypb.Any message.
// It returns protoregistry.NotFound if the corresponding message type could not
// be resolved in the global registry.
func Empty(any *anypb.Any) (proto.Message, error) {
	name, err := anyMessageName(any)
	if err != nil {
		return nil, err
	}
	mt, err := protoregistry.GlobalTypes.FindMessageByName(name)
	if err != nil {
		return nil, err
	}
	return proto.MessageV1(mt.New().Interface()), nil
}

// UnmarshalAny unmarshals the encoded value contained in the anypb.Any message
// into the provided message m. It returns an error if the target message
// does not match the type in the Any message or if an unmarshal error occurs.
//
// The target message m may be a *DynamicAny message. If the underlying message
// type could not be resolved, then this returns protoregistry.NotFound.
func UnmarshalAny(any *anypb.Any, m proto.Message) error {
	if dm, ok := m.(*DynamicAny); ok {
		if dm.Message == nil {
			var err error
			dm.Message, err = Empty(any)
			if err != nil {
				return err
			}
		}
		m = dm.Message
	}

	anyName, err := AnyMessageName(any)
	if err != nil {
		return err
	}
	msgName := proto.MessageName(m)
	if anyName != msgName {
		return fmt.Errorf("mismatched message type: got %q want %q", anyName, msgName)
	}
	return proto.Unmarshal(any.Value, m)
}

// Is returns true if any value contains a given message type.
func Is(any *any.Any, pb proto.Message) bool {
	// The following is equivalent to AnyMessageName(any) == proto.MessageName(pb),
	// but it avoids scanning TypeUrl for the slash.
	if any == nil {
		return false
	}
	name := proto.MessageName(pb)
	prefix := len(any.TypeUrl) - len(name)
	return prefix >= 1 && any.TypeUrl[prefix-1] == '/' && any.TypeUrl[prefix:] == name
}
