package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	rtb "go.redsock.ru/toolbox"
	"google.golang.org/protobuf/proto"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

type LifecycleSuite struct {
	suite.Suite
}

func (s *LifecycleSuite) Test_Stateless_HelloWorld() {
	t := s.T()
	t.Parallel()

	env := NewEnvironment(t)

	req := &velez_api.CreateSmerd_Request{
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
	}

	runLifecycle(t, env, req, func(t *testing.T, smerd *velez_api.Smerd) {})
}

func (s *LifecycleSuite) Test_Stateless_PlainNginx() {
	t := s.T()
	t.Parallel()

	env := NewEnvironment(t)

	req := &velez_api.CreateSmerd_Request{
		ImageName:     "nginx:alpine",
		IgnoreConfig:  true,
		UseImagePorts: true,
	}
	runLifecycle(t, env, req,
		func(t *testing.T, smerd *velez_api.Smerd) {
			require.NotEmpty(t, smerd.Ports)
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
			require.Len(t, smerd.Ports, 1)
			require.EqualValues(t, 5432, smerd.Ports[0].ServicePortNumber)
		})
}

func (s *LifecycleSuite) Test_ClusterMode_HelloWorld() {
	t := s.T()

	env := NewEnvironment(t, WithMatreshka())

	req := &velez_api.CreateSmerd_Request{
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
	}
	runLifecycle(t, env, req, func(t *testing.T, smerd *velez_api.Smerd) {})
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
		require.NotEmpty(t, smerd.Ports)
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

	runLifecycle(t, env, req, func(t *testing.T, smerd *velez_api.Smerd) {
		require.Len(t, smerd.Ports, 1)
		require.EqualValues(t, 5432, smerd.Ports[0].ServicePortNumber)
	})
}

func Test_Lifecycle(t *testing.T) {
	suite.Run(t, new(LifecycleSuite))
}

func runLifecycle(t *testing.T, env *TestEnvironment, req *velez_api.CreateSmerd_Request, verify func(t *testing.T, smerd *velez_api.Smerd)) {

	ctx := t.Context()
	name := GetServiceName(t)

	req = proto.Clone(req).(*velez_api.CreateSmerd_Request)
	req.Name = name

	t.Logf(`Creating smerd. Req: %v`, req)

	created := env.CreateSmerd(t, req)
	require.Equal(t, name, created.Name)
	require.Equal(t, velez_api.Smerd_running, created.Status)
	if req.UseImagePorts {
		require.NotEmpty(t, created.Ports)
	}

	verify(t, created)

	listReq := &velez_api.ListSmerds_Request{Name: rtb.ToPtr(name)}

	listed := env.ListSmerds(t, ctx, listReq)
	require.Len(t, listed.Smerds, 1)
	require.Equal(t, created.Uuid, listed.Smerds[0].Uuid)
}
