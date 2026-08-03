package velez_api_impl

import (
	"testing"

	"github.com/stretchr/testify/require"

	api "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func TestToConnection_MapsEnvironment(t *testing.T) {
	req := &api.Connection{Environment: "STAGE"}

	conn := toConnection(req)
	require.Equal(t, "STAGE", conn.Environment)
}
