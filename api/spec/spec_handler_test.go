package spec

import (
	"fmt"
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

func testSpecFromString(t *testing.T, opt string, val string) *api.VolumeSpec {
	s := NewSpecHandler()
	parsed, spec, _, _, _ := s.SpecFromString(fmt.Sprintf("name=volname,foo=bar,%s=%s", opt, val))
	require.True(t, parsed, "Failed to parse spec string")
	return spec
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
