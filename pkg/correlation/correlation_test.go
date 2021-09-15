package correlation_test

import (
	"bytes"
	"context"
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
	ctx := correlation.WithCorrelationContext(context.Background(), "test_origin")

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
