package vcn

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/local_storage"
)

func TestExternalNetworkRequiresCluster(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}
	data := storage.NewStorageContainer(local_storage.New(nil, nil, cfg))
	svc := New(data, nil, nil, nil)
	req := service.CreateVcnNetworkRequest{
		Name:     "external",
		External: &service.VcnExternalSource{Url: "http://127.0.0.1", Token: "unused"},
	}

	_, err := svc.Create(t.Context(), req)
	require.ErrorIs(t, err, user_errors.ErrExternalNetworkRequiresCluster)
}

func TestDeployedNetworkStateAndKeys(t *testing.T) {
	t.Parallel()

	cfg := config.Config{}

	cfg.Environment.LocalStatePath = t.TempDir() + "/state.json"

	node, err := node_clients.NewNodeClients(t.Context(), cfg)
	require.NoError(t, err)

	local := local_storage.New(node.Docker(), node.LocalStateManager(), cfg)
	data := storage.NewStorageContainer(local)
	runtime := container_runtime.NewResolver(node.Docker().Client(), node.Docker().Host(), nil, local)
	svc := New(data, runtime, node, nil)

	config := net.ListenConfig{}
	listener, err := config.Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)
	require.NoError(t, listener.Close())

	name := "vcn-test-" + uuid.NewString()[:8]
	deploy := &domain.SetupHeadscaleRequest{ExposeToPort: &port}
	req := service.CreateVcnNetworkRequest{Name: name, Host: "localhost", Deploy: deploy}

	record, err := svc.Create(t.Context(), req)
	require.NoError(t, err)

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		require.NoError(t, svc.DropLocal(ctx, record.Id))
	})

	restarted := New(data, runtime, node, nil)
	networks, err := restarted.List(t.Context())
	require.NoError(t, err)
	require.Contains(t, networkIds(networks), record.Id)

	details, err := restarted.Get(t.Context(), record.Id)
	require.NoError(t, err)
	require.Equal(t, record.Url, details.Network.Url)
	require.Empty(t, details.Peers)

	login, key, err := restarted.Connection(t.Context(), record.Id)
	require.NoError(t, err)
	require.Equal(t, record.Url, login)
	require.NotEmpty(t, key)

	_, reused, err := restarted.Connection(t.Context(), record.Id)
	require.NoError(t, err)
	require.Equal(t, key, reused)

	issued, err := restarted.IssueApiKey(t.Context(), record.Id)
	require.NoError(t, err)
	require.NotEmpty(t, issued)

	details, err = restarted.Get(t.Context(), record.Id)
	require.NoError(t, err)
	require.Len(t, details.ApiKeys, 2)

	prefix, _, _ := strings.Cut(issued, ".")

	err = restarted.ExpireApiKey(t.Context(), record.Id, prefix)
	require.NoError(t, err)

	err = restarted.Drop(t.Context(), record.Id)
	require.NoError(t, err)

	_, err = restarted.Get(t.Context(), record.Id)
	require.ErrorIs(t, err, user_errors.ErrVcnNetworkNotFound)
}

func networkIds(records []domain.Network) []string {
	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.Id)
	}

	return ids
}
