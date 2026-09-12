package container_runtime

import (
	"context"
	"sync"

	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox/closer"

	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/environments"
)

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
// label-based runtime per environment suffix - except for an environment
// bound to a Docker host other than the node's own, which gets its own
// dedicated connection instead (see dedicatedRuntime).
type resolver struct {
	cli         client.APIClient
	nodeHost    string
	bakedLabels []string
	envProvider EnvironmentsProvider

	dedicatedMu  sync.Mutex
	dedicatedCli map[string]client.APIClient
}

// NewResolver builds the resolver used in production wiring. cli is the node's
// shared Docker connection (node_clients.Docker().Client()), nodeHost that same
// connection's resolved address (node_clients.Docker().Host()) - the value an
// environment's DockerHost must equal to be served off cli rather than a
// dedicated connection - and bakedLabels the node's configured CustomLabels.
func NewResolver(
	cli client.APIClient,
	nodeHost string,
	bakedLabels []string,
	envProvider EnvironmentsProvider,
) RuntimeResolver {
	return &resolver{
		cli:          cli,
		nodeHost:     nodeHost,
		bakedLabels:  bakedLabels,
		envProvider:  envProvider,
		dedicatedCli: make(map[string]client.APIClient),
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

	if env.DockerHost != "" && env.DockerHost != r.nodeHost {
		dedicatedCli, dedicatedErr := r.dedicatedClient(env.DockerHost)
		if dedicatedErr != nil {
			return nil, rerrors.Wrapf(dedicatedErr,
				"error connecting to dedicated docker host '%s' for environment '%s'", env.DockerHost, env.Name)
		}

		return newDirectRuntime(dedicatedCli, r.bakedLabels), nil
	}

	runtime := newLabelBasedRuntime(r.cli, env.Suffix, r.bakedLabels)

	return runtime, nil
}

// dedicatedClient returns the cached client.APIClient for dockerHost,
// building and caching one on first use. Cached rather than reconnected on
// every call - a dedicated engine is a real TCP connection, unlike the
// resolver's own cli which is handed in already built - and registered with
// closer.Add the same way node_clients/docker.NewClientWithOpts registers the
// node's shared connection, so it's closed on process shutdown.
func (r *resolver) dedicatedClient(dockerHost string) (client.APIClient, error) {
	r.dedicatedMu.Lock()
	defer r.dedicatedMu.Unlock()

	cached, ok := r.dedicatedCli[dockerHost]
	if ok {
		return cached, nil
	}

	cli, err := client.NewClientWithOpts(client.WithHost(dockerHost), client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, rerrors.Wrap(err, "error building dedicated docker client")
	}

	closer.Add(cli.Close)

	r.dedicatedCli[dockerHost] = cli

	return cli, nil
}
