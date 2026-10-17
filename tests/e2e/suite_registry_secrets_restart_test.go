//go:build e2e_full

package e2e

import (
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

const registrySecretsRestartInstanceName = "e2e-cr-secret-rst"

func (s *ContainerRegistryLifecycleSuite) Test_ContainerRegistry_PasswordSurvivesVelezRestart() {
	t := s.T()
	t.Parallel()

	ctx := t.Context()

	env := s.plane.NewEnvironment(t)
	dockerClient := env.DockerClient()

	removeContainerRegistryInstance(dockerClient, registrySecretsRestartInstanceName)
	t.Cleanup(func() { removeContainerRegistryInstance(dockerClient, registrySecretsRestartInstanceName) })

	createReq := newCreateRegistryInstanceRequest(registrySecretsRestartInstanceName, false)

	_, err := env.ContainerRegistryClient().CreateRegistryInstance(ctx, createReq)
	require.NoError(t, err)

	waitForRegistryInstanceDeploy(t, env, registrySecretsRestartInstanceName, false)

	serviceName := registryServiceName(registrySecretsRestartInstanceName)
	credsReq := &velez_api.GetRegistryInstanceCredentials_Request{Name: serviceName}

	before, err := env.ContainerRegistryClient().GetRegistryInstanceCredentials(ctx, credsReq)
	require.NoError(t, err)
	require.NotEmpty(t, before.GetPassword())

	env = restartedEnvironment(t, env)

	after, err := env.ContainerRegistryClient().GetRegistryInstanceCredentials(ctx, credsReq)
	require.NoError(t, err)
	require.Equal(t, before.GetPassword(), after.GetPassword())

	dropRegistryInstanceAndAwait(t, env, serviceName)
}
