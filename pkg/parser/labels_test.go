package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLabelMembership(t *testing.T) {
	set := map[string]string{"a": "1", "b": "2", "c": "3", "d": "4"}
	subset := map[string]string{"b": "2", "c": "3"}
	someKeys := map[string]string{"b": "2", "d": "3"}
	subsetKey := map[string]string{"b": "bad", "c": "3"}
	badSubset := map[string]string{"e": "2", "f": "3"}

	require.True(t, HasLabels(set, subset))
	require.False(t, HasLabels(set, subsetKey))
	require.False(t, HasLabels(set, badSubset))
	require.False(t, HasLabels(set, someKeys))

	require.True(t, HasLabelKeys(set, subset))
	require.True(t, HasLabelKeys(set, subsetKey))
	require.False(t, HasLabelKeys(set, badSubset))

	require.True(t, HasAnyLabelKey(set, someKeys))
	require.False(t, HasAnyLabelKey(set, badSubset))
}
