package container_manager

import (
	"github.com/docker/docker/client"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/storage"
)

// labelTrue is the value Velez writes for its boolean container labels.
const (
	labelTrue = "true"
)

// EnvironmentsProvider yields the currently-live environments storage. It's
// the container (not a snapshot) so a runtime swap from local_storage to
// postgres - cluster_clients.ClusterStateManagerContainer.Set - is picked up.
type EnvironmentsProvider interface {
	Environments() storage.EnvironmentsStorage
}

type ContainerManager struct {
	// dockerWrapper is now unused: DropSmerds (smerds_drop.go) was its last
	// remaining caller and has since migrated to the resolved
	// container_runtime.ContainerRuntime, same as ListSmerds/InspectSmerd
	// before it. Kept on the struct rather than removed per this repo's
	// backward-compatibility-first policy - flagged here rather than silently
	// deleted so a maintainer can decide whether to drop it in a follow-up.
	//nolint:unused // see comment above; intentionally kept, not deleted
	dockerWrapper node_clients.Docker
	dockerAPI     client.APIClient

	// runtimes resolves the ContainerRuntime serving a given environment - see
	// docs/container_runtimes. ListSmerds routes through it instead of calling
	// dockerWrapper.ListContainers directly (environment name -> suffix
	// resolution now lives in environments.Resolve, shared with
	// runtimes.Runtime, rather than being duplicated here).
	runtimes container_runtime.RuntimeResolver
}

func New(
	internalClients node_clients.NodeClients,
	runtimes container_runtime.RuntimeResolver,
) *ContainerManager {
	return &ContainerManager{
		dockerAPI: internalClients.Docker().Client(),

		dockerWrapper: internalClients.Docker(),
		runtimes:      runtimes,
	}
}
