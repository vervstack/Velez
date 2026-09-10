package verv_services

import (
	"context"
	"database/sql"
	"testing"

	"github.com/sqlc-dev/pqtype"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
)

// testStorage is a minimal fake storage.Storage that exposes the injectable
// per-kind fakes below via Services()/Deployments()/TxManager(), mirroring
// the real storage.Container contract: VervService now holds a
// storage.Storage and resolves each sub-storage fresh on every call instead
// of caching the result of each getter at construction time (see
// service.go). Unused getters return nil since no test here exercises them.
type testStorage struct {
	services         storage.ServicesStorage
	deployments      storage.DeploymentsStorage
	environments     storage.EnvironmentsStorage
	registries       storage.RegistriesStorage
	resourceBoxes    storage.ResourceBoxesStorage
	serviceResources storage.ServiceResourcesStorage
	txManager        storage.Transactor
}

func (s *testStorage) IsStatefull() bool                 { return false }
func (s *testStorage) Nodes() storage.NodesStorage       { return nil }
func (s *testStorage) Services() storage.ServicesStorage { return s.services }

func (s *testStorage) Deployments() storage.DeploymentsStorage { return s.deployments }
func (s *testStorage) Plugins() storage.PluginsStorage         { return nil }

func (s *testStorage) ServiceDependencies() storage.ServiceDependenciesStorage { return nil }

func (s *testStorage) ServiceResources() storage.ServiceResourcesStorage {
	return s.serviceResources
}

func (s *testStorage) Environments() storage.EnvironmentsStorage   { return s.environments }
func (s *testStorage) Registries() storage.RegistriesStorage       { return s.registries }
func (s *testStorage) ResourceBoxes() storage.ResourceBoxesStorage { return s.resourceBoxes }
func (s *testStorage) Secrets() storage.SecretsStorage             { return nil }
func (s *testStorage) PgInstances() storage.PgInstancesStorage     { return nil }

func (s *testStorage) Tasks() storage.TasksStorage { return nil }
func (s *testStorage) Jobs() storage.JobsStorage   { return nil }

func (s *testStorage) TxManager() storage.Transactor { return s.txManager }

// fakeTransactor is a storage.Transactor that runs fn immediately with a nil
// *sql.Tx and no locking - enough for tests that only need executeDeployment
// to reach fn, mirroring what local_storage.deployments.Execute does modulo
// the mutex. The mutex/atomicity behavior itself is covered directly in
// internal/storage/local_storage against the real implementation.
type fakeTransactor struct{}

func (fakeTransactor) Execute(fn func(tx *sql.Tx) error) error {
	return fn(nil)
}

// testServicesStorageWithGetByName is a fake storage.ServicesStorage with an
// injectable GetByName, used to exercise the ServiceID-resolution error path
// added to CreateNewDeploy/UpgradeDeploy.
type testServicesStorageWithGetByName struct {
	getByNameFunc func(ctx context.Context, name string) (domain.Service, error)
}

func (m *testServicesStorageWithGetByName) List(
	_ context.Context, _ domain.ListServicesReq,
) (domain.ServiceList, error) {
	return domain.ServiceList{}, nil
}

func (m *testServicesStorageWithGetByName) GetByName(ctx context.Context, name string) (domain.Service, error) {
	if m.getByNameFunc != nil {
		return m.getByNameFunc(ctx, name)
	}

	return domain.Service{}, nil
}

func (m *testServicesStorageWithGetByName) UpsertService(_ context.Context, _ string) error {
	return nil
}

func (m *testServicesStorageWithGetByName) Delete(_ context.Context, _ string) error {
	return nil
}

// testDeploymentsStorage is a fake storage.DeploymentsStorage with injectable
// behavior for the parts of UpgradeDeploy's happy path that run before the
// new GetByName resolution (List, GetSpecificationById), plus a call counter
// on CreateSpecification so tests can prove it is never reached when
// GetByName errors.
type testDeploymentsStorage struct {
	listFunc                 func(ctx context.Context, req domain.ListDeploymentsReq) ([]domain.Deployment, error)
	listDeploymentsFunc      func(ctx context.Context, req domain.ListDeploymentsReq) (domain.DeploymentList, error)
	getSpecificationByIdFunc func(ctx context.Context, id int64) (deployments_queries.GetSpecificationByIdRow, error)

	createSpecificationCalls int
}

func (m *testDeploymentsStorage) List(
	ctx context.Context, req domain.ListDeploymentsReq,
) ([]domain.Deployment, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, req)
	}

	return nil, nil
}

func (m *testDeploymentsStorage) ListDeployments(
	ctx context.Context, req domain.ListDeploymentsReq,
) (domain.DeploymentList, error) {
	if m.listDeploymentsFunc != nil {
		return m.listDeploymentsFunc(ctx, req)
	}

	return domain.DeploymentList{}, nil
}

func (m *testDeploymentsStorage) CreateDeployment(
	_ context.Context, _ deployments_queries.CreateDeploymentParams,
) (any, error) {
	return nil, nil
}

func (m *testDeploymentsStorage) CreateSpecification(
	_ context.Context, _ deployments_queries.CreateSpecificationParams,
) (int64, error) {
	m.createSpecificationCalls++

	return 0, nil
}

func (m *testDeploymentsStorage) GetSpecificationById(
	ctx context.Context, id int64,
) (deployments_queries.GetSpecificationByIdRow, error) {
	if m.getSpecificationByIdFunc != nil {
		return m.getSpecificationByIdFunc(ctx, id)
	}

	return deployments_queries.GetSpecificationByIdRow{}, nil
}

func (m *testDeploymentsStorage) UpdateDeploymentStatus(
	_ context.Context, _ deployments_queries.UpdateDeploymentStatusParams,
) error {
	return nil
}

func (m *testDeploymentsStorage) WithTx(_ *sql.Tx) deployments_queries.Querier {
	return m
}

// TestCreateNewDeploy_GetByNameError proves that when the service lookup
// fails, CreateNewDeploy returns a wrapped error and never touches the
// transaction manager or deployments storage. testStorage.txManager is left
// nil (its zero value) since Execute must never be called on this path — if
// it were, this test would panic on the nil pointer dereference instead of
// passing.
func TestCreateNewDeploy_GetByNameError(t *testing.T) {
	wantErr := rerrors.New("service not found")

	servicesStorage := &testServicesStorageWithGetByName{
		getByNameFunc: func(_ context.Context, _ string) (domain.Service, error) {
			return domain.Service{}, wantErr
		},
	}

	service := &VervService{
		dataStorage: &testStorage{services: servicesStorage},
	}

	err := service.CreateNewDeploy(context.Background(), domain.CreateDeployReq{ServiceName: "my-service"})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if !rerrors.Is(err, wantErr) {
		t.Errorf("expected error to wrap %v, got %v", wantErr, err)
	}
}

// TestUpgradeDeploy_GetByNameError proves that when the service lookup
// fails, UpgradeDeploy returns a wrapped error before opening a transaction
// or calling CreateSpecification. Unlike CreateNewDeploy, UpgradeDeploy
// resolves the running deployment/spec (via deploymentsStorage) before the
// ServiceID lookup, so those fakes must succeed here to reach the GetByName
// call under test.
func TestUpgradeDeploy_GetByNameError(t *testing.T) {
	wantErr := rerrors.New("service not found")

	deploymentsStorage := &testDeploymentsStorage{
		listFunc: func(_ context.Context, _ domain.ListDeploymentsReq) ([]domain.Deployment, error) {
			return []domain.Deployment{
				{
					Id:     1,
					SpecId: 1,
					NodeId: 1,
					Status: deployments_queries.VelezDeploymentStatusRUNNING,
				},
			}, nil
		},
		getSpecificationByIdFunc: func(_ context.Context, _ int64) (deployments_queries.GetSpecificationByIdRow, error) {
			return deployments_queries.GetSpecificationByIdRow{
				VervPayload: pqtype.NullRawMessage{
					RawMessage: []byte(`{"name":"my-service"}`),
					Valid:      true,
				},
			}, nil
		},
	}

	servicesStorage := &testServicesStorageWithGetByName{
		getByNameFunc: func(_ context.Context, _ string) (domain.Service, error) {
			return domain.Service{}, wantErr
		},
	}

	service := &VervService{
		dataStorage: &testStorage{
			services:    servicesStorage,
			deployments: deploymentsStorage,
		},
	}

	req := domain.UpgradeDeployReq{ServiceName: "my-service"}

	err := service.UpgradeDeploy(context.Background(), req)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if !rerrors.Is(err, wantErr) {
		t.Errorf("expected error to wrap %v, got %v", wantErr, err)
	}

	if deploymentsStorage.createSpecificationCalls != 0 {
		t.Errorf(
			"expected CreateSpecification to never be called, got %d calls",
			deploymentsStorage.createSpecificationCalls,
		)
	}
}

// TestCreateNewDeploy_RunsThroughTransactor proves CreateNewDeploy's
// composite spec+deployment write always runs through
// storage.Storage.TxManager().Execute(...) - both backends now hand back a
// non-nil storage.Transactor (see storage.Transactor's doc comment), so
// executeDeployment no longer special-cases a nil TxManager.
func TestCreateNewDeploy_RunsThroughTransactor(t *testing.T) {
	servicesStorage := &testServicesStorageWithGetByName{}
	deploymentsStorage := &testDeploymentsStorage{}

	service := &VervService{
		dataStorage: &testStorage{
			services:    servicesStorage,
			deployments: deploymentsStorage,
			txManager:   fakeTransactor{},
		},
	}

	req := domain.CreateDeployReq{ServiceName: testServiceName}

	err := service.CreateNewDeploy(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if deploymentsStorage.createSpecificationCalls != 1 {
		t.Errorf(
			"expected CreateSpecification to be called once, got %d calls",
			deploymentsStorage.createSpecificationCalls,
		)
	}
}
