package clients

import (
	"context"

	"go.vervstack.ru/Velez/internal/domain"
)

type ClusterClients interface {
	Configurator() Configurator
	Vpn() VervPrivateNetworkClient
}

type clusterClientsContainer struct {
	matreshka Configurator
	headscale VervPrivateNetworkClient
}

func NewClusterClientsContainer(
	cfg Configurator,
	headscale VervPrivateNetworkClient,
) ClusterClients {
	return &clusterClientsContainer{
		matreshka: cfg,
		headscale: headscale,
	}
}

func (c *clusterClientsContainer) Configurator() Configurator {
	return c.matreshka
}

func (c *clusterClientsContainer) Vpn() VervPrivateNetworkClient {
	return c.headscale
}

type VervPrivateNetworkClient interface {
	CreateNamespace(ctx context.Context, name string) (domain.VpnNamespace, error)
	ListNamespaces(ctx context.Context) ([]domain.VpnNamespace, error)
	DeleteNamespace(ctx context.Context, id string) error

	IssueClientKey(ctx context.Context, namespace string) (string, error)
}
