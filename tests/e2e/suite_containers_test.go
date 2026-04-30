package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	rtb "go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

type ContainerSuite struct {
	suite.Suite

	env *TestEnvironment
	ctx context.Context
}

func (s *ContainerSuite) SetupSuite() {
	t := s.T()
	t.Parallel()

	s.ctx = t.Context()
	s.env = NewEnvironment(t)
}

// Test_CreateAndList — Suite 1.1: Create a hello_world container and verify
// it appears in ListSmerds with all expected fields populated.
func (s *ContainerSuite) Test_CreateAndList() {
	t := s.T()

	name := GetServiceName(t)

	createReq := &velez_api.CreateSmerd_Request{
		Name:         name,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
	}
	created, err := s.env.CreateSmerd(s.ctx, createReq)
	require.NoError(t, err)

	require.Equal(t, name, created.Name)
	require.Equal(t, velez_api.Smerd_running, created.Status)
	require.NotEmpty(t, created.Uuid)
	require.NotNil(t, created.CreatedAt)

	listReq := &velez_api.ListSmerds_Request{
		Name: rtb.ToPtr(name),
	}
	listed, err := s.env.ListSmerds(s.ctx, listReq)
	require.NoError(t, err)

	require.Len(t, listed.Smerds, 1)
	smerd := listed.Smerds[0]
	require.Equal(t, created.Uuid, smerd.Uuid)
	require.Equal(t, name, smerd.Name)
	require.Equal(t, velez_api.Smerd_running, smerd.Status)
}

func Test_Containers(t *testing.T) {
	suite.Run(t, new(ContainerSuite))
}
