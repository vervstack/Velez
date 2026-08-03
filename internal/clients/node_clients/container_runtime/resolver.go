package container_runtime

import (
	"context"

	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/environments"
)

// ErrDedicatedRuntimeNotImplemented is returned for an environment that asks
// for a Docker daemon of its own (domain.Environment.DockerHost set). Building
// that daemon - pkg/docker_setup, systemd units, connection pooling - is
// Phase 2 of docs/container_runtimes/roadmap.md; failing loudly beats silently
// serving such an environment off the shared daemon.
var ErrDedicatedRuntimeNotImplemented = rerrors.New("dedicated container runtime is not implemented yet")

// EnvironmentsProvider yields the currently-live environments storage.
//
// It's the *container* rather than a snapshot on purpose: single-node Velez
// starts on local_storage and swaps to postgres once cluster mode is enabled
// (cluster_clients.ClusterStateManagerContainer.Set), and a resolver
// constructed at boot must observe that swap.
// cluster_clients.ClusterStateManagerContainer satisfies this.
type EnvironmentsProvider interface {
	Environments() storage.EnvironmentsStorage
}

// resolver is the default RuntimeResolver: one shared Docker connection, one
// label-based runtime per environment suffix.
type resolver struct {
	cli         client.APIClient
	bakedLabels []string
	envProvider EnvironmentsProvider
}

// NewResolver builds the resolver used in production wiring. cli is the node's
// shared Docker connection (node_clients.Docker().Client()) and bakedLabels the
// node's configured CustomLabels.
func NewResolver(
	cli client.APIClient,
	bakedLabels []string,
	envProvider EnvironmentsProvider,
) RuntimeResolver {
	return &resolver{
		cli:         cli,
		bakedLabels: bakedLabels,
		envProvider: envProvider,
	}
}

// Runtime re-reads the environment's row on every call - see RuntimeResolver.
func (r *resolver) Runtime(ctx context.Context, environment string) (ContainerRuntime, error) {
	var envStorage storage.EnvironmentsStorage

	if r.envProvider != nil {
		envStorage = r.envProvider.Environments()
	}

	env, err := environments.Resolve(ctx, envStorage, environment)
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving environment")
	}

	if env.DockerHost != "" {
		return nil, rerrors.Wrapf(ErrDedicatedRuntimeNotImplemented,
			"environment '%s' is bound to docker host '%s'", env.Name, env.DockerHost)
	}

	common := commonRuntime{
		cli: r.cli,
	}

	runtime := &labelBasedRuntime{
		commonRuntime: common,
		suffix:        env.Suffix,
		bakedLabels:   r.bakedLabels,
	}

	return runtime, nil
}
