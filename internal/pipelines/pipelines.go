package pipelines

import (
	"context"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

type Pipeliner interface {
	// LaunchSmerd - prepares environment and launches container (smerd)
	// if possible (no other container with given name already exists dead or alive)
	LaunchSmerd(request domain.LaunchSmerd) Runner[domain.LaunchSmerdResult]
	// AssembleConfig - gathers information from
	// configuration file inside image
	// and matreshka instance
	// providing updated configuration
	AssembleConfig(request domain.AssembleConfig) Runner[domain.AppConfig]

	// UpgradeSmerd - upgrades already running smerd's
	// image to a new version (latest or specified)
	UpgradeSmerd(req domain.UpgradeSmerd) Runner[any]

	// EnableVervService - enables verv built-in services like
	// Matreshka, makosh, VPN with headscale and etc
	// TODO implement
	EnableVervService(req velez_api.VervServiceType) Runner[any]

	// ConnectServiceToVpn - connects any user service to cluster vpn
	ConnectServiceToVpn(vpn domain.ConnectServiceToVpn) Runner[any]

	CopyToVolume(req domain.CopyToVolumeRequest) Runner[any]
}

type Runner[T any] interface {
	Run(ctx context.Context) error
	Result() (*T, error)
}

type pipeliner struct {
	nodeClients    node_clients.NodeClients
	clusterClients cluster_clients.ClusterClients
	services       service.Services
}

func NewPipeliner(nodeClients node_clients.NodeClients,
	clusterClients cluster_clients.ClusterClients,
	services service.Services) Pipeliner {
	return &pipeliner{
		nodeClients:    nodeClients,
		clusterClients: clusterClients,

		services: services,
	}
}
