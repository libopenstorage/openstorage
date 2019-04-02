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

func TestQueueDepth(t *testing.T) {
	testSpecOptString(t, api.SpecQueueDepth, "10")
}

func TestNodiscard(t *testing.T) {
	testSpecOptString(t, api.SpecNodiscard, "true")

	spec := testSpecFromString(t, api.SpecNodiscard, "true")
	require.True(t, spec.Nodiscard, "failed to parse nodiscard option into spec")

	spec = testSpecFromString(t, api.SpecNodiscard, "false")
	require.False(t, spec.Nodiscard, "failed to parse nodiscard option into spec")
}

func TestEarlyAck(t *testing.T) {
	s := NewSpecHandler()
	spec, _, _, err := s.SpecFromOpts(map[string]string{
		api.SpecEarlyAck: "true",
	})

	require.NoError(t, err)
	ioStrategy := spec.GetIoStrategy()
	require.NotNil(t, ioStrategy)
	require.True(t, ioStrategy.EarlyAck)

	spec, _, _, err = s.SpecFromOpts(map[string]string{
		api.SpecEarlyAck: "false",
	})
	require.NoError(t, err)
	ioStrategy = spec.GetIoStrategy()
	require.NotNil(t, ioStrategy)
	require.False(t, ioStrategy.EarlyAck)

	spec, _, _, err = s.SpecFromOpts(map[string]string{})
	require.Nil(t, spec.GetIoStrategy())
	require.NoError(t, err)

	_, _, _, err = s.SpecFromOpts(map[string]string{
		api.SpecAsyncIo: "blah",
	})
	require.Error(t, err)
	require.Nil(t, spec.GetIoStrategy())

	spec = testSpecFromString(t, api.SpecAsyncIo, "true")
	ioStrategy = spec.GetIoStrategy()
	require.NotNil(t, ioStrategy)
	require.True(t, ioStrategy.AsyncIo)

	spec = testSpecFromString(t, api.SpecAsyncIo, "false")
	ioStrategy = spec.GetIoStrategy()
	require.NotNil(t, ioStrategy)
	require.False(t, ioStrategy.AsyncIo)
}

func TestAsyncIo(t *testing.T) {
	s := NewSpecHandler()
	spec, _, _, err := s.SpecFromOpts(map[string]string{
		api.SpecAsyncIo: "true",
	})

	require.NoError(t, err)
	ioStrategy := spec.GetIoStrategy()
	require.NotNil(t, ioStrategy)
	require.True(t, ioStrategy.AsyncIo)

	spec, _, _, err = s.SpecFromOpts(map[string]string{
		api.SpecAsyncIo: "false",
	})
	require.NoError(t, err)
	ioStrategy = spec.GetIoStrategy()
	require.NotNil(t, ioStrategy)
	require.False(t, ioStrategy.AsyncIo)

	spec, _, _, err = s.SpecFromOpts(map[string]string{})
	require.NoError(t, err)
	require.Nil(t, spec.GetIoStrategy())

	_, _, _, err = s.SpecFromOpts(map[string]string{
		api.SpecAsyncIo: "blah",
	})
	require.Error(t, err)
	require.Nil(t, spec.GetIoStrategy())

	spec = testSpecFromString(t, api.SpecAsyncIo, "true")
	ioStrategy = spec.GetIoStrategy()
	require.NotNil(t, ioStrategy)
	require.True(t, ioStrategy.AsyncIo)

	spec = testSpecFromString(t, api.SpecAsyncIo, "false")
	ioStrategy = spec.GetIoStrategy()
	require.NotNil(t, ioStrategy)
	require.False(t, ioStrategy.AsyncIo)
}

func TestForceUnsupportedFsType(t *testing.T) {
	s := NewSpecHandler()
	spec, _, _, err := s.SpecFromOpts(map[string]string{
		api.SpecForceUnsupportedFsType: "true",
	})
	require.True(t, spec.GetForceUnsupportedFsType())
	require.NoError(t, err)

	spec, _, _, err = s.SpecFromOpts(map[string]string{
		api.SpecForceUnsupportedFsType: "false",
	})
	require.False(t, spec.GetForceUnsupportedFsType())
	require.NoError(t, err)

	spec, _, _, err = s.SpecFromOpts(map[string]string{})
	require.False(t, spec.GetForceUnsupportedFsType())
	require.NoError(t, err)

	_, _, _, err = s.SpecFromOpts(map[string]string{
		api.SpecForceUnsupportedFsType: "blah",
	})
	require.Error(t, err)

	spec = testSpecFromString(t, api.SpecForceUnsupportedFsType, "true")
	require.True(t, spec.ForceUnsupportedFsType)

	spec = testSpecFromString(t, api.SpecForceUnsupportedFsType, "false")
	require.False(t, spec.ForceUnsupportedFsType)

	// Test that it is false when not present
	spec = testSpecFromString(t, api.SpecRack, "ignore")
	require.False(t, spec.ForceUnsupportedFsType)
}

func TestCopyingLabelsFromSpecToLocator(t *testing.T) {
	s := NewSpecHandler()
	opts := map[string]string{
		"hello": "world",
	}
	spec := &api.VolumeSpec{
		VolumeLabels: map[string]string{
			"goodbye": "fornow",
		},
	}
	_, locator, _, err := s.UpdateSpecFromOpts(opts, spec, nil, nil)
	require.NoError(t, err)
	require.Contains(t, locator.VolumeLabels, "hello")
	require.Contains(t, locator.VolumeLabels, "goodbye")
}

func TestGetTokenFromString(t *testing.T) {
	s := NewSpecHandler()

	token := "abcd.xyz.123"

	tokenParsed, ok := s.GetTokenFromString(fmt.Sprintf("token=%s", token))
	require.Equal(t, token, tokenParsed)
	require.Equal(t, ok, true)

	tokenParsed, ok = s.GetTokenFromString(fmt.Sprintf("toabcbn=%s", token))
	require.Equal(t, "", tokenParsed)
	require.Equal(t, ok, false)

}

func TestGetTokenSecretFromString(t *testing.T) {
	s := NewSpecHandler()

	tt := []struct {
		InputSecret    string
		ExpectedSecret string
		Successful     bool
	}{
		{
			"/px/secrets/alpha/token1",
			"px/secrets/alpha/token1",
			true,
		}, {
			"abcd/secrets/alpha//token1/",
			"abcd/secrets/alpha//token1",
			true,
		}, {
			"simplekey",
			"simplekey",
			true,
		},
	}

	for _, tc := range tt {
		str := "name=abcd,token_secret=" + tc.InputSecret + ",token="
		secretParsed, ok := s.GetTokenSecretFromString(str)
		require.Equal(t, tc.ExpectedSecret, secretParsed)
		require.Equal(t, ok, tc.Successful)
	}

	secretParsed, ok := s.GetTokenSecretFromString(fmt.Sprintf("toabcbn_secret=abcd"))
	require.Equal(t, "", secretParsed)
	require.Equal(t, ok, false)

}
