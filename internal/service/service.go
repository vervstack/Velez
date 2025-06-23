package service

import (
	"context"

	"go.redsock.ru/evon"
	"go.vervstack.ru/matreshka/pkg/matreshka"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

type Services interface {
	SmerdManager() ContainerService
	ConfigurationService() ConfigurationService
}

type ContainerService interface {
	ListSmerds(ctx context.Context, req *velez_api.ListSmerds_Request) (*velez_api.ListSmerds_Response, error)
	DropSmerds(ctx context.Context, req *velez_api.DropSmerd_Request) (*velez_api.DropSmerd_Response, error)
	InspectSmerd(ctx context.Context, contId string) (*velez_api.Smerd, error)
}

type ConfigurationService interface {
	GetVervFromApi(ctx context.Context, meta domain.ConfigMeta) (matreshka.AppConfig, error)
	GetEnvFromApi(ctx context.Context, meta domain.ConfigMeta) (*evon.Node, error)
	UpdateConfig(ctx context.Context, config domain.AppConfig) error
	GetPlainFromApi(ctx context.Context, meta domain.ConfigMeta) ([]byte, error)

	SubscribeOnChanges(serviceNames ...string) error
	UnsubscribeFromChanges(serviceNames ...string) error
	GetUpdates() <-chan domain.ConfigurationPatch
}
