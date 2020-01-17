package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/libopenstorage/openstorage/api"
	"github.com/stretchr/testify/require"
)

func TestRetryClient(t *testing.T) {
	count := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if count == 0 {
			count++
			w.Header().Add("Retry-After", "5")
			http.Error(w, "Unavailable", http.StatusServiceUnavailable)
			return
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

//func TestRetryDDosClient(t *testing.T) {
//	attempts := 0
//	var wg sync.WaitGroup
//	var attemptLock sync.Mutex
//
//	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		attemptLock.Lock()
//		defer attemptLock.Unlock()
//		attempts++
//		if attempts < 980 {
//			w.Header().Add("Retry-After", "1")
//			http.Error(w, "Unavailable", http.StatusServiceUnavailable)
//			return
//		}
//
//		var dcRes api.VolumeCreateResponse
//		dcRes.VolumeResponse = &api.VolumeResponse{Error: ""}
//
//		json.NewEncoder(w).Encode(&dcRes)
//	}))
//
//	defer ts.Close()
//	clnt, _ := NewClient(ts.URL, "pxd", "")
//
//	var outputErr error
//	for i := 0; i < 1000; i++ {
//		wg.Add(1)
//		go func() {
//			defer wg.Done()
//			outputErr = post(clnt)
//		}()
//	}
//
//	wg.Wait()
//
//	assert.NoError(t, outputErr)
//}

func post(clnt *Client) error {
	request := &api.VolumeCreateRequest{}
	response := &api.VolumeCreateResponse{}
	err := clnt.Post().Body(request).Do().Unmarshal(response)
	if err != nil {
		return err
	}

	return nil
}
