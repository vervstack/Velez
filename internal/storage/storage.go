package storage

import (
	"context"
	"database/sql"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/domain"
	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/jobs_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/plugins_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

var (
	ErrAlreadyExists = rerrors.New("already exists")
	ErrNotFound      = rerrors.New("not found")
)

//nolint:interfacebloat
type Storage interface {
	Nodes() NodesStorage
	Services() ServicesStorage
	Deployments() DeploymentsStorage
	Plugins() PluginsStorage
	ServiceDependencies() ServiceDependenciesStorage
	ServiceResources() ServiceResourcesStorage
	Tasks() TasksStorage
	Jobs() JobsStorage
	Environments() EnvironmentsStorage
	Registries() RegistriesStorage
	ResourceBoxes() ResourceBoxesStorage

	TxManager() *sqldb.TxManager
}

type ServicesStorage interface {
	GetByName(ctx context.Context, name string) (domain.Service, error)
	UpsertService(ctx context.Context, name string) error
	Delete(ctx context.Context, name string) error

	List(ctx context.Context, req domain.ListServicesReq) (domain.ServiceList, error)
}

type DeploymentsStorage interface {
	List(ctx context.Context, req domain.ListDeploymentsReq) ([]domain.Deployment, error)
	ListDeployments(ctx context.Context, req domain.ListDeploymentsReq) (domain.DeploymentList, error)

	deployments_queries.Querier
	WithTx(tx *sql.Tx) *deployments_queries.Queries
}

type NodesStorage interface {
	InitNode(ctx context.Context, region string) error
	UpdateOnline(ctx context.Context, cpuPercent, memPercent float64) error
	List(ctx context.Context, req domain.ListNodesReq) (domain.NodesList, error)
}

type PluginsStorage interface {
	ListPlugins(ctx context.Context) ([]domain.PluginBaseInfo, error)
	UpsertPlugin(ctx context.Context, arg plugins_queries.UpsertPluginParams) error
}

type ServiceDependenciesStorage interface {
	UpsertDependency(ctx context.Context, source, target, proto string) error
	GetDependencies(ctx context.Context, serviceName string) ([]domain.ServiceDependency, error)
	GetCallers(ctx context.Context, serviceName string) ([]domain.ServiceDependency, error)
}

type ServiceResourcesStorage interface {
	GetResources(ctx context.Context, serviceName string) ([]domain.BoundResource, error)
	UpsertResource(ctx context.Context, serviceName, resourceName, resourceType string) error
}

type TasksStorage interface {
	tasks_queries.Querier
	WithTx(tx *sql.Tx) *tasks_queries.Queries
}

type JobsStorage interface {
	jobs_queries.Querier
	WithTx(tx *sql.Tx) *jobs_queries.Queries
}

// EnvironmentsStorage - CRUD over velez.environments.
//
// Implementations: internal/storage/environments.NewPg (postgres/cluster mode)
// and internal/storage/environments.NewStatic (in-memory, used by
// local_storage in single-node/dev mode and as a test double).
type EnvironmentsStorage interface {
	ListEnvironments(ctx context.Context) ([]domain.Environment, error)
	GetEnvironmentByID(ctx context.Context, id int64) (domain.Environment, error)
	GetEnvironmentByName(ctx context.Context, name string) (domain.Environment, error)
	CreateEnvironment(ctx context.Context, req domain.CreateEnvironmentReq) (domain.Environment, error)
	UpdateEnvironment(ctx context.Context, req domain.UpdateEnvironmentReq) (domain.Environment, error)
	DeleteEnvironment(ctx context.Context, id int64) error
}

// RegistriesStorage - CRUD over velez.registries.
//
// Implementations: internal/storage/registries.NewPg (postgres/cluster mode)
// and internal/storage/registries.NewStatic (in-memory, used by local_storage
// in single-node/dev mode).
type RegistriesStorage interface {
	ListRegistries(ctx context.Context) ([]domain.Registry, error)
	GetRegistryByID(ctx context.Context, id int64) (domain.Registry, error)
	CreateRegistry(ctx context.Context, req domain.CreateRegistryReq) (domain.Registry, error)
	UpdateRegistry(ctx context.Context, req domain.UpdateRegistryReq) (domain.Registry, error)
	DeleteRegistry(ctx context.Context, id int64) error

	// ClearDefaultRegistry unsets is_default on every row. Called by the
	// service layer, inside the same transaction as a Create/Update that sets
	// is_default = true, to enforce the single-default invariant (at most one
	// registry may be default). Not part of the spec draft interface - added
	// because the service needs to reach it through this interface to run it
	// tx-bound alongside Create/Update.
	ClearDefaultRegistry(ctx context.Context) error

	WithTx(tx *sql.Tx) RegistriesStorage
}

// ResourceBoxesStorage - reads over velez.resource_boxes, the sizing tiers
// docs/features/vervonomicon.md's "Boxes" section resolves app.box /
// resources[].box against. Structurally the same shape as
// internal/service/service_manager/vervonomicon.BoxLookup, declared
// separately here (rather than storage importing the service package) to
// keep storage -> service a one-way dependency.
//
// Implementations: internal/storage/postgres.Storage.ResourceBoxes()
// (postgres/cluster mode) and internal/storage/resource_boxes.NewStatic
// (in-memory, used by local_storage in single-node/dev mode, seeded with the
// same three builtin tiers the postgres migration seeds).
type ResourceBoxesStorage interface {
	GetBox(ctx context.Context, name string) (verv.Box, error)
	ListBoxes(ctx context.Context) ([]verv.Box, error)
}
