package spec

import (
	"fmt"
	"strings"
	"testing"

	"github.com/libopenstorage/openstorage/api"
	"github.com/stretchr/testify/require"
)

func testSpecOptString(t *testing.T, opt string, val string) {
	s := NewSpecHandler()
	parsed, m, _ := s.SpecOptsFromString(fmt.Sprintf("name=volname,foo=bar,%s=%s", opt, val))
	require.True(t, parsed, "Failed to parse spec string")
	parsedVal, ok := m[opt]
	require.True(t, ok, fmt.Sprintf("Failed to set %q string", opt))
	require.Equal(t, parsedVal, val, fmt.Sprintf("Failed to set %q string value %q", opt, val))
}

func testSpecNodeOptString(t *testing.T, opt string, val string) {
	s := NewSpecHandler()
	parsed, m, _ := s.SpecOptsFromString(fmt.Sprintf("name=volname,foo=bar,%s=%s", opt, val))
	require.True(t, parsed, "Failed to parse spec string")

	parsedVal, ok := m[opt]
	require.True(t, ok, fmt.Sprintf("Failed to set %q string", opt))
	parsedVal = strings.Replace(parsedVal, ",", ";", -1)
	require.Equal(t, parsedVal, fmt.Sprintf("%s", val), fmt.Sprintf("Failed to parse string value %q", val))

	spec, _, _, err := s.UpdateSpecFromOpts(m, &api.VolumeSpec{}, &api.VolumeLocator{}, nil)
	require.NoError(t, err)

	nodes := strings.Split(parsedVal, ";")
	for i, node := range nodes {
		require.Equal(t, node, spec.ReplicaSet.Nodes[i])
	}
}

func testSpecFromString(t *testing.T, opt string, val string) *api.VolumeSpec {
	s := NewSpecHandler()
	parsed, spec, _, _, _ := s.SpecFromString(fmt.Sprintf("name=volname,foo=bar,%s=%s", opt, val))
	require.True(t, parsed, "Failed to parse spec string")
	return spec
}

func testSpecFromStringErr(t *testing.T, opt string, errVal string) {
	s := NewSpecHandler()
	parsed, _, _, _, _ := s.SpecFromString(fmt.Sprintf("name=volname,foo=bar,%s=%s", opt, errVal))
	require.False(t, parsed, "Failed to parse spec string")
}

func TestOptJournal(t *testing.T) {
	testSpecOptString(t, api.SpecJournal, "true")

	spec := testSpecFromString(t, api.SpecJournal, "true")
	require.True(t, spec.Journal, "Failed to parse journal option into spec")

	spec = testSpecFromString(t, api.SpecJournal, "false")
	require.False(t, spec.Journal, "Failed to parse journal option into spec")

	spec = testSpecFromString(t, api.SpecSize, "100")
	require.False(t, spec.Journal, "Default journal option spec")
}

func TestOptIoProfile(t *testing.T) {
	testSpecOptString(t, api.SpecIoProfile, "DB")

	spec := testSpecFromString(t, api.SpecIoProfile, "DB")
	require.Equal(t, spec.IoProfile, api.IoProfile(2), "Unexpected io_profile value")

	spec = testSpecFromString(t, api.SpecIoProfile, "db")
	require.Equal(t, spec.IoProfile, api.IoProfile(2), "Unexpected io_profile value")

	testSpecFromStringErr(t, api.SpecIoProfile, "2")
}

func TestOptNodes(t *testing.T) {
	testSpecNodeOptString(t, api.SpecNodes, "node1;node2")
	testSpecNodeOptString(t, api.SpecNodes, "node1")
}
