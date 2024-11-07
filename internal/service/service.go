package service

import (
	"context"

	"github.com/Red-Sock/evon"
	"github.com/godverv/matreshka"

	"github.com/godverv/Velez/pkg/velez_api"
)

type Services interface {
	SmerdLauncher
	OtherManagers
}

type OtherManagers interface {
	ListSmerds(ctx context.Context, req *velez_api.ListSmerds_Request) (*velez_api.ListSmerds_Response, error)
	DropSmerds(ctx context.Context, req *velez_api.DropSmerd_Request) (*velez_api.DropSmerd_Response, error)
	FetchConfig(ctx context.Context, req *velez_api.FetchConfig_Request) (*matreshka.AppConfig, error)
	InspectSmerd(ctx context.Context, contId string) (*velez_api.Smerd, error)
}

type SmerdLauncher interface {
	LaunchSmerd(ctx context.Context, req *velez_api.CreateSmerd_Request) (id string, err error)
}

type ConfigurationService interface {
	GetFromContainer(ctx context.Context, contId string) (matreshka.AppConfig, error)
	GetFromApi(ctx context.Context, serviceName string) (matreshka.AppConfig, error)
	GetEnvFromApi(ctx context.Context, serviceName string) ([]*evon.Node, error)
	UpdateConfig(ctx context.Context, serviceName string, config matreshka.AppConfig) error
}
