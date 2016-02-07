package testing

import (
	"bytes"
	"testing"

	"github.com/libopenstorage/openstorage/pkg/jsonpb"

	"go.pedge.io/pb/go/google/protobuf"
)

func TestTimestamp(t *testing.T) {
	timestamp := google_protobuf.Now()
	buffer := bytes.NewBuffer(nil)
	if err := (&jsonpb.Marshaler{}).Marshal(buffer, timestamp); err != nil {
		t.Fatal(err)
	}
	timestamp2 := &google_protobuf.Timestamp{}
	if err := jsonpb.Unmarshal(buffer, timestamp2); err != nil {
		t.Fatal(err)
	}
	if timestamp.Seconds != timestamp2.Seconds {
		t.Fatalf("%v %v", *timestamp, *timestamp2)
	}
}

func TestFoo(t *testing.T) {
	timestamp := google_protobuf.Now()
	status := Status_STATUS_OK
	foo := &Foo{
		Timestamp: timestamp,
		Status:    status,
	}
	buffer := bytes.NewBuffer(nil)
	if err := (&jsonpb.Marshaler{EnumsAsSimpleStrings: true}).Marshal(buffer, foo); err != nil {
		t.Fatal(err)
	}
	foo2 := &Foo{}
	if err := jsonpb.Unmarshal(buffer, foo2); err != nil {
		t.Fatal(err)
	}
	if foo.Timestamp.Seconds != foo2.Timestamp.Seconds {
		t.Fatalf("%v %v", *foo, *foo2)
	}
	if foo.Status != foo2.Status {
		t.Fatalf("%v %v", *foo, *foo2)
	}
}
