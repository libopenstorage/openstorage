package correlation

import "testing"

func TestGetLocalPackage(t *testing.T) {
	//	// From
	// Need this                             /openstorage/pkg/correlation
	// From /go/src/github.com/a/b/vendor/github.com/libopenstorage/openstorage/pkg/correlation
	// Need this                             					   /openstorage/pkg/correlation
	fullDir := "/go/src/github.com/libopenstorage/openstorage/pkg/correlation"
	localDir := getLocalPackage(fullDir)
	expectedDir := "openstorage/pkg/correlation"
	if localDir != expectedDir {
		t.Errorf("local package incorrect. expected: %s, got %s", expectedDir, localDir)
	}

	fullDir = "/go/src/github.com/a/b/vendor/github.com/libopenstorage/openstorage/csi"
	localDir = getLocalPackage(fullDir)
	expectedDir = "openstorage/csi"
	if localDir != expectedDir {
		t.Errorf("local package incorrect. expected: %s, got %s", expectedDir, localDir)
	}
}
