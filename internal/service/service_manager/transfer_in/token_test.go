package transfer_in

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_NewToken_IsUnpredictableAndUrlSafe(t *testing.T) {
	first, err := newToken()
	require.NoError(t, err)

	second, err := newToken()
	require.NoError(t, err)

	require.NotEqual(t, first, second)
	require.Len(t, first, 43)
	require.NotContains(t, first, "=")
	require.NotContains(t, first, "+")
	require.NotContains(t, first, "/")
}

func Test_HashToken_Scenarios(t *testing.T) {
	require.Equal(t, hashToken("abc"), hashToken("abc"))
	require.NotEqual(t, hashToken("abc"), hashToken("abd"))
	require.Len(t, hashToken("abc"), 64)
	require.NotEqual(t, "abc", hashToken("abc"))
}

func Test_TokenExpiry_IsOneHourAhead(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)

	require.Equal(t, now.Add(time.Hour), tokenExpiry(now))
}
