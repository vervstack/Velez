package service

import (
	"context"

	"go.redsock.ru/evon"
	"go.vervstack.ru/matreshka/pkg/matreshka"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/storage"
)

//nolint:interfacebloat
type Services interface {
	SmerdManager() ContainerService
	ConfigurationService() ConfigurationService

	Docker() node_clients.Docker

	VervServices() VervServicesService
	NodeService() NodeService
	PluginService() PluginService
	StorageContainer() *storage.Container
	Secrets() secrets.Store
	Postgres() PostgresService
	Runners() RunnersService
	ContainerRegistry() ContainerRegistryService
	ImageVersions() ImageVersionsService
	Settings() SettingsService
	Dinds() DindService
	S3() S3Service
	ConfigResolver() ServiceConfigResolver
}

type ContainerService interface {
	ListSmerds(ctx context.Context, req *velez_api.ListSmerds_Request) (*velez_api.ListSmerds_Response, error)
	DropSmerds(ctx context.Context, req *velez_api.DropSmerd_Request) (*velez_api.DropSmerd_Response, error)
	InspectSmerd(ctx context.Context, environment, contID string) (*velez_api.Smerd, error)
	ListContainers(ctx context.Context, req *velez_api.ListContainers_Request) (*velez_api.ListContainers_Response, error)
	GetContainer(ctx context.Context, req *velez_api.GetContainer_Request) (*velez_api.DockerContainer, error)
	FinishOnboarding(
		ctx context.Context, req *velez_api.FinishOnboarding_Request,
	) (*velez_api.FinishOnboarding_Response, error)

	ConnectToNetwork(ctx context.Context, req domain.Connection) error
	DisconnectFromNetwork(ctx context.Context, req domain.Connection) error
}

type ImageVersionsService interface {
	ListContainerImageVersions(
		ctx context.Context, req *velez_api.ListContainerImageVersions_Request,
	) (*velez_api.ListContainerImageVersions_Response, error)
}

// VervonomiconSource reads the raw .verv/ descriptor files out of a
// service's image. Declared here where it's consumed (verv_services) rather
// than next to its implementation -
// internal/service/service_manager/vervonomicon.ImageSource.
type VervonomiconSource interface {
	Read(ctx context.Context, serviceName, imageName string) (map[string][]byte, error)
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

	// IsStatefull reports whether this node runs against a real Postgres
	// backend (statefull/cluster mode) rather than the single-node/dev
	// in-memory one. Transport gates statefull-only features - e.g. picking a
	// registry by explicit id in SearchImages - on it.
	IsStatefull() bool

	// Registry management (velez.registries) - container image registries
	// (Docker Hub, generic v2) used by SearchImages.
	ListRegistries(ctx context.Context) ([]domain.Registry, error)
	GetRegistry(ctx context.Context, id int64) (domain.Registry, error)
	CreateRegistry(ctx context.Context, req domain.CreateRegistryReq) (domain.Registry, error)
	UpdateRegistry(ctx context.Context, req domain.UpdateRegistryReq) (domain.Registry, error)
	DeleteRegistry(ctx context.Context, req domain.DeleteRegistryReq) error

	// Vervonomicon declarative deployment (docs/features/vervonomicon.md).
	GetVervonomicon(ctx context.Context, req domain.GetVervonomiconReq) (domain.VervonomiconResult, error)
	CreateDeployFromVervonomicon(ctx context.Context, req domain.CreateDeployFromVervonomiconReq) error
}

type NodeService interface {
	ListNodes(ctx context.Context, req domain.ListNodesReq) (domain.NodesList, error)
}

type PluginService interface {
	ListPlugins(ctx context.Context) (*velez_api.ListPlugins_Response, error)
}

// PostgresService provides Postgres-as-a-Service: a PG instance is a normal
// Velez service, deployed through the ordinary VervServicesService
// .CreateNewDeploy path from the builtin postgres vervonomicon descriptor -
// see docs/features/pgaas_and_registry_plugin.md section 3.
type PostgresService interface {
	ListPgInstances(ctx context.Context, req domain.ListPgInstancesReq) (domain.PgInstanceList, error)
	CreatePgInstance(ctx context.Context, req domain.CreatePgInstanceReq) (domain.PgInstanceView, error)
	DropPgInstance(ctx context.Context, name string) error
	// GetPgInstanceCredentials is the only PostgresService operation that
	// resolves a secret_ref to its plaintext value.
	GetPgInstanceCredentials(ctx context.Context, name string) (domain.PgInstanceCredentials, error)
}

// RunnersService provides Runners-as-a-Service: git-agnostic self-hosted CI
// runner provisioning - see runners_api.proto's RunnersAPI. GitHub Actions is
// the first provider; others are added as new runneraas.Provider
// implementations, never as contract changes.
type RunnersService interface {
	ListRunners(ctx context.Context, req domain.ListRunnersReq) (domain.RunnerList, error)
	CreateRunner(ctx context.Context, req domain.CreateRunnerReq) error
	DropRunner(ctx context.Context, name string) error
	// ReregisterRunner re-runs registration against the runner's existing
	// container without minting a new token or deploying a new container.
	ReregisterRunner(ctx context.Context, name string) error
	// GetRunnerCredentials is the only RunnersService operation that
	// resolves a secret_ref to its plaintext value.
	GetRunnerCredentials(ctx context.Context, name string) (domain.RunnerCredentials, error)
	// GetRunnerConfig reads the runner's currently stored provider config.
	GetRunnerConfig(ctx context.Context, name string) (domain.RunnerConfig, error)
	// UpdateRunnerConfig persists new provider config values without
	// applying them - see domain.UpdateRunnerConfigResult's doc comment.
	UpdateRunnerConfig(ctx context.Context, req domain.UpdateRunnerConfigReq) (domain.UpdateRunnerConfigResult, error)
	// RedeployRunner recreates the runner's container, overlaying its
	// currently stored docker_socket_address as DOCKER_HOST.
	RedeployRunner(ctx context.Context, name string) error
}

// ContainerRegistryService provides Container-Registry-as-a-Service: a
// registry instance is a normal Velez service (plus an optional UI sidecar
// service), deployed through the ordinary VervServicesService.CreateNewDeploy
// path from the builtin registry/registry_ui vervonomicon descriptors - see
// docs/features/pgaas_and_registry_plugin.md section 4. Unlike PostgresService,
// CreateRegistryInstance is multi-step and runs through the jobs engine
// rather than calling CreateNewDeploy directly - it only enqueues the task
// and returns, it never waits for a terminal status; callers watch progress
// through TasksApi.WatchTask(name, jobs.CreateRegistryInstanceAction) and
// refetch ListRegistryInstances once it reaches DONE.
type ContainerRegistryService interface {
	ListRegistryInstances(
		ctx context.Context, req domain.ListRegistryInstancesReq,
	) (domain.RegistryInstanceList, error)
	CreateRegistryInstance(ctx context.Context, req domain.CreateRegistryInstanceReq) error
	DropRegistryInstance(ctx context.Context, name string) error
	// GetRegistryInstanceCredentials is the only ContainerRegistryService
	// operation that resolves a secret_ref to its plaintext value.
	GetRegistryInstanceCredentials(ctx context.Context, name string) (domain.RegistryInstanceCredentials, error)
}

type SettingsService interface {
	GetSettings(ctx context.Context) (domain.Settings, error)
	UpdateSettings(ctx context.Context, req domain.UpdateSettingsReq) (domain.Settings, error)
	GetSysboxStatus(ctx context.Context) (domain.SysboxStatus, error)
	RunSysboxSmokeTest(ctx context.Context) (domain.SysboxSmokeTestResult, error)
}

type DindService interface {
	CreateDind(ctx context.Context, req domain.CreateDindReq) error
	ListDinds(ctx context.Context) ([]domain.DindView, error)
	DropDind(ctx context.Context, name string) error
}

// S3Service - Garage-backed S3 instances, their buckets and keys.
//
//nolint:interfacebloat
type S3Service interface {
	// CreateInstance enqueues the create_s3_instance task; the caller watches
	// it by the instance name.
	CreateInstance(ctx context.Context, req *velez_api.CreateS3Instance_Request) error
	ListInstances(ctx context.Context, paging domain.Paging) ([]domain.S3Instance, uint64, error)
	DropInstance(ctx context.Context, name string) error
	GetInstanceCredentials(ctx context.Context, name string) (domain.S3InstanceCredentials, error)
	ListBuckets(ctx context.Context, instanceName string) ([]domain.S3Bucket, error)
	CreateBucket(ctx context.Context, instanceName, bucketName, ownerService string) (domain.S3Bucket, error)
	DeleteBucket(ctx context.Context, instanceName, bucketName string) error
	SetBucketAccess(ctx context.Context, instanceName string, access domain.S3BucketAccess) (domain.S3Bucket, error)
	ListKeys(ctx context.Context, instanceName string) ([]domain.S3Key, error)
	CreateKey(
		ctx context.Context, instanceName, keyName string, access []domain.S3BucketAccess,
	) (domain.S3KeyCredentials, error)
	DeleteKey(ctx context.Context, instanceName, accessKeyId string) error
	GetKeyCredentials(ctx context.Context, instanceName, accessKeyId string) (domain.S3KeyCredentials, error)
}

// ServiceConfigResolver - one config per service. Writes go to matreshka
// when it is enabled, else straight onto the CreateSmerd request; reads come
// from matreshka, else from the running container.
type ServiceConfigResolver interface {
	WriteEnv(ctx context.Context, request *velez_api.CreateSmerd_Request, env map[string]string) error
	WriteFile(ctx context.Context, request *velez_api.CreateSmerd_Request, path string, content []byte) error
	ReadEnv(ctx context.Context, serviceName, environment string) (map[string]string, error)
	ReadFile(ctx context.Context, serviceName, environment, path string) ([]byte, error)
	Delete(ctx context.Context, serviceName string) error
}
