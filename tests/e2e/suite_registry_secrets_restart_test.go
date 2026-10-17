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
	require.Equal(t, before.GetUsername(), after.GetUsername())
	require.Equal(t, before.GetPassword(), after.GetPassword())
	require.NotEmpty(t, before.GetRegistryUrl())
	require.Equal(t, before.GetRegistryUrl(), after.GetRegistryUrl())

	listed, err := env.ControlPlaneClient().ListRegistries(ctx, newListRegistriesRequest())
	require.NoError(t, err)
	require.Contains(t, registryNames(listed.GetRegistries()), serviceName)

	dropRegistryInstanceAndAwait(t, env, serviceName)
}

func newListRegistriesRequest() *velez_api.ListRegistries_Request {
	return &velez_api.ListRegistries_Request{}
}

func registryNames(registries []*velez_api.Registry) []string {
	names := make([]string, 0, len(registries))
	for _, reg := range registries {
		names = append(names, reg.GetName())
	}

	return names
}
