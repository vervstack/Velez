package address_book

import (
	"context"
	"errors"
	"maps"
	"sync"

	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/user_errors"
)

var _ service.AddressBook = (*Book)(nil)

type Book struct {
	runtimes    container_runtime.RuntimeResolver
	dataStorage storage.Storage
	docker      node_clients.Docker
	vcn         cluster_clients.VervClosedNetworkClient
	registry    cluster_clients.AddressRegistry

	progress *rebuildProgress

	mu    sync.Mutex
	known map[string]struct{}
}

func New(
	runtimes container_runtime.RuntimeResolver,
	dataStorage storage.Storage,
	docker node_clients.Docker,
	vcn cluster_clients.VervClosedNetworkClient,
	registry cluster_clients.AddressRegistry,
) *Book {
	return &Book{
		runtimes:    runtimes,
		dataStorage: dataStorage,
		docker:      docker,
		vcn:         vcn,
		registry:    registry,
		progress:    newRebuildProgress(),
		known:       make(map[string]struct{}),
	}
}

func (b *Book) Rebuild(ctx context.Context) error {
	err := b.progress.start()
	if err != nil {
		return rerrors.Wrap(err, "error starting addresses rebuild")
	}

	return b.runRebuild(ctx)
}

func (b *Book) RebuildAsync(ctx context.Context) error {
	err := b.progress.start()
	if err != nil {
		return rerrors.Wrap(err, "error starting addresses rebuild")
	}

	go b.runRebuildInBackground(context.WithoutCancel(ctx))

	return nil
}

func (b *Book) RebuildStatus() domain.AddressRebuildStatus {
	return b.progress.snapshot()
}

func (b *Book) Sync(ctx context.Context, serviceName string) error {
	found, err := b.crawl(ctx, nil)
	if err != nil {
		return rerrors.Wrap(err, "error crawling addresses")
	}

	addresses, isFound := found[serviceName]
	if isFound {
		b.remember(serviceName)
	} else {
		b.forget(serviceName)
	}

	err = b.store(ctx, serviceName, addresses)
	if err != nil {
		return rerrors.Wrap(err, "error storing addresses")
	}

	return nil
}

func (b *Book) runRebuildInBackground(ctx context.Context) {
	err := b.runRebuild(ctx)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("error rebuilding addresses")
	}
}

func (b *Book) runRebuild(ctx context.Context) (err error) {
	defer func() {
		recovered := recover()
		if recovered != nil {
			log.Ctx(ctx).Error().Interface("panic", recovered).Msg("addresses rebuild panicked")

			err = rerrors.Wrap(errRebuildPanicked)
		}

		b.progress.finish(err)
	}()

	found, err := b.crawl(ctx, b.progress)
	if err != nil {
		return rerrors.Wrap(err, "error crawling addresses")
	}

	stale := b.replaceKnown(found)
	b.progress.addSteps(len(found) + len(stale))

	var rebuildErrs []error

	for serviceName, addresses := range found {
		err = b.store(ctx, serviceName, addresses)
		if err != nil {
			rebuildErrs = append(rebuildErrs, err)
		}

		b.progress.step()
	}

	for _, serviceName := range stale {
		err = b.store(ctx, serviceName, nil)
		if err != nil {
			rebuildErrs = append(rebuildErrs, err)
		}

		b.progress.step()
	}

	err = errors.Join(rebuildErrs...)
	if err != nil {
		return rerrors.Wrap(err, "error storing addresses")
	}

	return nil
}

func (b *Book) crawl(ctx context.Context, progress *rebuildProgress) (map[string]serviceAddresses, error) {
	environments, err := b.dataStorage.Environments().ListEnvironments(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing environments")
	}

	progress.addSteps(len(environments) + 1)

	discovered := make(map[string]webUiDiscovery)

	for _, env := range environments {
		found, discoverErr := b.discoverEnvironment(ctx, env)

		progress.step()

		if discoverErr != nil {
			log.Ctx(ctx).Warn().
				Str("environment", env.Name).
				Err(discoverErr).
				Msg("error discovering environment web uis")

			continue
		}

		maps.Copy(discovered, found)
	}

	nodes := b.vcnNodes(ctx)

	progress.step()

	result := make(map[string]serviceAddresses, len(discovered))
	for serviceName, found := range discovered {
		result[serviceName] = buildAddresses(serviceName, found, nodes)
	}

	return result, nil
}

func (b *Book) discoverEnvironment(
	ctx context.Context,
	env domain.Environment,
) (map[string]webUiDiscovery, error) {
	runtime, err := b.runtimes.Runtime(ctx, env.Name)
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving environment runtime")
	}

	listReq := &velez_api.ListSmerds_Request{Environment: env.Name}

	containers, err := runtime.ListContainers(ctx, listReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing environment containers")
	}

	host := env.PublishedHost(b.docker.Host())

	return discoverWebUis(containers, host), nil
}

func (b *Book) vcnNodes(ctx context.Context) []domain.VcnNode {
	nodes, err := b.vcn.ListNodes(ctx)
	if errors.Is(err, user_errors.ErrServiceIsDisabled) {
		log.Ctx(ctx).Debug().Msg("vcn is disabled, skipping vcn addresses")

		return nil
	}

	if err != nil {
		log.Ctx(ctx).Warn().
			Err(err).
			Msg("error listing vcn nodes, skipping vcn addresses")

		return nil
	}

	return nodes
}

func (b *Book) remember(serviceName string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.known[serviceName] = struct{}{}
}

func (b *Book) forget(serviceName string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.known, serviceName)
}

// replaceKnown stores the freshly written service names and returns the
// previously written ones that are gone now.
func (b *Book) replaceKnown(found map[string]serviceAddresses) []string {
	b.mu.Lock()
	defer b.mu.Unlock()

	var stale []string

	for serviceName := range b.known {
		_, isFound := found[serviceName]
		if !isFound {
			stale = append(stale, serviceName)
		}
	}

	b.known = make(map[string]struct{}, len(found))
	for serviceName := range found {
		b.known[serviceName] = struct{}{}
	}

	return stale
}
