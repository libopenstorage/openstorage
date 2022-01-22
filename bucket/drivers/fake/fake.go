package fake

import (
	"net/http"

	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/libopenstorage/openstorage/bucket"
	"github.com/sirupsen/logrus"
)

var _ bucket.BucketDriver = &Fake{}

type Fake struct {
	backend    *s3mem.Backend
	httpServer *http.Server
}

func New() *Fake {
	backend := s3mem.New()
	return &Fake{
		backend: backend,
		httpServer: &http.Server{
			Addr:    ":8085",
			Handler: gofakes3.New(backend).Server(),
		},
	}
}

// String name representation of driver
func (f *Fake) String() string {
	return "fake"
}

// Start starts a new fake object storage server
func (f *Fake) Start() error {
	logrus.Infof("Starting fake object storage driver on %s", f.httpServer.Addr)
	return f.httpServer.ListenAndServe()
}

// Stop closes the http server for the fake driver
func (f *Fake) Stop() error {
	return f.httpServer.Close()
}

// CreateBucket provisions a new in-memory bucket
func (f *Fake) CreateBucket(name string) (string, error) {
	logrus.Info("bucket_driver.Fake create bucket received")
	return name, f.backend.CreateBucket(name)
}

// DeleteBucket deprovisions an in-memory bucket
func (f *Fake) DeleteBucket(name string) error {
	logrus.Info("bucket_driver.Fake delete bucket received")
	return f.backend.DeleteBucket(name)
}
