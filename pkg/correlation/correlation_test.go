package correlation_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/libopenstorage/openstorage/pkg/correlation"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewPackageLogger(t *testing.T) {
	clogger := correlation.NewPackageLogger("test")
	clogger.SetReportCaller(true)
	var buf bytes.Buffer
	clogger.SetOutput(&buf)
	clogger.SetLevel(logrus.DebugLevel)
	ctx := correlation.WithCorrelationContext(context.Background(), "test_origin")

	clogger.WithContext(ctx).Info("test info log")
	clogger.WithContext(ctx).Errorf("test error log: %v", errors.New("test err"))
	logStr := buf.String()

	// Log lines are separated by "\n". Break them up into separate strings.
	lines := strings.Split(logStr, "\n")

	// Log line-specific checks
	expectedInfoLog := `level=info msg="test info log"`
	if !strings.Contains(lines[0], expectedInfoLog) {
		t.Fatalf("failed to check for log line %s in %s", expectedInfoLog, lines[0])
	}

	expectedErrorLog := `msg="test error log: test err"`
	if !strings.Contains(lines[1], expectedErrorLog) {
		t.Fatalf("failed to check for log line %s in %s", expectedErrorLog, lines[1])
	}

	// All log lines should have these fields
	for _, logLine := range lines {
		// last line is empty
		if logLine == "" {
			continue
		}
		expectedComponentLog := `component=test`
		if !strings.Contains(logLine, expectedComponentLog) {
			t.Fatalf("failed to check for log line %s in %s", expectedComponentLog, logLine)
		}

		expectedCorrelationLog := `correlation-id=`
		if !strings.Contains(logLine, expectedCorrelationLog) {
			t.Fatalf("failed to check for log line %s in %s", expectedCorrelationLog, logLine)
		}
	}
}

func TestFunctionLogger(t *testing.T) {
	ctx := correlation.WithCorrelationContext(context.Background(), "test_origin")
	correlation.RegisterComponent("register_comp_test")

	clogger := correlation.NewFunctionLogger(ctx)
	clogger.SetReportCaller(true)
	var buf bytes.Buffer
	clogger.SetOutput(&buf)

	clogger.Info("test info log")
	logStr := buf.String()

	expectedInfoLog := `level=info msg="test info log"`
	if !strings.Contains(logStr, expectedInfoLog) {
		t.Fatalf("failed to check for log line %s", expectedInfoLog)
	}

	expectedComponentLog := `component=register_comp_test`
	if !strings.Contains(logStr, expectedComponentLog) {
		t.Fatalf("failed to check for log line %s", expectedComponentLog)
	}

	expectedCorrelationLog := `correlation-id=`
	if !strings.Contains(logStr, expectedCorrelationLog) {
		t.Fatalf("failed to check for log line %s", expectedCorrelationLog)
	}
}

func TestRegisterGlobalLogger(t *testing.T) {
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	logrus.SetReportCaller(true)
	correlation.RegisterGlobalHook()
	ctx := correlation.WithCorrelationContext(context.Background(), "test_origin")

	logrus.WithContext(ctx).Info("test info log")
	logStr := buf.String()

	expectedInfoLog := `level=info msg="test info log"`
	if !strings.Contains(logStr, expectedInfoLog) {
		t.Fatalf("failed to check for log line %s", expectedInfoLog)
	}

	expectedCorrelationLog := `correlation-id=`
	if !strings.Contains(logStr, expectedCorrelationLog) {
		t.Fatalf("failed to check for log line %s", expectedCorrelationLog)
	}
}

func TestWithCorrelationContext(t *testing.T) {
	ctx := correlation.WithCorrelationContext(context.Background(), "test")
	cc, ok := ctx.Value(correlation.ContextKey).(*correlation.RequestContext)
	if !ok {
		t.Error("correlation context not found")
	}

	id := cc.ID
	assert.NotEmpty(t, id)
	fmt.Println(id)

	ctx = correlation.WithCorrelationContext(ctx, "test")
	cc, ok = ctx.Value(correlation.ContextKey).(*correlation.RequestContext)
	if !ok {
		t.Error("correlation context not found")
	}
	newID := cc.ID
	fmt.Println(newID)
	assert.Equal(t, id, newID)
}
