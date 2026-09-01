package service

import (
	"context"

	"go.redsock.ru/evon"
	"go.vervstack.ru/matreshka/pkg/matreshka"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
)

type Services interface {
	SmerdManager() ContainerService
	ConfigurationService() ConfigurationService

	Docker() node_clients.Docker

	VervServices() VervServicesService
	NodeService() NodeService
	PluginService() PluginService
	StorageContainer() *storage.Container
}

type ContainerService interface {
	ListSmerds(ctx context.Context, req *velez_api.ListSmerds_Request) (*velez_api.ListSmerds_Response, error)
	DropSmerds(ctx context.Context, req *velez_api.DropSmerd_Request) (*velez_api.DropSmerd_Response, error)
	InspectSmerd(ctx context.Context, environment, contID string) (*velez_api.Smerd, error)

	ConnectToNetwork(ctx context.Context, req domain.Connection) error
	DisconnectFromNetwork(ctx context.Context, req domain.Connection) error
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

// VervServicesService provides API for operations over Verv Services.
// TODO
//
//nolint:interfacebloat
type VervServicesService interface {
	Get(ctx context.Context, r domain.GetServiceReq) (domain.Service, error)
	CreateNewDeploy(ctx context.Context, request domain.CreateDeployReq) error
	UpgradeDeploy(ctx context.Context, request domain.UpgradeDeployReq) error
	List(ctx context.Context, req domain.ListServicesReq) (domain.ServiceList, error)
	ListDeployments(ctx context.Context, req domain.ListDeploymentsReq) (domain.DeploymentList, error)
	StopService(ctx context.Context, name, environment string) error
	RestartService(ctx context.Context, name, environment string) error
	Remove(ctx context.Context, req domain.RemoveServiceReq) error

	GetServiceMetrics(ctx context.Context, serviceName, environment string) (domain.ServiceMetrics, error)
	GetServiceResources(ctx context.Context, serviceName string) ([]domain.BoundResource, error)
	GetServiceGraph(ctx context.Context, serviceName string) (domain.ServiceGraph, error)
	GetServiceEnvironments(ctx context.Context, serviceName string) ([]domain.ServiceEnvironment, error)

	// Environment management (velez.environments). NOTE: domain.Environment is
	// a deployment environment - unrelated to domain.ServiceEnvironment above,
	// which is a per-service dashboard status projection.
	ListEnvironments(ctx context.Context) ([]domain.Environment, error)
	GetEnvironment(ctx context.Context, name string) (domain.Environment, error)
	CreateEnvironment(ctx context.Context, req domain.CreateEnvironmentReq) (domain.Environment, error)
	UpdateEnvironment(ctx context.Context, req domain.UpdateEnvironmentReq) (domain.Environment, error)
	DeleteEnvironment(ctx context.Context, req domain.DeleteEnvironmentReq) error

	// ResolveEnvironmentSuffix maps an environment name (what callers pass on
	// the wire) to the suffix used for Docker naming/labels. Also the
	// validation entry point for the `environment` request field.
	ResolveEnvironmentSuffix(ctx context.Context, name string) (string, error)

	// Registry management (velez.registries) - container image registries
	// (Docker Hub, generic v2) used by SearchImages.
	ListRegistries(ctx context.Context) ([]domain.Registry, error)
	GetRegistry(ctx context.Context, id int64) (domain.Registry, error)
	CreateRegistry(ctx context.Context, req domain.CreateRegistryReq) (domain.Registry, error)
	UpdateRegistry(ctx context.Context, req domain.UpdateRegistryReq) (domain.Registry, error)
	DeleteRegistry(ctx context.Context, req domain.DeleteRegistryReq) error
}

type NodeService interface {
	ListNodes(ctx context.Context, req domain.ListNodesReq) (domain.NodesList, error)
}

type PluginService interface {
	ListPlugins(ctx context.Context) (*velez_api.ListPlugins_Response, error)
}
