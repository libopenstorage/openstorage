package prototime // import "go.pedge.io/proto/time"

// the functionality in here is moving directly to go.pedge.io/pb

import (
	"time"

	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
)

// TimeToTimestamp converts a go Time to a protobuf Timestamp.
func TimeToTimestamp(t time.Time) *timestamp.Timestamp {
	return &timestamp.Timestamp{
		Seconds: t.UnixNano() / int64(time.Second),
		Nanos:   int32(t.UnixNano() % int64(time.Second)),
	}
}

// TimestampToTime converts a protobuf Timestamp to a go Time.
func TimestampToTime(timestamp *timestamp.Timestamp) time.Time {
	if timestamp == nil {
		return time.Unix(0, 0).UTC()
	}
	return time.Unix(
		timestamp.Seconds,
		int64(timestamp.Nanos),
	).UTC()
}

// TimestampLess returns true if i is before j.
func TimestampLess(i *timestamp.Timestamp, j *timestamp.Timestamp) bool {
	if j == nil {
		return false
	}
	if i == nil {
		return true
	}
	if i.Seconds < j.Seconds {
		return true
	}
	if i.Seconds > j.Seconds {
		return false
	}
	return i.Nanos < j.Nanos
}

// Now returns the current time as a protobuf Timestamp.
func Now() *timestamp.Timestamp {
	return TimeToTimestamp(time.Now().UTC())
}

// DurationToProto converts a go Duration to a protobuf Duration.
func DurationToProto(d time.Duration) *duration.Duration {
	return &duration.Duration{
		Seconds: int64(d) / int64(time.Second),
		Nanos:   int32(int64(d) % int64(time.Second)),
	}
}

// DurationFromProto converts a protobuf Duration to a go Duration.
func DurationFromProto(d *duration.Duration) time.Duration {
	if d == nil {
		return 0
	}
	return time.Duration((d.Seconds * int64(time.Second)) + int64(d.Nanos))
}
