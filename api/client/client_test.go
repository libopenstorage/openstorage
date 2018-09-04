package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRetryClient(t *testing.T) {
	count := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if count == 0 {
			count++
			w.Header().Add("Retry-After", "5")
			http.Error(w, "Unavailable", http.StatusServiceUnavailable)
		}
	}))

	defer ts.Close()

	clnt, _ := NewClient(ts.URL, "pxd", "")

	start := time.Now()
	clnt.Get().Resource("").Do()
	now := time.Now()
	elapsed := now.Sub(start)
	require.True(t, elapsed >= time.Duration(5*time.Second))
}
