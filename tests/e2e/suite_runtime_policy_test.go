//go:build e2e_full

package e2e

import (
	"context"
	"fmt"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	policyWhitelistedImage = "tailscale/tailscale:e2e-policy"
	policyNonWhitelisted   = NginxAlpineImage
	policyNetAdminCap      = "NET_ADMIN"
	policyStatefullSuffix  = ""
)

type policyCreator struct {
	name   string
	create func(t *testing.T, env *TestEnvironment, name string, cfg *container.Config, hostCfg *container.HostConfig,
	) (container.CreateResponse, error)
}

func policyCreators() []policyCreator {
	return []policyCreator{
		{name: "runtime from resolver", create: createViaResolverRuntime},
		{name: "node docker client", create: createViaNodeDocker},
	}
}

func newPolicySettingsCases() []domain.Settings {
	return []domain.Settings{
		{IsSysboxEnabled: false, IsSysboxWhitelistIgnored: false},
		{IsSysboxEnabled: true, IsSysboxWhitelistIgnored: false},
		{IsSysboxEnabled: true, IsSysboxWhitelistIgnored: true},
	}
}

func newPolicyContainerConfig(imageName string) *container.Config {
	return &container.Config{
		Image:  imageName,
		Labels: map[string]string{},
	}
}

func newPlainHostConfig() *container.HostConfig {
	return &container.HostConfig{}
}

func newPrivilegedHostConfig() *container.HostConfig {
	return &container.HostConfig{Privileged: true}
}

func newCapAddHostConfig() *container.HostConfig {
	return &container.HostConfig{CapAdd: []string{policyNetAdminCap}}
}

func newHostNetworkHostConfig() *container.HostConfig {
	return &container.HostConfig{NetworkMode: container.NetworkMode("host")}
}

func newDockerSocketBindHostConfig() *container.HostConfig {
	return &container.HostConfig{Binds: []string{dindDockerSockPath + ":" + dindDockerSockPath}}
}

func newDockerSocketMountHostConfig() *container.HostConfig {
	socketMount := mount.Mount{
		Type:   mount.TypeBind,
		Source: dindDockerSockPath,
		Target: dindDockerSockPath,
	}

	return &container.HostConfig{Mounts: []mount.Mount{socketMount}}
}

func newElevatedHostConfigs() map[string]func() *container.HostConfig {
	return map[string]func() *container.HostConfig{
		"privileged":         newPrivilegedHostConfig,
		"cap_add":            newCapAddHostConfig,
		"host network":       newHostNetworkHostConfig,
		"docker.sock bind":   newDockerSocketBindHostConfig,
		"docker.sock mounts": newDockerSocketMountHostConfig,
	}
}

type settingsRpcProvider struct {
	client velez_api.SettingsAPIClient
}

func (p *settingsRpcProvider) GetSettings(ctx context.Context) (domain.Settings, error) {
	req := &velez_api.GetSettings_Request{}

	resp, err := p.client.GetSettings(ctx, req)
	if err != nil {
		return domain.Settings{}, rerrors.Wrap(err, "error getting settings")
	}

	settings := domain.Settings{
		IsSysboxEnabled:          resp.GetSettings().GetIsSysboxEnabled(),
		IsSysboxWhitelistIgnored: resp.GetSettings().GetIsSysboxWhitelistIgnored(),
	}

	return settings, nil
}

func newSettingsProvider(env *TestEnvironment) container_runtime.SettingsProvider {
	return &settingsRpcProvider{client: env.SettingsClient()}
}

func newProductionRuntime(t *testing.T, env *TestEnvironment) container_runtime.ContainerRuntime {
	t.Helper()

	dockerClient := env.DockerClient()

	runtimes := container_runtime.NewResolver(
		dockerClient, dockerClient.DaemonHost(), env.Cfg.Environment.CustomLabels, nil,
		container_runtime.WithSettings(newSettingsProvider(env)))

	runtime, err := runtimes.Runtime(t.Context(), "")
	require.NoError(t, err)

	return runtime
}

func newProductionNodeDocker(t *testing.T, env *TestEnvironment) *docker.Docker {
	t.Helper()

	hostOpt := client.WithHost(env.DockerClient().DaemonHost())
	versionOpt := client.WithAPIVersionNegotiation()

	nodeDocker, err := docker.NewClientWithOpts(env.Cfg.Environment.CustomLabels, hostOpt, versionOpt)
	require.NoError(t, err)

	nodeDocker.SetSettingsProvider(newSettingsProvider(env))

	return nodeDocker
}

func createViaResolverRuntime(
	t *testing.T, env *TestEnvironment, name string, cfg *container.Config, hostCfg *container.HostConfig,
) (container.CreateResponse, error) {
	t.Helper()

	runtime := newProductionRuntime(t, env)

	createReq := container_runtime.ContainerCreateRequest{
		Config:        &container_runtime.ContainerConfig{Config: cfg},
		HostConfig:    &container_runtime.HostConfig{HostConfig: hostCfg},
		ContainerName: name,
	}

	return runtime.ContainerCreate(t.Context(), createReq)
}

func createViaNodeDocker(
	t *testing.T, env *TestEnvironment, name string, cfg *container.Config, hostCfg *container.HostConfig,
) (container.CreateResponse, error) {
	t.Helper()

	nodeDocker := newProductionNodeDocker(t, env)

	return nodeDocker.ContainerCreate(t.Context(), cfg, hostCfg, nil, nil, name, "")
}

// The SettingsAPI refuses to enable sysbox on a daemon without the runtime
// registered, while these tests need exactly that state to assert the daemon
// rejects the runtime name, so the settings row is written straight to storage.
func setNodeSettings(t *testing.T, env *TestEnvironment, settings domain.Settings) {
	t.Helper()

	stateManager := env.Custom.ClusterClients.StateManager()

	_, err := stateManager.Settings().UpdateSettings(t.Context(), settings)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = stateManager.Settings().UpdateSettings(context.Background(), domain.Settings{})
	})
}

func ensureImageTagged(t *testing.T, env *TestEnvironment, target string) {
	t.Helper()

	dockerClient := env.DockerClient()

	ensureImagePresent(t, dockerClient, NginxAlpineImage)

	err := dockerClient.ImageTag(t.Context(), NginxAlpineImage, target)
	require.NoError(t, err)
}

func removeCreatedOnCleanup(t *testing.T, env *TestEnvironment, createResp container.CreateResponse) {
	t.Helper()

	dockerClient := env.DockerClient()

	t.Cleanup(func() {
		removeOpts := container.RemoveOptions{Force: true}
		_ = dockerClient.ContainerRemove(context.Background(), createResp.ID, removeOpts)
	})
}

func requireCreatedOnDefaultRuntime(t *testing.T, env *TestEnvironment, createResp container.CreateResponse, err error) {
	t.Helper()

	require.NoError(t, err)
	removeCreatedOnCleanup(t, env, createResp)

	dockerClient := env.DockerClient()

	inspected, inspectErr := dockerClient.ContainerInspect(t.Context(), createResp.ID)
	require.NoError(t, inspectErr)
	require.NotEqual(t, sysboxRuntimeName, inspected.HostConfig.Runtime)
}

// requireSysboxRuntimeRequested proves the policy stamped HostConfig.Runtime:
// with the runtime installed the container carries it, without it the daemon
// rejects the unknown runtime name at create time.
func requireSysboxRuntimeRequested(t *testing.T, env *TestEnvironment, createResp container.CreateResponse, err error) {
	t.Helper()

	if !isSysboxRuntimeRegistered(t, env) {
		require.Error(t, err)
		require.NotErrorIs(t, err, user_errors.ErrElevatedAccessImageNotWhitelisted)
		require.ErrorContains(t, err, sysboxRuntimeName)

		return
	}

	require.NoError(t, err)
	removeCreatedOnCleanup(t, env, createResp)

	dockerClient := env.DockerClient()

	inspected, inspectErr := dockerClient.ContainerInspect(t.Context(), createResp.ID)
	require.NoError(t, inspectErr)
	require.Equal(t, sysboxRuntimeName, inspected.HostConfig.Runtime)
}

type policyCheck func(t *testing.T, env *TestEnvironment, createResp container.CreateResponse, err error)

func settingsCaseName(settings domain.Settings) string {
	return fmt.Sprintf("sysbox_%t_ignore_%t", settings.IsSysboxEnabled, settings.IsSysboxWhitelistIgnored)
}

func runElevatedMatrix(
	t *testing.T, env *TestEnvironment, settingsCases []domain.Settings, imageName string, check policyCheck,
) {
	t.Helper()

	for _, settings := range settingsCases {
		t.Run(settingsCaseName(settings), func(t *testing.T) {
			setNodeSettings(t, env, settings)

			runElevatedCreators(t, env, imageName, check)
		})
	}
}

func runElevatedCreators(t *testing.T, env *TestEnvironment, imageName string, check policyCheck) {
	t.Helper()

	for _, creator := range policyCreators() {
		t.Run(creator.name, func(t *testing.T) {
			runElevatedHostConfigs(t, env, creator, imageName, check)
		})
	}
}

func runElevatedHostConfigs(
	t *testing.T, env *TestEnvironment, creator policyCreator, imageName string, check policyCheck,
) {
	t.Helper()

	for elevationName, newHostConfig := range newElevatedHostConfigs() {
		t.Run(elevationName, func(t *testing.T) {
			cfg := newPolicyContainerConfig(imageName)

			createResp, err := creator.create(t, env, GetServiceName(t), cfg, newHostConfig())

			check(t, env, createResp, err)
		})
	}
}

func requireRejectedAsNotWhitelisted(
	t *testing.T, env *TestEnvironment, createResp container.CreateResponse, err error,
) {
	t.Helper()

	if err == nil {
		removeCreatedOnCleanup(t, env, createResp)
	}

	require.ErrorIs(t, err, user_errors.ErrElevatedAccessImageNotWhitelisted)
}

func Test_RuntimePolicy_ElevatedNonWhitelistedImage_IsRejected(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)

	runElevatedMatrix(t, env, newPolicySettingsCases(), policyNonWhitelisted, requireRejectedAsNotWhitelisted)
}

func Test_RuntimePolicy_ElevatedWhitelistedImage_StaysOnDefaultRuntime(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)
	ensureImageTagged(t, env, policyWhitelistedImage)

	settingsCases := []domain.Settings{
		{IsSysboxEnabled: false, IsSysboxWhitelistIgnored: false},
		{IsSysboxEnabled: false, IsSysboxWhitelistIgnored: true},
		{IsSysboxEnabled: true, IsSysboxWhitelistIgnored: false},
	}

	runElevatedMatrix(t, env, settingsCases, policyWhitelistedImage, requireCreatedOnDefaultRuntime)
}

func Test_RuntimePolicy_SysboxOn_PlainContainerRequestsSysboxRuntime(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)
	setNodeSettings(t, env, domain.Settings{IsSysboxEnabled: true})

	for _, creator := range policyCreators() {
		t.Run(creator.name, func(t *testing.T) {
			ensureImagePresent(t, env.DockerClient(), NginxAlpineImage)

			cfg := newPolicyContainerConfig(NginxAlpineImage)

			createResp, createErr := creator.create(t, env, GetServiceName(t), cfg, newPlainHostConfig())
			requireSysboxRuntimeRequested(t, env, createResp, createErr)
		})
	}
}

func Test_RuntimePolicy_SysboxOff_PlainContainerStaysOnDefaultRuntime(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)

	for _, creator := range policyCreators() {
		t.Run(creator.name, func(t *testing.T) {
			ensureImagePresent(t, env.DockerClient(), NginxAlpineImage)

			cfg := newPolicyContainerConfig(NginxAlpineImage)

			createResp, createErr := creator.create(t, env, GetServiceName(t), cfg, newPlainHostConfig())
			requireCreatedOnDefaultRuntime(t, env, createResp, createErr)
		})
	}
}

func Test_RuntimePolicy_WhitelistIgnored_ElevatedWhitelistedImageRequestsSysboxRuntime(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)
	ensureImageTagged(t, env, policyWhitelistedImage)

	for _, creator := range policyCreators() {
		t.Run(creator.name+"/ignore on", func(t *testing.T) {
			setNodeSettings(t, env, domain.Settings{IsSysboxEnabled: true, IsSysboxWhitelistIgnored: true})

			cfg := newPolicyContainerConfig(policyWhitelistedImage)

			createResp, err := creator.create(t, env, GetServiceName(t), cfg, newCapAddHostConfig())
			requireSysboxRuntimeRequested(t, env, createResp, err)
		})

		t.Run(creator.name+"/ignore off", func(t *testing.T) {
			setNodeSettings(t, env, domain.Settings{IsSysboxEnabled: true, IsSysboxWhitelistIgnored: false})

			cfg := newPolicyContainerConfig(policyWhitelistedImage)

			createResp, err := creator.create(t, env, GetServiceName(t), cfg, newCapAddHostConfig())
			requireCreatedOnDefaultRuntime(t, env, createResp, err)
		})
	}
}

func newEnableStatefullRequest(env *TestEnvironment) *velez_api.EnablePlugin_Request {
	statefullReq := &velez_api.EnableStatefullCluster{
		IsExposePort: toolbox.ToPtr(true),
		ExposeToPort: toolbox.ToPtr(env.clusterPgPort),
	}
	payload := &velez_api.EnablePlugin_Request_StatefullCluster{
		StatefullCluster: statefullReq,
	}

	return &velez_api.EnablePlugin_Request{
		Plugin:  velez_api.VervPluginType_statefull_pg,
		Payload: payload,
	}
}

func Test_RuntimePolicy_SysboxOn_StatefullPgJob_UsesSysboxRuntime(t *testing.T) {
	t.Parallel()

	env, pgName := newStatefullEnvironment(t, Planes[1], policyStatefullSuffix)
	setNodeSettings(t, env, domain.Settings{IsSysboxEnabled: true})

	resp, err := env.ControlPlaneClient().EnablePlugin(t.Context(), newEnableStatefullRequest(env))
	require.NoError(t, err)

	finalTask := env.AwaitTask(t, resp.GetEntityId(), resp.GetAction())

	if !isSysboxRuntimeRegistered(t, env) {
		require.Equal(t, velez_api.TaskStatus_FAILED, finalTask.GetStatus())
		require.Contains(t, finalTask.GetError(), sysboxRuntimeName)

		return
	}

	require.Equal(t, velez_api.TaskStatus_DONE, finalTask.GetStatus(),
		"enable statefull task error: %s", finalTask.GetError())

	dockerClient := env.DockerClient()

	inspected, inspectErr := dockerClient.ContainerInspect(t.Context(), pgName)
	require.NoError(t, inspectErr)
	require.Equal(t, sysboxRuntimeName, inspected.HostConfig.Runtime)
}
