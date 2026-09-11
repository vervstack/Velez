package state

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPgName_EmptySuffix_ReturnsBaseName(t *testing.T) {
	got := PgName("")

	require.Equal(t, "verv-cluster-state", got)
}

func TestPgName_WithSuffix_AppendsSuffix(t *testing.T) {
	got := PgName("e2e")

	require.Equal(t, "verv-cluster-state-e2e", got)
}
