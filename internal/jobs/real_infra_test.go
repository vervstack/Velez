package jobs

import (
	"testing"

	"github.com/docker/docker/client"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/ports"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/service_manager/container_manager"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/environments"
	"go.vervstack.ru/Velez/tests/test_helper"
)

// staticEnvironmentsProvider adapts a single storage.EnvironmentsStorage into
// container_runtime.EnvironmentsProvider. Mirrors
// container_runtime/resolver_test.go's unexported staticProvider (same
// shape) - that type can't be imported from here, so it's duplicated rather
// than reused.
type staticEnvironmentsProvider struct {
	envs storage.EnvironmentsStorage
}

func (p staticEnvironmentsProvider) Environments() storage.EnvironmentsStorage {
	return p.envs
}

// dockerOnlyNodeClients implements node_clients.NodeClients with a real
// Docker() and stubbed-out everything else. container_manager.New - the only
// consumer these real-infra fixtures need - only ever calls .Docker() on its
// NodeClients argument at construction time and never retains the reference
// (confirmed by reading
// internal/service/service_manager/container_manager/service.go's New), so
// the other four methods returning nil/zero values is safe here.
type dockerOnlyNodeClients struct {
	docker node_clients.Docker
}

func (n dockerOnlyNodeClients) Docker() node_clients.Docker {
	return n.docker
}

func (n dockerOnlyNodeClients) PortManager() node_clients.PortManager {
	return nil
}

func (n dockerOnlyNodeClients) PortManagerContainer() *ports.Container {
	return nil
}

func (n dockerOnlyNodeClients) LocalStateManager() node_clients.StateManager {
	return nil
}

func (n dockerOnlyNodeClients) HardwareManager() node_clients.HardwareManager {
	return nil
}

// newRealUpgradeFixture wires a fully real service.ContainerService and
// container_runtime.RuntimeResolver against the local Docker daemon, backed
// by a single-environment (envName) in-memory environments.NewStatic
// storage. Also returns the raw client.APIClient so tests can create
// fixtures through the resolved runtime and assert against real
// ContainerInspect results.
func newRealUpgradeFixture(
	t *testing.T, envName string,
) (service.ContainerService, container_runtime.RuntimeResolver, client.APIClient) {
	t.Helper()

	cli := test_helper.NewRealDockerAPI(t)
	realDocker := test_helper.NewRealDocker(t)

	envs := environments.NewStatic([]string{envName}, "")
	provider := staticEnvironmentsProvider{envs: envs}

	runtimes := container_runtime.NewResolver(cli, nil, provider)

	nodeClients := dockerOnlyNodeClients{docker: realDocker}

	containerService := container_manager.New(nodeClients, runtimes)

	return containerService, runtimes, cli
}
