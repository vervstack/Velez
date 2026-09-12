//go:build e2e_full

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/containerd/errdefs"
	dockernetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	rtb "go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	helloWorldImageV0015 = "vervstack/hello_world:v0.0.15"
	postgresAlias        = "postgres"
	serviceTypeWeb       = "web"

	// Container names double as Docker hostnames, capped at 64 characters -
	// GetServiceName(t) on a RunPlaneSuite subtest blows past that, so this
	// suite uses its own short, suite-unique names instead.
	hwClusterNetworkName   = "e2e_hwcluster_net"
	hwClusterPgName        = "e2e_hwcluster_pg"
	hwClusterPgAppName     = "e2e_hwcluster_pg_app"
	hwClusterSqliteAppName = "e2e_hwcluster_sqlite_app"
)

type HelloWorldClusterSuite struct {
	suite.Suite

	plane Plane

	env          *TestEnvironment
	dockerClient client.APIClient

	networkName string

	pgName string

	pgAppName     string
	sqliteAppName string

	pgSmerd        *velez_api.Smerd
	pgAppSmerd     *velez_api.Smerd
	sqliteAppSmerd *velez_api.Smerd
}

func (s *HelloWorldClusterSuite) SetupTest() {
	t := s.T()

	s.env = s.plane.NewEnvironment(t)
	s.dockerClient = s.env.Custom.NodeClients.Docker().Client()

	s.pgName = hwClusterPgName

	s.pgAppName = hwClusterPgAppName
	s.sqliteAppName = hwClusterSqliteAppName
}

func (s *HelloWorldClusterSuite) Test_ConnectedCluster() {
	t := s.T()

	s._prepareNetwork()

	s._preparePostgresContainer()

	s._preparePgApp()
	s._prepareSqliteApp()

	// Assert obligatory labels on the PG hello_world
	checkClientLabels(t, s.pgAppSmerd.GetLabels(), map[string]string{
		labels.VervServiceLabel: s.pgAppName,
		labels.DependsOnLabel:   postgresAlias,
		labels.ServiceTypeLabel: serviceTypeWeb,
	})

	// Assert all labels on the SQLite hello_world
	checkClientLabels(t, s.sqliteAppSmerd.GetLabels(), map[string]string{
		labels.VervServiceLabel: s.sqliteAppName,
		labels.Sidecar:          labelValueFalse,
		labels.AutoUpgrade:      labelValueFalse,
		labels.DependsOnLabel:   s.pgAppName,
		labels.DescriptionLabel: "Hello World SQLite instance",
		labels.ServiceTypeLabel: serviceTypeWeb,
		labels.TeamLabel:        "test-team",
		labels.RepoLabel:        "github.com/godverv/hello_world",
		labels.PortLabel:        "80",
		labels.EnvLabel:         "test",
	})

	s._testAPIIsolation()
}

func (s *HelloWorldClusterSuite) _testAPIIsolation() {
	s.T().Helper()

	t := s.T()
	ctx := t.Context()

	require.NotEmpty(t, s.pgAppSmerd.GetPorts(), "pg app must have exposed ports")
	require.NotEmpty(t, s.sqliteAppSmerd.GetPorts(), "sqlite app must have exposed ports")

	pgBase := "http://" + dindHostAddr(t, s.pgAppSmerd.GetPorts()[0].GetExposedTo())
	sqliteBase := "http://" + dindHostAddr(t, s.sqliteAppSmerd.GetPorts()[0].GetExposedTo())

	s._waitForApp(ctx, t, pgBase)
	s._waitForApp(ctx, t, sqliteBase)

	// 1. sqlite get '1' — empty store
	s._assertGetKey(ctx, t, sqliteBase, "1", false, "")

	// 2. pg get '1' — empty store
	s._assertGetKey(ctx, t, pgBase, "1", false, "")

	// 3. set '1' in sqlite, set '2' in pg
	s._setKey(ctx, t, sqliteBase, "1", "val1")
	s._setKey(ctx, t, pgBase, "2", "val2")

	// 4. sqlite get '1' — found locally
	s._assertGetKey(ctx, t, sqliteBase, "1", true, "val1")

	// 5. pg get '1' — not in pg (only in sqlite)
	s._assertGetKey(ctx, t, pgBase, "1", false, "")

	// 6. both get '2' — sqlite proxies to peer, pg finds it locally
	s._assertGetKey(ctx, t, sqliteBase, "2", true, "val2")
	s._assertGetKey(ctx, t, pgBase, "2", true, "val2")
}

func (s *HelloWorldClusterSuite) _waitForApp(ctx context.Context, t *testing.T, baseURL string) {
	s.T().Helper()

	t.Helper()
	require.Eventually(t, func() bool {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/info", nil)
		if err != nil {
			return false
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return false
		}

		_ = resp.Body.Close()

		return resp.StatusCode == http.StatusOK
	}, 30*time.Second, 500*time.Millisecond, "app at %s did not become ready", baseURL)
}

func (s *HelloWorldClusterSuite) _setKey(ctx context.Context, t *testing.T, baseURL, key, value string) {
	s.T().Helper()

	t.Helper()

	body := fmt.Sprintf(`{"vals":{"values":[{"key":%q,"value":%q}]}}`, key, value)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/set", strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func (s *HelloWorldClusterSuite) _assertGetKey(
	ctx context.Context,
	t *testing.T,
	baseURL, key string,
	expectFound bool,
	expectedValue string,
) {
	s.T().Helper()

	t.Helper()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/get/"+key, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })

	if !expectFound {
		require.NotEqual(t, http.StatusOK, resp.StatusCode, "expected key %q to be absent", key)

		return
	}

	require.Equal(t, http.StatusOK, resp.StatusCode, "expected key %q to be present", key)

	var result struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, expectedValue, result.Value)
}

func (s *HelloWorldClusterSuite) _preparePostgresContainer() {
	s.T().Helper()

	t := s.T()
	ctx := t.Context()

	timeoutSec := uint32(5)

	testCaseNetwork := &velez_api.NetworkBind{
		NetworkName: s.networkName,
		Aliases:     []string{postgresAlias},
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
			labels.VervServiceLabel: postgresAlias,
			labels.ServiceTypeLabel: "resource.database",
			"VERV_SERVICE_OWNER":    s.pgAppName,
		},
		IgnoreConfig: true,
	}

	s.pgSmerd = s.env.CreateSmerd(t, pgReq)
	require.Equal(t, velez_api.Smerd_running, s.pgSmerd.GetStatus())

	require.Eventually(t, func() bool {
		info, inspectErr := s.dockerClient.ContainerInspect(ctx, s.pgSmerd.GetUuid())
		if inspectErr != nil {
			return false
		}

		return info.State.Health != nil && info.State.Health.Status == "healthy"
	}, 30*time.Second, 500*time.Millisecond, "postgres did not become healthy in time")
}

func (s *HelloWorldClusterSuite) _prepareNetwork() {
	s.T().Helper()

	t := s.T()
	ctx := t.Context()

	// Uses docker network for now. Won't later
	// TODO

	s.networkName = hwClusterNetworkName

	err := s.dockerClient.NetworkRemove(ctx, s.networkName)
	if err != nil && !errdefs.IsNotFound(err) {
		require.NoError(t, err)
	}

	createNetOpts := dockernetwork.CreateOptions{
		Driver: "bridge",
	}

	_, err = s.dockerClient.NetworkCreate(ctx, s.networkName, createNetOpts)
	require.NoError(t, err)
}

func (s *HelloWorldClusterSuite) TeardownTest() {
	s.T().Helper()

	t := s.T()
	ctx := t.Context()

	err := s.dockerClient.NetworkRemove(ctx, s.networkName)
	if err != nil && !errdefs.IsNotFound(err) {
		assert.NoError(t, err)
	}
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
			labels.DependsOnLabel:   postgresAlias,
			labels.ServiceTypeLabel: serviceTypeWeb,
		},
		IgnoreConfig:  true,
		UseImagePorts: true,
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
			labels.CreatedWithVelezLabel: labelValueTrue,
			labels.Sidecar:               labelValueFalse,
			labels.VervServiceLabel:      s.sqliteAppName,
			labels.MatreshkaConfigLabel:  labelValueFalse,
			labels.AutoUpgrade:           labelValueFalse,
			labels.DependsOnLabel:        s.pgAppName,
			labels.DescriptionLabel:      "Hello World SQLite instance",
			labels.ServiceTypeLabel:      serviceTypeWeb,
			labels.TeamLabel:             "test-team",
			labels.RepoLabel:             "github.com/godverv/hello_world",
			labels.PortLabel:             "80",
			labels.EnvLabel:              "test",
		},
		IgnoreConfig:  true,
		UseImagePorts: true,
	}

	s.sqliteAppSmerd = s.env.CreateSmerd(t, sqliteReq)
}

func Test_HelloWorldCluster(t *testing.T) {
	t.Parallel()
	RunPlaneSuite(t, ClusterPlanes, func(plane Plane) suite.TestingSuite {
		return &HelloWorldClusterSuite{plane: plane}
	})
}

// dindHostAddr translates a DinD-side port Velez exposed a container on
// into the bootstrap-host address this process can dial it at.
func dindHostAddr(t *testing.T, exposedTo uint32) string {
	t.Helper()

	addr, ok := sharedDind.Addr(int(exposedTo))
	require.True(t, ok, "dind did not publish container port %d (outside the pinned band?)", exposedTo)

	return addr
}
