package units

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func testParse(t *testing.T, suffix string, b int64, bi int64) {
	n, err := Parse("10" + suffix)
	require.NoError(t, err, "Parse")
	require.Equal(t, int64(10*bi), n, "Parse")

	n, err = Parse("10" + " " + strings.ToLower(suffix))
	require.NoError(t, err, "Parse")
	require.Equal(t, int64(10*bi), n, "Parse")

	n, err = Parse("10" + " " + strings.ToUpper(suffix))
	require.NoError(t, err, "Parse")
	require.Equal(t, int64(10*bi), n, "Parse")

	_, err = Parse("1  0" + suffix)
	require.Error(t, err, "Parse")

	_, err = Parse("10" + suffix + "z")
	require.Error(t, err, "Parse")

	if len(suffix) == 0 {
		return
	}

	/*
		// No support for Mega
		n, err = Parse("10" + strings.ToLower(suffix) + "b")
		require.NoError(t, err, "Parse")
		require.Equal(t, int64(10*b), n, "Parse")
		n, err = Parse("10" + strings.ToUpper(suffix) + "B")
		require.NoError(t, err, "Parse")
		require.Equal(t, int64(10*b), n, "Parse")

		n, err = Parse("10" + " " + strings.ToLower(suffix) + "b")
		require.NoError(t, err, "Parse")
		require.Equal(t, int64(10*b), n, "Parse")

		n, err = Parse("10" + " " + strings.ToUpper(suffix) + "B")
		require.NoError(t, err, "Parse")
		require.Equal(t, int64(10*b), n, "Parse")
	*/
}

func TestParse(t *testing.T) {
	testParse(t, "k", 1000, 1024)
	testParse(t, "m", 1000*1000, 1024*1024)
	testParse(t, "", 1000, 1024*1024*1024)
	testParse(t, "g", 1000*1000*1000, 1024*1024*1024)
	testParse(t, "t", 1000*1000*1000*1000, 1024*1024*1024*1024)
	testParse(t, "p", 1000*1000*1000*1000*1000, 1024*1024*1024*1024*1024)
}

func testString(t *testing.T, quantity uint64, expected string) {
	s := String(quantity)
	require.Equal(t, s, expected)
}
func TestString(t *testing.T) {
	testString(t, 1, "1 bytes")
	testString(t, uint64(KiB), "1 KiB")
	testString(t, uint64(MiB), "1 MiB")
	testString(t, uint64(GiB), "1.0 GiB")
	testString(t, uint64(TiB), "1.00 TiB")
	testString(t, uint64(PiB), "1.00 PiB")
}
