package jobs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	protobuf "google.golang.org/protobuf/proto"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func Test_HaproxyBackendTaskJsonRoundTrip(t *testing.T) {
	t.Parallel()

	original := &velez_api.CreateHaproxyInstanceTaskPayload{
		Request: &velez_api.CreateHaproxyInstance_Request{
			Name: "proxy",
			Backends: []*velez_api.PgBackend{
				{Name: "primary", Source: &velez_api.PgBackend_VelezService{VelezService: "pgaas_primary"}},
				{
					Name: "follower",
					Source: &velez_api.PgBackend_Endpoint{
						Endpoint: &velez_api.HaproxyConnectionTarget_Endpoint{Host: "pg.example.org", Port: 5432},
					},
					User: "reader", Password: "example", Role: velez_api.PgBackend_BACKUP,
				},
			},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	restored := &velez_api.CreateHaproxyInstanceTaskPayload{}

	err = json.Unmarshal(data, restored)
	require.NoError(t, err)
	require.True(t, protobuf.Equal(original, restored))
}

func Test_RenderHaproxyBackends_RendersAllRolesAndEmptyConfig(t *testing.T) {
	t.Parallel()

	backends := []haproxyBackend{
		{Name: "primary", Address: "pg1:5432"},
		{Name: "standby", Address: "[::1]:5433", IsBackup: true},
		{
			Name: "balanced", Address: "pg3:5432", IsBalanced: true,
			HasCredentials: true, User: "reader", Password: "example\n# two lines",
		},
	}

	config, err := renderHaproxyBackends(5432, "", backends)
	require.NoError(t, err)
	require.Contains(t, string(config), "balance roundrobin")
	require.Contains(t, string(config), "server primary pg1:5432 check\n")
	require.Contains(t, string(config), "server standby [::1]:5433 check backup\n")
	require.Contains(t, string(config), "server balanced pg3:5432 check\n")
	require.Contains(t, string(config), `user="reader" password="example\n# two lines"`)

	emptyRequest := &velez_api.CreateHaproxyInstance_Request{Name: "empty"}
	job := &deployHaproxyJob{request: emptyRequest}

	empty, _, _, err := job.prepareConfig(t.Context(), 5432)
	require.NoError(t, err)
	require.Contains(t, string(empty), "bind *:5432\n")
	require.NotContains(t, string(empty), "\n    server ")
}
