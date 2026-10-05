package service_discovery

import (
	"context"
	"time"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/makosh/pkg/makosh_be"
	matreshkaSd "go.vervstack.ru/matreshka/pkg/matreshka/service_discovery"
	"google.golang.org/grpc/resolver"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
)

const (
	vervScheme = "verv"

	resolveTimeout = 5 * time.Second
)

// NewVervResolverBuilder resolves verv://<service> targets without makosh:
// a configured override wins, otherwise the node's own service discovery is
// asked. It is handed to a single client via grpc.WithResolvers instead of
// being registered globally, so every app in a process resolves against its
// own overrides and discovery.
func NewVervResolverBuilder(
	overrides matreshkaSd.Overrides,
	discovery cluster_clients.ServiceDiscovery,
) resolver.Builder {
	b := &vervResolverBuilder{
		overrides: make(map[string][]string, len(overrides)),
		discovery: discovery,
	}

	for _, o := range overrides {
		b.overrides[o.ServiceName] = o.Urls
	}

	return b
}

type vervResolverBuilder struct {
	overrides map[string][]string
	discovery cluster_clients.ServiceDiscovery
}

func (b *vervResolverBuilder) Scheme() string {
	return vervScheme
}

func (b *vervResolverBuilder) Build(
	target resolver.Target,
	cc resolver.ClientConn,
	_ resolver.BuildOptions,
) (resolver.Resolver, error) {
	r := &vervResolver{
		serviceName: target.URL.Host,
		builder:     b,
		cc:          cc,
	}

	r.ResolveNow(resolver.ResolveNowOptions{})

	return r, nil
}

type vervResolver struct {
	serviceName string
	builder     *vervResolverBuilder
	cc          resolver.ClientConn
}

func (r *vervResolver) ResolveNow(resolver.ResolveNowOptions) {
	ctx, cancel := context.WithTimeout(context.Background(), resolveTimeout)
	defer cancel()

	urls, err := r.builder.lookup(ctx, r.serviceName)
	if err != nil {
		r.cc.ReportError(err)

		return
	}

	var state resolver.State

	state.Addresses = make([]resolver.Address, 0, len(urls))

	for _, u := range urls {
		state.Addresses = append(state.Addresses, resolver.Address{Addr: u})
	}

	err = r.cc.UpdateState(state)
	if err != nil {
		r.cc.ReportError(err)
	}
}

func (r *vervResolver) Close() {}

func (b *vervResolverBuilder) lookup(ctx context.Context, serviceName string) ([]string, error) {
	urls := b.overrides[serviceName]
	if len(urls) > 0 {
		return urls, nil
	}

	req := &makosh_be.ListEndpoints_Request{ServiceName: serviceName}

	resp, err := b.discovery.ListEndpoints(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing endpoints from service discovery")
	}

	return resp.GetUrls(), nil
}
