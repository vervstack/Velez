package e2e

import (
	"testing"
	"time"

	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	rtb "go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const helloWorldImageV0015 = "vervstack/hello_world:v0.0.15"

type HelloWorldClusterSuite struct {
	suite.Suite

	env          *TestEnvironment
	dockerClient client.APIClient

	networkName string
	svcSuffix   string

	pgName string

	pgAppName     string
	sqliteAppName string

	pgSmerd        *velez_api.Smerd
	pgAppSmerd     *velez_api.Smerd
	sqliteAppSmerd *velez_api.Smerd
}

func (s *HelloWorldClusterSuite) SetupTest() {
	t := s.T()

	s.env = NewEnvironment(t)
	s.dockerClient = s.env.Custom.NodeClients.Docker().Client()

	s.pgName = GetServiceName(t) + "_db"

	s.pgAppName = GetServiceName(t) + "_app_pg"
	s.sqliteAppName = GetServiceName(t) + "_hw_sqlite"

	s.svcSuffix = GetServiceName(t)

}

func (s *HelloWorldClusterSuite) Test_ConnectedCluster() {
	t := s.T()

	s._prepareNetwork()

	s._preparePostgresContainer()

	s._preparePgApp()
	s._prepareSqliteApp()

	// Assert obligatory labels on the PG hello_world
	assert.Equal(t, s.pgAppName, s.pgAppSmerd.Labels[labels.VervServiceLabel])
	assert.Equal(t, "postgres", s.pgAppSmerd.Labels[labels.DependsOnLabel])
	assert.Equal(t, s.svcSuffix, s.pgAppSmerd.Labels[labels.SuffixLabel])
	assert.Equal(t, "web", s.pgAppSmerd.Labels[labels.ServiceTypeLabel])

	// Assert all labels on the SQLite hello_world
	assert.Equal(t, s.sqliteAppName, s.sqliteAppSmerd.Labels[labels.VervServiceLabel])
	assert.Equal(t, "false", s.sqliteAppSmerd.Labels[labels.Sidecar])
	assert.Equal(t, "false", s.sqliteAppSmerd.Labels[labels.AutoUpgrade])
	assert.Equal(t, s.pgAppName, s.sqliteAppSmerd.Labels[labels.DependsOnLabel])
	assert.Equal(t, s.svcSuffix, s.sqliteAppSmerd.Labels[labels.SuffixLabel])
	assert.Equal(t, "Hello World SQLite instance", s.sqliteAppSmerd.Labels[labels.DescriptionLabel])
	assert.Equal(t, "web", s.sqliteAppSmerd.Labels[labels.ServiceTypeLabel])
	assert.Equal(t, "test-team", s.sqliteAppSmerd.Labels[labels.TeamLabel])
	assert.Equal(t, "github.com/godverv/hello_world", s.sqliteAppSmerd.Labels[labels.RepoLabel])
	assert.Equal(t, "80", s.sqliteAppSmerd.Labels[labels.PortLabel])
	assert.Equal(t, "test", s.sqliteAppSmerd.Labels[labels.EnvLabel])
}

func (s *HelloWorldClusterSuite) _preparePostgresContainer() {
	t := s.T()
	ctx := t.Context()

	timeoutSec := uint32(5)

	testCaseNetwork := &velez_api.NetworkBind{
		NetworkName: s.networkName,
		Aliases:     []string{"postgres"},
	}
	pgSettings := &velez_api.Container_Settings{
		Network: []*velez_api.NetworkBind{testCaseNetwork},
	}
	pgHealthcheck := &velez_api.Container_Healthcheck{
		Command:        rtb.ToPtr("pg_isready -U hello -d hello_world"),
		IntervalSecond: 2,
		TimeoutSecond:  &timeoutSec,
		Retries:        5,
	}
	pgReq := &velez_api.CreateSmerd_Request{
		Name:      s.pgName,
		ImageName: PostgresImage,
		Env: map[string]string{
			"POSTGRES_DB":       "hello_world",
			"POSTGRES_USER":     "hello",
			"POSTGRES_PASSWORD": "world",
		},
		Settings:    pgSettings,
		Healthcheck: pgHealthcheck,
		Labels: map[string]string{
			labels.VervServiceLabel: "postgres",
			labels.ServiceTypeLabel: "resource.database",
			labels.DependsOnLabel:   "",
			labels.SuffixLabel:      s.svcSuffix,
			"VERV_SERVICE_OWNER":    s.pgAppName,
		},
		IgnoreConfig: true,
	}

	s.pgSmerd = s.env.CreateSmerd(t, pgReq)
	require.Equal(t, velez_api.Smerd_running, s.pgSmerd.Status)

	require.Eventually(t, func() bool {
		info, inspectErr := s.dockerClient.ContainerInspect(ctx, s.pgSmerd.Uuid)
		if inspectErr != nil {
			return false
		}
		return info.State.Health != nil && info.State.Health.Status == "healthy"
	}, 30*time.Second, 2*time.Second, "postgres did not become healthy in time")
}

func (s *HelloWorldClusterSuite) _prepareNetwork() {
	t := s.T()
	ctx := t.Context()

	// Uses docker network for now. Won't later
	// TODO

	s.networkName = GetServiceName(t) + "_net"

	err := s.dockerClient.NetworkRemove(ctx, s.networkName)
	assert.NoError(t, err)

	createNetOpts := dockernetwork.CreateOptions{
		Driver: "bridge",
	}
	_, err = s.dockerClient.NetworkCreate(ctx, s.networkName, createNetOpts)
	require.NoError(t, err)
}

func (s *HelloWorldClusterSuite) TeardownTest() {
	t := s.T()
	ctx := t.Context()

	err := s.dockerClient.NetworkRemove(ctx, s.networkName)
	assert.NoError(t, err)
}

func (s *HelloWorldClusterSuite) _preparePgApp() {
	t := s.T()

	pgHWNetBind := &velez_api.NetworkBind{
		NetworkName: s.networkName,
		Aliases:     []string{"hello_world_pg"},
	}
	pgHWSettings := &velez_api.Container_Settings{
		Network: []*velez_api.NetworkBind{pgHWNetBind},
	}
	pgHWReq := &velez_api.CreateSmerd_Request{
		Name:      s.pgAppName,
		ImageName: helloWorldImageV0015,
		Env: map[string]string{
			"STATEFULL_PG_URL": "postgres://hello:world@postgres:5432/hello_world?sslmode=disable",
		},
		Settings: pgHWSettings,
		Labels: map[string]string{
			labels.VervServiceLabel: s.pgAppName,
			labels.DependsOnLabel:   "postgres",
			labels.SuffixLabel:      s.svcSuffix,
			labels.ServiceTypeLabel: "web",
		},
		IgnoreConfig: true,
	}

	s.pgAppSmerd = s.env.CreateSmerd(t, pgHWReq)
}

func (s *HelloWorldClusterSuite) _prepareSqliteApp() {
	t := s.T()

	sqliteNetBind := &velez_api.NetworkBind{
		NetworkName: s.networkName,
	}
	sqliteSettings := &velez_api.Container_Settings{
		Network: []*velez_api.NetworkBind{sqliteNetBind},
	}
	sqliteReq := &velez_api.CreateSmerd_Request{
		Name:      s.sqliteAppName,
		ImageName: helloWorldImageV0015,
		Env: map[string]string{
			"PEER_GRPC_URL": "hello_world_pg:80",
		},
		Settings: sqliteSettings,
		Labels: map[string]string{
			labels.CreatedWithVelezLabel: "true",
			labels.Sidecar:               "false",
			labels.VervServiceLabel:      s.sqliteAppName,
			labels.MatreshkaConfigLabel:  "false",
			labels.AutoUpgrade:           "false",
			labels.DependsOnLabel:        s.pgAppName,
			labels.SuffixLabel:           s.svcSuffix,
			labels.DescriptionLabel:      "Hello World SQLite instance",
			labels.ServiceTypeLabel:      "web",
			labels.TeamLabel:             "test-team",
			labels.RepoLabel:             "github.com/godverv/hello_world",
			labels.PortLabel:             "80",
			labels.EnvLabel:              "test",
		},
		IgnoreConfig: true,
	}

	s.sqliteAppSmerd = s.env.CreateSmerd(t, sqliteReq)
}

func Test_HelloWorldCluster(t *testing.T) {
	suite.Run(t, new(HelloWorldClusterSuite))
}
