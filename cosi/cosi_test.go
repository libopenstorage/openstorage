package cosi

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/libopenstorage/openstorage/bucket/drivers/fake"
	"github.com/libopenstorage/openstorage/pkg/grpcserver"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type testServer struct {
	conn   *grpc.ClientConn
	server grpcserver.Server
	driver *fake.Fake
}

func newCOSITestServer(t *testing.T) *testServer {
	// Start fake driver
	fakeDriver := fake.New()
	go func() {
		if err := fakeDriver.Start(); err != http.ErrServerClosed {
			logrus.Errorf("failed to start fake driver: %v", err)
		}
	}()

	// Start COSI server
	tempDir := os.TempDir()
	cosisock := tempDir + "/cosi.sock"
	os.Remove(cosisock)
	if err := os.MkdirAll(filepath.Dir(cosisock), 0750); err != nil {
		logrus.Errorf("failed to create COSI sock")
	}
	s, err := NewServer(&Config{
		Driver:  fakeDriver,
		Net:     "tcp",
		Address: "127.0.0.1:0",
	})
	assert.NoError(t, err)

	err = s.Start()
	assert.NoError(t, err)

	// Setup a connection to the driver
	conn, err := grpc.Dial(s.Address(), grpc.WithInsecure())
	assert.NoError(t, err)

	// Return test server
	return &testServer{
		conn:   conn,
		server: s,
		driver: fakeDriver,
	}
}

// Stop stops a given test server
func (ts *testServer) Stop() {
	if err := ts.conn.Close(); err != nil {
		logrus.Errorf("failed to close test server conn: %v", err)
	}
	ts.server.Stop()

	if err := ts.driver.Stop(); err != nil {
		logrus.Errorf("failed to stop driver: %v", err)
	}
}
