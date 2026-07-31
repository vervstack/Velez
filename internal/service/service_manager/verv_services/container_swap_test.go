package verv_services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
)

// TestVervService_ReflectsStorageContainerSwap proves the fix for the bug
// where VervService cached each sub-storage getter's result (Services(),
// Deployments(), ServiceDependencies(), ServiceResources(), TxManager()) at
// construction time instead of resolving it fresh on every call (see
// service.go's New(), which now stores the storage.Storage/*storage.Container
// itself as a single dataStorage field).
//
// In production, internal/storage/plugin_storage.go's *storage.Container
// starts backed by local_storage (single-node/dev) and is swapped to the
// real Postgres-backed storage.Storage at runtime via Container.Set() when a
// node is converted to cluster mode (internal/jobs/enable_statefull.go's
// registerPluginJob, internal/app/custom.go's
// "c.Services.StorageContainer().Set(...)"). A VervService constructed
// against the container before that swap must still see post-swap data on
// every subsequent call - this test exercises the real *storage.Container
// (not a fake) to prove that end-to-end at the VervService level, without
// needing Docker/Postgres/e2e.
func TestVervService_ReflectsStorageContainerSwap(t *testing.T) {
	preSwapWant := domain.DeploymentList{Total: 0}
	postSwapWant := domain.DeploymentList{
		Total: 1,
		Deployments: []domain.Deployment{
			{Id: 1},
		},
	}

	preSwapDeployments := &testDeploymentsStorage{
		listDeploymentsFunc: func(_ context.Context, _ domain.ListDeploymentsReq) (domain.DeploymentList, error) {
			return preSwapWant, nil
		},
	}
	postSwapDeployments := &testDeploymentsStorage{
		listDeploymentsFunc: func(_ context.Context, _ domain.ListDeploymentsReq) (domain.DeploymentList, error) {
			return postSwapWant, nil
		},
	}

	preSwapBackend := &testStorage{deployments: preSwapDeployments}
	postSwapBackend := &testStorage{deployments: postSwapDeployments}

	container := storage.NewStorageContainer(preSwapBackend)

	service := &VervService{dataStorage: container}

	got, err := service.ListDeployments(context.Background(), domain.ListDeploymentsReq{})
	require.NoError(t, err)
	require.Equal(t, preSwapWant, got)

	container.Set(postSwapBackend)

	got, err = service.ListDeployments(context.Background(), domain.ListDeploymentsReq{})
	require.NoError(t, err)
	require.Equal(t, postSwapWant, got)
}
