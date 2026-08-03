package container_manager

import (
	"context"

	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/environments"
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
	dockerWrapper node_clients.Docker
	dockerAPI     client.APIClient

	environments EnvironmentsProvider
}

func New(
	internalClients node_clients.NodeClients,
	environments EnvironmentsProvider,
) *ContainerManager {
	return &ContainerManager{
		dockerAPI: internalClients.Docker().Client(),

		dockerWrapper: internalClients.Docker(),
		environments:  environments,
	}
}

// resolveSuffix maps an environment name onto its Docker suffix.
//
// An empty name means "the default environment"
// (environments.DefaultEnvironmentName) rather than "no environment at all":
// the RPC entry points no longer reject an empty environment, and internal,
// node-wide callers (the ShutDownOnExit smerd dropper, GetServiceEnvironments,
// ...) never set one. The default environment's suffix is this node's
// pre-environments ContainerSuffix, so those callers keep seeing exactly the
// containers they saw before environments existed.
//
// Resolving the DEFAULT environment is best-effort - no storage yet or a
// missing row degrades to the empty suffix (= "don't scope by environment",
// which is what an unconfigured ContainerSuffix always produced). An EXPLICIT
// name stays strict: unknown environments are an error.
func (c *ContainerManager) resolveSuffix(ctx context.Context, name string) (string, error) {
	isDefault := name == ""
	if isDefault {
		name = environments.DefaultEnvironmentName
	}

	if c.environments == nil {
		if isDefault {
			return "", nil
		}

		return "", rerrors.New("environments storage is not available")
	}

	envStorage := c.environments.Environments()
	if envStorage == nil {
		if isDefault {
			return "", nil
		}

		return "", rerrors.New("environments storage is not available")
	}

	env, err := envStorage.GetEnvironmentByName(ctx, name)
	if err != nil {
		if isDefault {
			return "", nil
		}

		return "", rerrors.Wrapf(err, "unknown environment '%s'", name)
	}

	return env.Suffix, nil
}
