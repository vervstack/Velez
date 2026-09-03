package verv_services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.redsock.ru/evon"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka"
	"go.vervstack.ru/matreshka/pkg/matreshka/resources"

	"go.vervstack.ru/Velez/internal/domain"
	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
	"go.vervstack.ru/Velez/internal/storage"
)

const (
	testResourceTypePostgres = "postgres"
	testBindsToPostgres      = "data_sources.postgres"
)

// fakeServiceResourcesStorage is a hand-written fake storage.ServiceResourcesStorage
// with an injectable GetResources - allowed at the service layer per this
// repo's testing rules (no fakes for Docker/Postgres-facing code).
type fakeServiceResourcesStorage struct {
	getResourcesFunc func(ctx context.Context, serviceName string) ([]domain.BoundResource, error)
}

func (f *fakeServiceResourcesStorage) GetResources(
	ctx context.Context, serviceName string,
) ([]domain.BoundResource, error) {
	if f.getResourcesFunc != nil {
		return f.getResourcesFunc(ctx, serviceName)
	}

	return nil, nil
}

func (f *fakeServiceResourcesStorage) UpsertResource(_ context.Context, _, _, _ string) error {
	return nil
}

// fakeConfigurationService is a hand-written fake service.ConfigurationService
// with an injectable GetVervFromApi - the only method reconcileResources uses.
type fakeConfigurationService struct {
	getVervFromApiFunc func(ctx context.Context, meta domain.ConfigMeta) (matreshka.AppConfig, error)
}

func (f *fakeConfigurationService) GetVervFromApi(
	ctx context.Context, meta domain.ConfigMeta,
) (matreshka.AppConfig, error) {
	if f.getVervFromApiFunc != nil {
		return f.getVervFromApiFunc(ctx, meta)
	}

	return matreshka.NewEmptyConfig(), nil
}

func (f *fakeConfigurationService) GetEnvFromApi(_ context.Context, _ domain.ConfigMeta) (*evon.Node, error) {
	return nil, nil
}

func (f *fakeConfigurationService) UpdateConfig(_ context.Context, _ domain.AppConfig) error {
	return nil
}

func (f *fakeConfigurationService) GetPlainFromApi(_ context.Context, _ domain.ConfigMeta) ([]byte, error) {
	return nil, nil
}

func (f *fakeConfigurationService) SubscribeOnChanges(_ ...string) error         { return nil }
func (f *fakeConfigurationService) UnsubscribeFromChanges(_ ...string) error     { return nil }
func (f *fakeConfigurationService) GetUpdates() <-chan domain.ConfigurationPatch { return nil }

func TestReconcileResources_NoResourcesInDescriptor(t *testing.T) {
	service := &VervService{
		dataStorage:   &testStorage{},
		configService: &fakeConfigurationService{},
	}

	decisions, err := service.reconcileResources(context.Background(), "my-service", nil)
	require.NoError(t, err)
	require.Empty(t, decisions)
}

func TestReconcileResources_AlreadyConnectedViaBinding(t *testing.T) {
	serviceResources := &fakeServiceResourcesStorage{
		getResourcesFunc: func(_ context.Context, _ string) ([]domain.BoundResource, error) {
			return []domain.BoundResource{{Name: "pg", ResourceType: testResourceTypePostgres}}, nil
		},
	}

	service := &VervService{
		dataStorage:   &testStorage{serviceResources: serviceResources},
		configService: &fakeConfigurationService{},
	}

	pgResource := verv.Resource{Name: "pg", Type: testResourceTypePostgres, BindsTo: testBindsToPostgres}

	decisions, err := service.reconcileResources(context.Background(), "my-service", []verv.Resource{pgResource})
	require.NoError(t, err)
	require.Len(t, decisions, 1)
	require.Equal(t, domain.ResourceAlreadyConnected, decisions[0].Status)
}

func TestReconcileResources_AlreadyConnectedViaMatreshkaHost(t *testing.T) {
	configService := &fakeConfigurationService{
		getVervFromApiFunc: func(_ context.Context, _ domain.ConfigMeta) (matreshka.AppConfig, error) {
			cfg := matreshka.NewEmptyConfig()

			cfg.DataSources = matreshka.DataSources{
				&resources.Postgres{Name: testResourceTypePostgres, Host: "pg.internal"},
			}

			return cfg, nil
		},
	}

	service := &VervService{
		dataStorage:   &testStorage{serviceResources: &fakeServiceResourcesStorage{}},
		configService: configService,
	}

	pgResource := verv.Resource{Name: "pg", Type: testResourceTypePostgres, BindsTo: testBindsToPostgres}

	decisions, err := service.reconcileResources(context.Background(), "my-service", []verv.Resource{pgResource})
	require.NoError(t, err)
	require.Len(t, decisions, 1)
	require.Equal(t, domain.ResourceAlreadyConnected, decisions[0].Status)
}

func TestReconcileResources_MustProvisionWhenNeitherFound(t *testing.T) {
	service := &VervService{
		dataStorage:   &testStorage{serviceResources: &fakeServiceResourcesStorage{}},
		configService: &fakeConfigurationService{},
	}

	pgResource := verv.Resource{Name: "pg", Type: testResourceTypePostgres, BindsTo: testBindsToPostgres}

	decisions, err := service.reconcileResources(context.Background(), "my-service", []verv.Resource{pgResource})
	require.NoError(t, err)
	require.Len(t, decisions, 1)
	require.Equal(t, domain.ResourceMustProvision, decisions[0].Status)
}

func TestReconcileResources_BindingsLookupErrorIsWrapped(t *testing.T) {
	wantErr := rerrors.New("storage unavailable")

	serviceResources := &fakeServiceResourcesStorage{
		getResourcesFunc: func(_ context.Context, _ string) ([]domain.BoundResource, error) {
			return nil, wantErr
		},
	}

	service := &VervService{
		dataStorage:   &testStorage{serviceResources: serviceResources},
		configService: &fakeConfigurationService{},
	}

	pgResource := verv.Resource{Name: "pg", Type: testResourceTypePostgres, BindsTo: testBindsToPostgres}

	_, err := service.reconcileResources(context.Background(), "my-service", []verv.Resource{pgResource})
	require.Error(t, err)
	require.True(t, rerrors.Is(err, wantErr))
}

// TestReconcileResources_MatreshkaLookupErrorIsSwallowed asserts that a
// matreshka config that cannot be read does not fail reconciliation - see
// reconcileResources's own doc comment. A resource with no storage binding
// and no matreshka-resolved host falls back to MUST_PROVISION rather than
// taking the whole call down.
func TestReconcileResources_MatreshkaLookupErrorIsSwallowed(t *testing.T) {
	configService := &fakeConfigurationService{
		getVervFromApiFunc: func(_ context.Context, _ domain.ConfigMeta) (matreshka.AppConfig, error) {
			return matreshka.AppConfig{}, rerrors.New("matreshka unreachable")
		},
	}

	service := &VervService{
		dataStorage:   &testStorage{serviceResources: &fakeServiceResourcesStorage{}},
		configService: configService,
	}

	pgResource := verv.Resource{Name: "pg", Type: testResourceTypePostgres, BindsTo: testBindsToPostgres}

	decisions, err := service.reconcileResources(context.Background(), "my-service", []verv.Resource{pgResource})
	require.NoError(t, err)
	require.Len(t, decisions, 1)
	require.Equal(t, domain.ResourceMustProvision, decisions[0].Status)
}

var _ storage.ServiceResourcesStorage = (*fakeServiceResourcesStorage)(nil)
