package jobs

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/environments"
)

// EnvironmentsProvider yields the currently-live environments storage.
//
// It's the *container* rather than a snapshot on purpose: single-node Velez
// starts on local_storage and swaps to postgres once cluster mode is enabled
// (cluster_clients.ClusterStateManagerContainer.Set), and a handler
// constructed at boot must observe that swap.
// cluster_clients.ClusterStateManagerContainer satisfies this.
type EnvironmentsProvider interface {
	Environments() storage.EnvironmentsStorage
}

// resolveEnvironmentSuffix maps an environment NAME (as carried on the wire by
// CreateSmerd.Request.environment and friends) onto the SUFFIX used for Docker
// naming/labels.
//
// An empty name means "the default environment"
// (environments.DefaultEnvironmentName), NOT "no environment": Velez must keep
// behaving like a single-environment node for every caller that predates
// environments - the RPC layer no longer rejects an empty environment, and
// internally built requests (node-level sidecars, config-fetcher containers)
// have never carried one. The default environment's suffix is whatever
// ContainerSuffix this node was configured with before environments existed,
// so the resulting container names/labels are byte-for-byte what they used to
// be.
//
// Resolution of the DEFAULT environment is best-effort: if there's no
// environments storage yet (boot-time paths, single-node before the storage
// swap) or the row is missing, it degrades to the empty suffix - exactly what
// an unconfigured ContainerSuffix produced before. An EXPLICIT name is still
// strict: an unknown environment is an error, never a silent fallback onto
// another environment's containers.
func resolveEnvironmentSuffix(ctx context.Context, provider EnvironmentsProvider, name string) (string, error) {
	isDefault := name == ""
	if isDefault {
		name = environments.DefaultEnvironmentName
	}

	if provider == nil {
		if isDefault {
			return "", nil
		}

		return "", rerrors.New("environments storage is not available")
	}

	envStorage := provider.Environments()
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
