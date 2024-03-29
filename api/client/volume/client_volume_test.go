package volume

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/libopenstorage/openstorage/api"
	"github.com/stretchr/testify/require"
)

func TestClientTLS(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var vol *api.Volume

		json.NewEncoder(w).Encode(vol)
	}))

	defer ts.Close()

	clnt, err := NewDriverClient(ts.URL, "pxd", "", "")
	require.NoError(t, err)

	clnt.SetTLS(&tls.Config{InsecureSkipVerify: true})

	_, err = VolumeDriver(clnt).Inspect(context.TODO(), []string{"12345"})

	require.NoError(t, err)
}
