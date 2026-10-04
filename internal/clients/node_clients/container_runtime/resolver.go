package container_runtime

import (
	"context"
	"sync"

	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox/closer"

	"go.vervstack.ru/Velez/internal/clients/node_clients/runtime_policy"
	"go.vervstack.ru/Velez/internal/domain"
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

// SettingsProvider yields the node-wide settings. Nil means disabled.
type SettingsProvider = runtime_policy.SettingsProvider

// SettingsStorageProvider yields the currently-live settings storage, for the
// same swap-awareness reason as EnvironmentsProvider.
type SettingsStorageProvider interface {
	Settings() storage.SettingsStorage
}

type liveSettingsProvider struct {
	provider SettingsStorageProvider
}

// NewSettingsProvider adapts a storage container to a SettingsProvider that
// re-resolves the live settings storage on every call.
func NewSettingsProvider(provider SettingsStorageProvider) SettingsProvider {
	return &liveSettingsProvider{provider: provider}
}

func (l *liveSettingsProvider) GetSettings(ctx context.Context) (domain.Settings, error) {
	settings, err := l.provider.Settings().GetSettings(ctx)
	if err != nil {
		return domain.Settings{}, rerrors.Wrap(err, "error getting settings")
	}

	return settings, nil
}

// ResolverOption customizes NewResolver.
type ResolverOption func(*resolver)

func WithSettings(provider SettingsProvider) ResolverOption {
	return func(r *resolver) {
		r.settings = provider
	}
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
	settings    SettingsProvider

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
	opts ...ResolverOption,
) RuntimeResolver {
	r := &resolver{
		cli:          cli,
		nodeHost:     nodeHost,
		bakedLabels:  bakedLabels,
		envProvider:  envProvider,
		dedicatedCli: make(map[string]client.APIClient),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
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

		directRuntime := newDirectRuntime(dedicatedCli, r.bakedLabels)

		directRuntime.settings = r.settings

		return directRuntime, nil
	}

	runtime := newLabelBasedRuntime(r.cli, env.Suffix, r.bakedLabels)

	runtime.settings = r.settings

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
