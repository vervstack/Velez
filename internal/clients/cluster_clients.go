package clients

import (
	"context"

	"github.com/godverv/matreshka"
)

type ClusterClients interface {
	ServiceDiscovery() ServiceDiscovery
	Configurator() Configurator
}

type Configurator interface {
	GetFromContainer(ctx context.Context, contId string) (matreshka.AppConfig, error)
	GetFromApi(ctx context.Context, serviceName string) (matreshka.AppConfig, error)
	UpdateConfig(ctx context.Context, serviceName string, config matreshka.AppConfig) error
}

type ServiceDiscovery interface {
}
