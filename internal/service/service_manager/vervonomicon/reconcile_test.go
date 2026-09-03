package vervonomicon

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.vervstack.ru/matreshka/pkg/matreshka"
	"go.vervstack.ru/matreshka/pkg/matreshka/resources"

	"go.vervstack.ru/Velez/internal/domain"
	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
)

const (
	testResourceTypePostgres = "postgres"
	testResourceTypeRedis    = "redis"
	testBindsToPostgres      = "data_sources.postgres"
	testHostPgInternal       = "pg.internal"
)

func TestReconcileResources_Decision(t *testing.T) {
	pgResource := verv.Resource{
		Name:    "pg",
		Type:    testResourceTypePostgres,
		BindsTo: testBindsToPostgres,
	}

	cases := []struct {
		name            string
		bindings        []domain.BoundResource
		dataSourceHosts map[string]string
		want            domain.ResourceConnectionStatus
	}{
		{
			name:            "placeholder host empty string",
			dataSourceHosts: map[string]string{testBindsToPostgres: placeholderHostEmpty},
			want:            domain.ResourceMustProvision,
		},
		{
			name:            "placeholder host localhost",
			dataSourceHosts: map[string]string{testBindsToPostgres: placeholderHostLocalhost},
			want:            domain.ResourceMustProvision,
		},
		{
			name:            "placeholder host 127.0.0.1",
			dataSourceHosts: map[string]string{testBindsToPostgres: placeholderHostLoopback},
			want:            domain.ResourceMustProvision,
		},
		{
			name:            "placeholder host 0.0.0.0",
			dataSourceHosts: map[string]string{testBindsToPostgres: placeholderHostAllZeros},
			want:            domain.ResourceMustProvision,
		},
		{
			name:            "non-placeholder host",
			dataSourceHosts: map[string]string{testBindsToPostgres: testHostPgInternal},
			want:            domain.ResourceAlreadyConnected,
		},
		{
			name:            "binding present, no matreshka entry",
			bindings:        []domain.BoundResource{{Name: "pg", ResourceType: testResourceTypePostgres}},
			dataSourceHosts: map[string]string{},
			want:            domain.ResourceAlreadyConnected,
		},
		{
			name:            "matreshka entry present, no binding",
			dataSourceHosts: map[string]string{testBindsToPostgres: testHostPgInternal},
			want:            domain.ResourceAlreadyConnected,
		},
		{
			name:            "neither binding nor matreshka entry",
			dataSourceHosts: map[string]string{},
			want:            domain.ResourceMustProvision,
		},
		{
			name:            "binds_to points at a key path that does not exist",
			dataSourceHosts: map[string]string{"data_sources.some_other_resource": testHostPgInternal},
			want:            domain.ResourceMustProvision,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			decisions := ReconcileResources([]verv.Resource{pgResource}, tc.bindings, tc.dataSourceHosts)

			require.Len(t, decisions, 1)
			require.Equal(t, "pg", decisions[0].Name)
			require.Equal(t, testResourceTypePostgres, decisions[0].Type)
			require.Equal(t, tc.want, decisions[0].Status)
		})
	}
}

func TestReconcileResources_EmptyResourceList(t *testing.T) {
	decisions := ReconcileResources(nil, nil, nil)

	require.Empty(t, decisions)
}

func TestReconcileResources_ResourceWithNoBindsTo(t *testing.T) {
	volume := verv.Resource{
		Name: "storage",
		Type: "local_volume",
	}

	decisions := ReconcileResources([]verv.Resource{volume}, nil, map[string]string{"data_sources.storage": "10.0.0.1"})

	require.Len(t, decisions, 1)
	require.Equal(t, domain.ResourceMustProvision, decisions[0].Status)
}

func TestReconcileResources_MultipleResourcesIndependentDecisions(t *testing.T) {
	pg := verv.Resource{Name: "pg", Type: testResourceTypePostgres, BindsTo: testBindsToPostgres}
	redis := verv.Resource{Name: "cache", Type: testResourceTypeRedis, BindsTo: "data_sources.redis"}

	bindings := []domain.BoundResource{{Name: "cache", ResourceType: testResourceTypeRedis}}
	hosts := map[string]string{testBindsToPostgres: placeholderHostLocalhost}

	decisions := ReconcileResources([]verv.Resource{pg, redis}, bindings, hosts)

	require.Len(t, decisions, 2)
	require.Equal(t, domain.ResourceMustProvision, decisions[0].Status)
	require.Equal(t, domain.ResourceAlreadyConnected, decisions[1].Status)
}

func TestDataSourceHosts_ExtractsPostgresAndRedisHosts(t *testing.T) {
	cfg := matreshka.AppConfig{
		DataSources: matreshka.DataSources{
			&resources.Postgres{Name: testResourceTypePostgres, Host: testHostPgInternal},
			&resources.Redis{Name: testResourceTypeRedis, Host: placeholderHostLocalhost},
			&resources.Telegram{Name: "telegram", ApiKey: "secret"},
		},
	}

	hosts := DataSourceHosts(cfg)

	require.Equal(t, map[string]string{
		testBindsToPostgres:  testHostPgInternal,
		"data_sources.redis": placeholderHostLocalhost,
	}, hosts)
}

func TestDataSourceHosts_EmptyDataSources(t *testing.T) {
	hosts := DataSourceHosts(matreshka.AppConfig{})

	require.Empty(t, hosts)
}
