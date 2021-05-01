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
		api.SpecEarlyAck: "blah",
	})
	require.Error(t, err)
	require.Nil(t, spec.GetIoStrategy())

	spec = testSpecFromString(t, api.SpecEarlyAck, "true")
	ioStrategy = spec.GetIoStrategy()
	require.NotNil(t, ioStrategy)
	require.True(t, ioStrategy.EarlyAck)

	spec = testSpecFromString(t, api.SpecEarlyAck, "false")
	ioStrategy = spec.GetIoStrategy()
	require.NotNil(t, ioStrategy)
	require.False(t, ioStrategy.EarlyAck)
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

func TestDirectIo(t *testing.T) {
	s := NewSpecHandler()
	spec, _, _, err := s.SpecFromOpts(map[string]string{
		api.SpecDirectIo: "true",
	})

	require.NoError(t, err)
	ioStrategy := spec.GetIoStrategy()
	require.NotNil(t, ioStrategy)
	require.True(t, ioStrategy.DirectIo)

	spec, _, _, err = s.SpecFromOpts(map[string]string{
		api.SpecDirectIo: "false",
	})
	require.NoError(t, err)
	ioStrategy = spec.GetIoStrategy()
	require.NotNil(t, ioStrategy)
	require.False(t, ioStrategy.DirectIo)

	spec, _, _, err = s.SpecFromOpts(map[string]string{})
	require.NoError(t, err)
	require.Nil(t, spec.GetIoStrategy())

	_, _, _, err = s.SpecFromOpts(map[string]string{
		api.SpecDirectIo: "blah",
	})
	require.Error(t, err)
	require.Nil(t, spec.GetIoStrategy())

	spec = testSpecFromString(t, api.SpecDirectIo, "true")
	ioStrategy = spec.GetIoStrategy()
	require.NotNil(t, ioStrategy)
	require.True(t, ioStrategy.DirectIo)

	spec = testSpecFromString(t, api.SpecDirectIo, "false")
	ioStrategy = spec.GetIoStrategy()
	require.NotNil(t, ioStrategy)
	require.False(t, ioStrategy.DirectIo)
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

func TestGetTokenSecretContextFromString(t *testing.T) {
	s := NewSpecHandler()

	tt := []struct {
		InputName       string
		ExpectedRequest api.TokenSecretContext
		Successful      bool
	}{
		{
			"name=abcd,token_secret=/px/secrets/alpha/token1," +
				"token_secret_namespace=ns2,abcd=sd,token_secret_public_data=y," +
				"token_secret_custom_data=n",
			api.TokenSecretContext{
				SecretName:      "px/secrets/alpha/token1",
				SecretNamespace: "ns2",
			},
			true,
		}, {
			"name=abcd,token_secret=abcd/secrets/alpha//token1/",
			api.TokenSecretContext{
				SecretName:      "abcd/secrets/alpha//token1",
				SecretNamespace: "",
			},
			true,
		}, {
			"name=abcd,token_secret=simplekey",
			api.TokenSecretContext{
				SecretName:      "simplekey",
				SecretNamespace: "",
			},
			true,
		},
	}

	for _, tc := range tt {
		secretParsed, ok := s.GetTokenSecretContextFromString(tc.InputName)
		require.Equal(t, tc.Successful, ok)
		require.Equal(t, tc.ExpectedRequest, *secretParsed)
	}

	_, ok := s.GetTokenSecretContextFromString(fmt.Sprintf("toabcbn_secret=abcd"))
	require.Equal(t, false, ok)

}

func TestOptProxyRepl(t *testing.T) {
	testSpecOptString(t, api.SpecProxyWrite, "true")

	spec := testSpecFromString(t, api.SpecProxyWrite, "true")
	require.True(t, spec.ProxyWrite, "Failed to parse proxy_write option into spec")

	spec = testSpecFromString(t, api.SpecProxyWrite, "false")
	require.False(t, spec.ProxyWrite, "Failed to parse proxy_write option into spec")
}

func TestExportSpec(t *testing.T) {
	s := NewSpecHandler()
	spec, _, _, err := s.SpecFromOpts(map[string]string{
		api.SpecExportProtocol: api.SpecExportProtocolPXD,
	})
	require.NoError(t, err)
	exportSpec := spec.GetExportSpec()
	require.NotNil(t, exportSpec)
	require.Equal(t, api.ExportProtocol_PXD, exportSpec.ExportProtocol)

	spec, _, _, err = s.SpecFromOpts(map[string]string{
		api.SpecExportProtocol: api.SpecExportProtocolNFS,
	})
	require.NoError(t, err)
	exportSpec = spec.GetExportSpec()
	require.NotNil(t, exportSpec)
	require.Equal(t, api.ExportProtocol_NFS, exportSpec.ExportProtocol)

	spec, _, _, err = s.SpecFromOpts(map[string]string{
		api.SpecExportProtocol: api.SpecExportProtocolISCSI,
	})
	require.NoError(t, err)
	exportSpec = spec.GetExportSpec()
	require.NotNil(t, exportSpec)
	require.Equal(t, api.ExportProtocol_ISCSI, exportSpec.ExportProtocol)

	spec, _, _, err = s.SpecFromOpts(map[string]string{
		api.SpecExportProtocol: api.SpecExportProtocolCustom,
	})
	require.NoError(t, err)
	exportSpec = spec.GetExportSpec()
	require.NotNil(t, exportSpec)
	require.Equal(t, api.ExportProtocol_CUSTOM, exportSpec.ExportProtocol)

	spec, _, _, err = s.SpecFromOpts(map[string]string{
		api.SpecExportProtocol: "invalid",
	})
	require.Error(t, err)

	spec = testSpecFromString(t, api.SpecExportProtocol, api.SpecExportProtocolPXD)
	exportSpec = spec.GetExportSpec()
	require.NotNil(t, exportSpec)
	require.Equal(t, api.ExportProtocol_PXD, exportSpec.ExportProtocol)

	spec = testSpecFromString(t, api.SpecExportProtocol, api.SpecExportProtocolNFS)
	exportSpec = spec.GetExportSpec()
	require.NotNil(t, exportSpec)
	require.Equal(t, api.ExportProtocol_NFS, exportSpec.ExportProtocol)

	spec = testSpecFromString(t, api.SpecExportProtocol, api.SpecExportProtocolCustom)
	exportSpec = spec.GetExportSpec()
	require.NotNil(t, exportSpec)
	require.Equal(t, api.ExportProtocol_CUSTOM, exportSpec.ExportProtocol)

	spec, _, _, err = s.SpecFromOpts(map[string]string{
		api.SpecExportOptions: "exportOptions",
	})
	exportSpec = spec.GetExportSpec()
	require.NotNil(t, exportSpec)
	require.Equal(t, "exportOptions", exportSpec.ExportOptions)
}

func TestPureBackendSpec(t *testing.T) {
	s := NewSpecHandler()
	spec, _, _, err := s.SpecFromOpts(map[string]string{
		api.SpecBackendType: api.SpecBackendPureBlock,
	})
	require.NoError(t, err)
	proxySpec := spec.GetProxySpec()
	require.NotNil(t, proxySpec)
	require.Equal(t, api.SpecBackendPureBlock, proxySpec.ProxyProtocol.SimpleString())
	require.True(t, proxySpec.IsPureBackend())

	spec, _, _, err = s.SpecFromOpts(map[string]string{
		api.SpecBackendType: api.SpecBackendPureFile,
	})
	require.NoError(t, err)
	proxySpec = spec.GetProxySpec()
	require.NotNil(t, proxySpec)
	require.Equal(t, api.SpecBackendPureFile, proxySpec.ProxyProtocol.SimpleString())
	require.True(t, proxySpec.IsPureBackend())

	rule := "*(rw)"
	spec, _, _, err = s.SpecFromOpts(map[string]string{
		api.SpecPureFileExportRules: rule,
	})
	require.NoError(t, err)
	proxySpec = spec.GetProxySpec()
	require.NotNil(t, proxySpec)
	require.NotNil(t, proxySpec.PureFileSpec)
	require.Equal(t, rule, proxySpec.PureFileSpec.ExportRules)
	require.False(t, proxySpec.IsPureBackend())

	_, _, _, err = s.SpecFromOpts(map[string]string{
		api.SpecBackendType: "unknown_backend",
	})
	require.Error(t, err, "Failed to parse backend parameter")
}

func TestXattr(t *testing.T) {
	testSpecOptString(t, api.SpecCowOnDemand, "true")

	spec := testSpecFromString(t, api.SpecRack, "ignore")
	require.Equal(t, api.Xattr_COW_ON_DEMAND, spec.Xattr)

	spec = testSpecFromString(t, api.SpecCowOnDemand, "false")
	require.Equal(t, api.Xattr_UNSPECIFIED, spec.Xattr)

	spec = testSpecFromString(t, api.SpecCowOnDemand, "true")
	require.Equal(t, api.Xattr_COW_ON_DEMAND, spec.Xattr)

}

func TestMountOptions(t *testing.T) {
	testSpecOptString(t, api.SpecMountOptions, "k")

	spec := testSpecFromString(t, api.SpecRack, "ignore")
	require.Nil(t, spec.MountOptions)

	spec = testSpecFromString(t, api.SpecMountOptions, "k1;k2:v2")
	require.NotNil(t, spec.MountOptions)
	require.Equal(t, len(spec.MountOptions.Options), 2)
	val, ok := spec.MountOptions.Options["k1"]
	require.True(t, ok)
	require.Equal(t, val, "")
	val, ok = spec.MountOptions.Options["k2"]
	require.True(t, ok)
	require.Equal(t, val, "v2")
	_, ok = spec.MountOptions.Options["k3"]
	require.False(t, ok)
}

func TestSharedv4MountOptions(t *testing.T) {
	testSpecOptString(t, api.SpecSharedv4MountOptions, "k")

	spec := testSpecFromString(t, api.SpecRack, "ignore")
	require.Nil(t, spec.MountOptions)

	spec = testSpecFromString(t, api.SpecSharedv4MountOptions, "k1;k2:v2")
	require.NotNil(t, spec.Sharedv4MountOptions)
	require.Equal(t, len(spec.Sharedv4MountOptions.Options), 2)
	val, ok := spec.Sharedv4MountOptions.Options["k1"]
	require.True(t, ok)
	require.Equal(t, val, "")
	val, ok = spec.Sharedv4MountOptions.Options["k2"]
	require.True(t, ok)
	require.Equal(t, val, "v2")
	_, ok = spec.Sharedv4MountOptions.Options["k3"]
	require.False(t, ok)
}

func TestOptFastpath(t *testing.T) {
	testSpecOptString(t, api.SpecFastpath, "true")

	spec := testSpecFromString(t, api.SpecFastpath, "true")
	require.True(t, spec.FpPreference, "Failed to parse fastpath option into spec")

	spec = testSpecFromString(t, api.SpecFastpath, "false")
	require.False(t, spec.FpPreference, "Failed to parse faspath option into spec")
}
