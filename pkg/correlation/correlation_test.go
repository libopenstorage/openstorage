package correlation_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/libopenstorage/openstorage/pkg/correlation"
)

// TestNewLogger mainly is for testing API
func TestNewLogger(t *testing.T) {
	clogger := correlation.NewLogger("test")

	var buf bytes.Buffer
	clogger.SetOutput(&buf)
	clogger.AddHook(&correlation.LogHook{})
	ctx := correlation.NewContext(context.Background(), "test_origin")

	clogger.WithContext(ctx).Info("test info log")
	logStr := buf.String()

	expectedInfoLog := `level=info msg="test info log"`
	if !strings.Contains(logStr, expectedInfoLog) {
		t.Fatalf("failed to check for log line %s", expectedInfoLog)
	}

	expectedComponentLog := `component=test`
	if !strings.Contains(logStr, expectedComponentLog) {
		t.Fatalf("failed to check for log line %s", expectedComponentLog)
	}

	expectedCorrelationLog := `correlation-id=`
	if !strings.Contains(logStr, expectedCorrelationLog) {
		t.Fatalf("failed to check for log line %s", expectedCorrelationLog)
	}
}
