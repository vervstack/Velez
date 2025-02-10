package service

import (
	"context"

	"go.redsock.ru/evon"
	"go.vervstack.ru/matreshka"

	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/pkg/velez_api"
)

type Services interface {
	SmerdManagers
	ConfigurationService
}

type SmerdManagers interface {
	ListSmerds(ctx context.Context, req *velez_api.ListSmerds_Request) (*velez_api.ListSmerds_Response, error)
	DropSmerds(ctx context.Context, req *velez_api.DropSmerd_Request) (*velez_api.DropSmerd_Response, error)
	FetchConfig(ctx context.Context, req *velez_api.AssembleConfig_Request) (*matreshka.AppConfig, error)
	InspectSmerd(ctx context.Context, contId string) (*velez_api.Smerd, error)
}

type ConfigurationService interface {
	GetFromContainer(ctx context.Context, contId string) (matreshka.AppConfig, error)
	GetFromApi(ctx context.Context, serviceName string) (matreshka.AppConfig, error)
	GetEnvFromApi(ctx context.Context, serviceName string) ([]*evon.Node, error)
	UpdateConfig(ctx context.Context, serviceName string, config matreshka.AppConfig) error

	SubscribeOnChanges(serviceNames ...string) error
	UnsubscribeFromChanges(serviceNames ...string) error
	GetUpdates() <-chan domain.ConfigurationPatch
}
