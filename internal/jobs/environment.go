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
// The name -> row resolution itself (including what an empty name means and
// when a missing row is tolerated) lives in environments.Resolve, which is also
// what container_runtime's resolver uses - one source of truth for "what does
// this environment name mean" across both.
func resolveEnvironmentSuffix(ctx context.Context, provider EnvironmentsProvider, name string) (string, error) {
	var envStorage storage.EnvironmentsStorage

	if provider != nil {
		envStorage = provider.Environments()
	}

	env, err := environments.Resolve(ctx, envStorage, name)
	if err != nil {
		return "", rerrors.Wrap(err, "error resolving environment")
	}

	return env.Suffix, nil
}
