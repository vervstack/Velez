package environments

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
)

// Resolve maps an environment NAME (as carried on the wire by
// CreateSmerd.Request.environment and friends) onto its stored row - the
// single place that owns "what does this environment name actually mean".
//
// An empty name means "the default environment" (DefaultEnvironmentName), NOT
// "no environment": Velez must keep behaving like a single-environment node for
// every caller that predates environments - the RPC layer no longer rejects an
// empty environment, and internally built requests (node-level sidecars,
// config-fetcher containers) have never carried one. The default environment's
// suffix is whatever ContainerSuffix this node was configured with before
// environments existed, so the resulting container names/labels are
// byte-for-byte what they used to be.
//
// Resolution of the DEFAULT environment is best-effort: with no environments
// storage yet (boot-time paths, single-node before the storage swap) or a
// missing row it degrades to the zero Environment - empty suffix, empty docker
// host - exactly what an unconfigured ContainerSuffix produced before. An
// EXPLICIT name is still strict: an unknown environment is an error, never a
// silent fallback onto another environment's containers.
//
//nolint:forbidigo // package-private sentinel, not shared/user-facing
var errEnvironmentsStorageUnavailable = rerrors.New("environments storage is not available")

func Resolve(ctx context.Context, envStorage storage.EnvironmentsStorage, name string) (domain.Environment, error) {
	isDefault := name == ""
	if isDefault {
		name = DefaultEnvironmentName
	}

	if envStorage == nil {
		if isDefault {
			return domain.Environment{}, nil
		}

		return domain.Environment{}, errEnvironmentsStorageUnavailable
	}

	env, err := envStorage.GetEnvironmentByName(ctx, name)
	if err != nil {
		if isDefault {
			return domain.Environment{}, nil
		}

		return domain.Environment{}, rerrors.Wrapf(err, "unknown environment '%s'", name)
	}

	return env, nil
}
