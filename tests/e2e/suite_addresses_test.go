//go:build e2e_full

package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	addressesCarryInstanceName   = "e2e-addr-carry"
	addressesRemovedInstanceName = "e2e-addr-gone"
	addressesRebuildInstanceName = "e2e-addr-rebuild"
)

func Test_Addresses_S3WebUiSidecarResourceCarriesDockerAddress(t *testing.T) {
	t.Parallel()

	requireDindLoopbackBridge(t)

	env := singleNodePlane().NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	removeS3Instance(dockerClient, addressesCarryInstanceName)
	t.Cleanup(func() { removeS3Instance(dockerClient, addressesCarryInstanceName) })

	createS3Instance(t, env, addressesCarryInstanceName, true)

	serviceName := domain.S3ServiceName(addressesCarryInstanceName)
	webUiPort := publishedHostPort(t, dockerClient, serviceName, domain.S3WebUiContainerPort)

	resource := waitForWebUiResource(t, env, serviceName)

	requireSingleDockerAddress(t, resource, webUiPort)
	require.Equal(t, webUiPort, resource.GetWebUiPort(), "legacy web ui port must match the address port")

	dropS3Instance(t, env, addressesCarryInstanceName)
}

func Test_Addresses_RemovedWithInstance(t *testing.T) {
	t.Parallel()

	requireDindLoopbackBridge(t)

	env := singleNodePlane().NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	removeS3Instance(dockerClient, addressesRemovedInstanceName)
	t.Cleanup(func() { removeS3Instance(dockerClient, addressesRemovedInstanceName) })

	createS3Instance(t, env, addressesRemovedInstanceName, true)

	serviceName := domain.S3ServiceName(addressesRemovedInstanceName)

	waitForWebUiResource(t, env, serviceName)

	dropS3Instance(t, env, addressesRemovedInstanceName)

	requireNoContainersOfService(t, dockerClient, serviceName)

	resources := getServiceResources(t, env, serviceName)
	require.Nil(t, findBoundResource(resources, s3WebUiResourceType), "web ui resource must be gone after drop")
}

func Test_Addresses_RebuildKeepsAddresses(t *testing.T) {
	t.Parallel()

	requireDindLoopbackBridge(t)

	env := singleNodePlane().NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	removeS3Instance(dockerClient, addressesRebuildInstanceName)
	t.Cleanup(func() { removeS3Instance(dockerClient, addressesRebuildInstanceName) })

	createS3Instance(t, env, addressesRebuildInstanceName, true)

	serviceName := domain.S3ServiceName(addressesRebuildInstanceName)
	webUiPort := publishedHostPort(t, dockerClient, serviceName, domain.S3WebUiContainerPort)

	before := waitForWebUiResource(t, env, serviceName)
	addressBefore := requireSingleDockerAddress(t, before, webUiPort)

	rebuildAddresses(t, env)

	after := findBoundResource(getServiceResources(t, env, serviceName), s3WebUiResourceType)
	require.NotNil(t, after, "web ui resource must survive a rebuild")

	addressAfter := requireSingleDockerAddress(t, after, webUiPort)
	require.Equal(t, addressBefore.GetHost(), addressAfter.GetHost())

	dropS3Instance(t, env, addressesRebuildInstanceName)
}

func Test_Addresses_RebuildWhileRunningIsRejected(t *testing.T) {
	t.Parallel()

	env := singleNodePlane().NewEnvironment(t)
	req := newRebuildAddressesRequest()

	_, err := env.Custom.SettingsApiImpl.RebuildAddresses(t.Context(), req)
	require.NoError(t, err)

	_, err = env.Custom.SettingsApiImpl.RebuildAddresses(t.Context(), req)
	if err != nil {
		require.ErrorIs(t, err, user_errors.ErrAddressesRebuildInProgress)
	}

	idle := waitForAddressesRebuildIdle(t, env)
	require.False(t, idle.GetIsRunning())
	require.Positive(t, idle.GetTotalSteps())
	require.Equal(t, idle.GetTotalSteps(), idle.GetDoneSteps())
}
