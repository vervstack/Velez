package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	rtb "go.redsock.ru/toolbox"
	"google.golang.org/protobuf/proto"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/tests/config_mocks"
)

type LifecycleSuite struct {
	suite.Suite
}

func (s *LifecycleSuite) Test_Stateless_HelloWorld() {
	t := s.T()

	env := NewEnvironment(t)

	req := &velez_api.CreateSmerd_Request{
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
	}

	runLifecycle(t, env, req, func(_ *testing.T, _ *velez_api.Smerd) {})
}

func (s *LifecycleSuite) Test_Stateless_HelloWorld_WithHealthcheck() {
	t := s.T()

	env := NewEnvironment(t)

	req := &velez_api.CreateSmerd_Request{
		ImageName: HelloWorldAppImage,
		Healthcheck: &velez_api.Container_Healthcheck{
			IntervalSecond: 1,
			Retries:        3,
		},
		IgnoreConfig: true,
	}
	runLifecycle(t, env, req, func(_ *testing.T, _ *velez_api.Smerd) {})
}

func (s *LifecycleSuite) Test_Stateless_HelloWorld_DefaultConfig() {
	t := s.T()

	env := NewEnvironment(t)

	req := &velez_api.CreateSmerd_Request{
		ImageName: HelloWorldAppImage,
	}
	runLifecycle(t, env, req, func(t *testing.T, smerd *velez_api.Smerd) {
		t.Helper()

		require.Equal(t, labelValueTrue, smerd.GetLabels()[labels.MatreshkaConfigLabel])
		require.Equal(t, smerd.GetName(), smerd.GetEnv()["VERV_NAME"])
	})
}

func (s *LifecycleSuite) Test_Stateless_Nginx() {
	t := s.T()

	env := NewEnvironment(t)

	req := &velez_api.CreateSmerd_Request{
		ImageName:     "nginx:alpine",
		IgnoreConfig:  true,
		UseImagePorts: true,
	}
	runLifecycle(t, env, req,
		func(t *testing.T, smerd *velez_api.Smerd) {
			t.Helper()

			require.NotEmpty(t, smerd.GetPorts())
		})
}

func (s *LifecycleSuite) Test_Stateless_Postgres() {
	timeoutSec := uint32(5)

	t := s.T()

	env := NewEnvironment(t)

	req := &velez_api.CreateSmerd_Request{
		ImageName: PostgresImage,
		Env:       map[string]string{"POSTGRES_HOST_AUTH_METHOD": "trust"},
		Healthcheck: &velez_api.Container_Healthcheck{
			Command:        rtb.ToPtr("pg_isready -U postgres"),
			IntervalSecond: 2,
			TimeoutSecond:  &timeoutSec,
			Retries:        5,
		},
		IgnoreConfig:  true,
		UseImagePorts: true,
	}
	runLifecycle(t, env, req,
		func(t *testing.T, smerd *velez_api.Smerd) {
			t.Helper()

			require.Len(t, smerd.GetPorts(), 1)
			require.EqualValues(t, 5432, smerd.GetPorts()[0].GetServicePortNumber())
		})
}

func (s *LifecycleSuite) Test_StatelessMode_Loki() {
	t := s.T()

	env := NewEnvironment(t)

	req := &velez_api.CreateSmerd_Request{
		ImageName: "grafana/loki:main-bc418c4",
		Settings: &velez_api.Container_Settings{
			Network: []*velez_api.NetworkBind{
				{NetworkName: "redsockru"},
			},
		},
		Restart: &velez_api.RestartPolicy{
			Type: velez_api.RestartPolicyType_always,
		},
		Plain: []*velez_api.FileConfig{
			{
				Path:    "/etc/loki/local-config.yaml",
				Content: config_mocks.Loki,
			},
		},
	}

	runLifecycle(t, env, req, func(_ *testing.T, _ *velez_api.Smerd) {})
}

// Test_ClusterMode_* subtests below run in parallel against one shared
// matreshka container fixture (see WithMatreshka / main_test.go). This only
// works because they live in this package/test binary — don't move these
// to another package without reading main_test.go first.
func (s *LifecycleSuite) Test_ClusterMode_HelloWorld() {
	t := s.T()

	env := NewEnvironment(t, WithMatreshka())

	req := &velez_api.CreateSmerd_Request{
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
	}
	runLifecycle(t, env, req, func(_ *testing.T, _ *velez_api.Smerd) {})
}

func (s *LifecycleSuite) Test_ClusterMode_PlainNginx() {
	t := s.T()

	env := NewEnvironment(t, WithMatreshka())

	req := &velez_api.CreateSmerd_Request{
		ImageName:     "nginx:alpine",
		IgnoreConfig:  true,
		UseImagePorts: true,
	}
	runLifecycle(t, env, req, func(t *testing.T, smerd *velez_api.Smerd) {
		t.Helper()

		require.NotEmpty(t, smerd.GetPorts())
	})
}

func (s *LifecycleSuite) Test_ClusterMode_Postgres() {
	t := s.T()

	env := NewEnvironment(t, WithMatreshka())

	timeoutSec := uint32(5)

	req := &velez_api.CreateSmerd_Request{
		ImageName: PostgresImage,
		Env:       map[string]string{"POSTGRES_HOST_AUTH_METHOD": "trust"},
		Healthcheck: &velez_api.Container_Healthcheck{
			Command:        rtb.ToPtr("pg_isready -U postgres"),
			IntervalSecond: 2,
			TimeoutSecond:  &timeoutSec,
			Retries:        5,
		},
		IgnoreConfig:  true,
		UseImagePorts: true,
	}

	runLifecycle(t, env, req,
		func(t *testing.T, smerd *velez_api.Smerd) {
			t.Helper()

			require.Len(t, smerd.GetPorts(), 1)
			require.EqualValues(t, 5432, smerd.GetPorts()[0].GetServicePortNumber())
		})
}

func (s *LifecycleSuite) Test_DropSmerd_ByUuid() {
	t := s.T()

	env := NewEnvironment(t)
	ctx := t.Context()
	name := GetServiceName(t)

	req := &velez_api.CreateSmerd_Request{
		Name:         name,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
	}

	created := env.CreateSmerd(t, req)
	require.NotEmpty(t, created.GetUuid())

	dropReq := &velez_api.DropSmerd_Request{
		Uuids: []string{created.GetUuid()},
	}
	dropped := env.DropSmerd(ctx, t, dropReq)

	require.Empty(t, dropped.GetFailed())
	require.Equal(t, []string{created.GetUuid()}, dropped.GetSuccessful())

	listReq := &velez_api.ListSmerds_Request{Name: rtb.ToPtr(name)}

	listed := env.ListSmerds(t, ctx, listReq)
	for _, smerd := range listed.GetSmerds() {
		require.NotEqual(t, created.GetUuid(), smerd.GetUuid(), "expected dropped smerd to no longer be listed")
	}
}

func Test_Lifecycle(t *testing.T) {
	t.Skip("flaky: intermittent server-manager startup race, see docs/plans/e2e_flaky_lifecycle_matreshka.md#3")
	suite.Run(t, new(LifecycleSuite))
}

func runLifecycle(
	t *testing.T,
	env *TestEnvironment,
	req *velez_api.CreateSmerd_Request,
	verify func(t *testing.T, smerd *velez_api.Smerd),
) {
	t.Helper()

	ctx := t.Context()
	name := GetServiceName(t)

	clonedReq, ok := proto.Clone(req).(*velez_api.CreateSmerd_Request)
	require.True(t, ok, "proto.Clone must preserve the concrete CreateSmerd_Request type")

	req = clonedReq
	req.Name = name

	t.Logf(`Creating smerd. Req: %v`, req)

	created := env.CreateSmerd(t, req)
	require.Equal(t, name, created.GetName())
	require.Equal(t, velez_api.Smerd_running.String(), created.GetStatus().String())
	require.NotEmpty(t, created.GetUuid())
	require.NotNil(t, created.GetCreatedAt())

	if req.GetUseImagePorts() {
		require.NotEmpty(t, created.GetPorts())
	}

	verify(t, created)

	listReq := &velez_api.ListSmerds_Request{Name: rtb.ToPtr(name)}

	listed := env.ListSmerds(t, ctx, listReq)

	var actualSmerd *velez_api.Smerd

	for _, smerd := range listed.GetSmerds() {
		if name == smerd.GetName() {
			actualSmerd = smerd
		}
	}

	require.NotNil(t, actualSmerd)
	require.Equal(t, created.GetUuid(), actualSmerd.GetUuid())
}
