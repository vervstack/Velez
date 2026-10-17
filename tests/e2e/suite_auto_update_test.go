//go:build e2e_full

package e2e

import (
	"context"
	"io"
	"strconv"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	autoUpdateRegistryName = "e2e_autoupd_registry"
	autoUpdateRegistryPort = dindPortBandEnd - 2

	autoUpdateMoveTimeout     = 90 * time.Second
	autoUpdateMoveTick        = time.Second
	autoUpdateNoMoveWindow    = 8 * time.Second
	autoUpdateCheckPeriodSecs = 2

	autoUpdateRepoPrefix = "e2e/autoupd_"
)

// AutoUpdateSuite shares one registry across its cases (the DinD band has a
// single free port left); every case uses its own repo and service name, so
// the node-wide workers of parallel environments never see each other's data.
// The pushed image is registry:2 itself - unlike alpine it keeps running
// after upgrade_smerd recreates it without a command.
type AutoUpdateSuite struct {
	suite.Suite

	plane        Plane
	dockerClient client.APIClient
}

func (s *AutoUpdateSuite) SetupSuite() {
	t := s.T()

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	require.NoError(t, err)

	t.Cleanup(func() { _ = dockerClient.Close() })

	s.dockerClient = dockerClient

	pullRegistryImage(t, dockerClient)
	startAutoUpdateRegistry(t, dockerClient)
}

func (s *AutoUpdateSuite) Test_AutoUpdate_Semver_PicksHighestNumericTag() {
	t := s.T()
	env := s.plane.NewEnvironment(t, WithAutoUpdate())

	repo := autoUpdateRepoPrefix + "semver"
	serviceName := "e2e_autoupd_semver"

	s.pushTags(t, repo, "1.0.0", "1.9.0", "1.10.0")

	s.createService(t, env, newAutoUpdateCreateRequest(serviceName, autoUpdateImage(repo, "1.0.0"), true))

	s.requireMovesTo(t, serviceName, autoUpdateImage(repo, "1.10.0"))
}

func (s *AutoUpdateSuite) Test_AutoUpdate_BumpLevel_Major_SkipsMinorAndPatch() {
	t := s.T()
	env := s.plane.NewEnvironment(t, WithAutoUpdate())

	repo := autoUpdateRepoPrefix + "major"
	serviceName := "e2e_autoupd_major"

	s.pushTags(t, repo, "1.0.0", "1.0.1", "1.1.0")

	s.createService(t, env, newAutoUpdateCreateRequest(serviceName, autoUpdateImage(repo, "1.0.0"), false))
	s.setAutoUpdate(t, env, newSetAutoUpdateRequest(serviceName, true, velez_api.AutoUpdateBumpLevel_MAJOR))

	s.requireStaysOn(t, serviceName, autoUpdateImage(repo, "1.0.0"))

	s.pushTags(t, repo, "2.0.0")

	s.requireMovesTo(t, serviceName, autoUpdateImage(repo, "2.0.0"))
}

func (s *AutoUpdateSuite) Test_AutoUpdate_BumpLevel_Minor_SkipsPatch() {
	t := s.T()
	env := s.plane.NewEnvironment(t, WithAutoUpdate())

	repo := autoUpdateRepoPrefix + "minor"
	serviceName := "e2e_autoupd_minor"

	s.pushTags(t, repo, "1.0.0", "1.0.1")

	s.createService(t, env, newAutoUpdateCreateRequest(serviceName, autoUpdateImage(repo, "1.0.0"), false))
	s.setAutoUpdate(t, env, newSetAutoUpdateRequest(serviceName, true, velez_api.AutoUpdateBumpLevel_MINOR))

	s.requireStaysOn(t, serviceName, autoUpdateImage(repo, "1.0.0"))

	s.pushTags(t, repo, "1.1.0")

	s.requireMovesTo(t, serviceName, autoUpdateImage(repo, "1.1.0"))
}

func (s *AutoUpdateSuite) Test_AutoUpdate_Set_EnableAndDisable() {
	t := s.T()
	env := s.plane.NewEnvironment(t, WithAutoUpdate())

	repo := autoUpdateRepoPrefix + "toggle"
	serviceName := "e2e_autoupd_toggle"

	s.pushTags(t, repo, "1.0.0")

	s.createService(t, env, newAutoUpdateCreateRequest(serviceName, autoUpdateImage(repo, "1.0.0"), false))

	s.setAutoUpdate(t, env, newSetAutoUpdateRequest(serviceName, true, velez_api.AutoUpdateBumpLevel_PATCH))

	enabled := s.listAutoUpdateService(t, env, serviceName)
	require.True(t, enabled.GetIsEnabled())
	require.Equal(t, velez_api.AutoUpdateBumpLevel_PATCH, enabled.GetBumpLevel())
	s.requireAutoUpdateLabels(t, serviceName, labelValueTrue, domain.BumpLevelLabelValue(velez_api.AutoUpdateBumpLevel_PATCH))

	s.requireWorkerSettings(t, env)

	s.setAutoUpdate(t, env, newSetAutoUpdateRequest(serviceName, false, velez_api.AutoUpdateBumpLevel_PATCH))

	disabled := s.listAutoUpdateService(t, env, serviceName)
	require.False(t, disabled.GetIsEnabled())

	s.pushTags(t, repo, "1.0.1")

	s.requireStaysOn(t, serviceName, autoUpdateImage(repo, "1.0.0"))
}

func (s *AutoUpdateSuite) Test_AutoUpdate_Legacy_LabelOnlyUsesLastTag() {
	t := s.T()
	env := s.plane.NewEnvironment(t, WithAutoUpdate())

	repo := autoUpdateRepoPrefix + "legacy"
	serviceName := "e2e_autoupd_legacy"

	s.pushTags(t, repo, "legacy", "zzz-last")

	lastTag := lastRegistryTag(t, repo)
	require.NotEqual(t, "legacy", lastTag, "registry must list another tag after legacy")

	s.createService(t, env, newAutoUpdateCreateRequest(serviceName, autoUpdateImage(repo, "legacy"), true))

	s.requireMovesTo(t, serviceName, autoUpdateImage(repo, lastTag))
}

func (s *AutoUpdateSuite) Test_AutoUpdate_ListPlugins_ContainsAutoUpdate() {
	t := s.T()
	env := s.plane.NewEnvironment(t, WithAutoUpdate())

	resp, err := env.ControlPlaneClient().ListPlugins(t.Context(), newListPluginsRequest())
	require.NoError(t, err)

	var autoUpdate *velez_api.Plugin

	for _, plugin := range resp.GetPlugins() {
		if plugin.GetType() == velez_api.VervPluginType_auto_update {
			autoUpdate = plugin

			break
		}
	}

	require.NotNil(t, autoUpdate, "expected auto_update plugin in ListPlugins response")
	require.Equal(t, velez_api.VervPlugin_running, autoUpdate.GetState())
}

func Test_AutoUpdate(t *testing.T) {
	t.Parallel()

	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &AutoUpdateSuite{plane: plane}
	})
}

func newAutoUpdateCreateRequest(serviceName, imageName string, isAutoUpgrade bool) *velez_api.CreateSmerd_Request {
	return &velez_api.CreateSmerd_Request{
		Name:         serviceName,
		ImageName:    imageName,
		IgnoreConfig: true,
		AutoUpgrade:  isAutoUpgrade,
	}
}

func newSetAutoUpdateRequest(
	serviceName string, isEnabled bool, level velez_api.AutoUpdateBumpLevel,
) *velez_api.SetServiceAutoUpdate_Request {
	return &velez_api.SetServiceAutoUpdate_Request{
		Name:      serviceName,
		IsEnabled: isEnabled,
		BumpLevel: level,
	}
}

func newListAutoUpdateServicesRequest() *velez_api.ListAutoUpdateServices_Request {
	return &velez_api.ListAutoUpdateServices_Request{}
}

func newListPluginsRequest() *velez_api.ListPlugins_Request {
	return &velez_api.ListPlugins_Request{}
}

func autoUpdateRegistryHost() string {
	return "localhost:" + strconv.Itoa(autoUpdateRegistryPort)
}

func autoUpdateImage(repo, tag string) string {
	return autoUpdateRegistryHost() + "/" + repo + ":" + tag
}

func startAutoUpdateRegistry(t *testing.T, dockerClient client.APIClient) {
	t.Helper()

	cfg := &container.Config{Image: registryRegImage}

	binding := nat.PortBinding{HostIP: pgRegAllInterfaces, HostPort: strconv.Itoa(autoUpdateRegistryPort)}
	hostCfg := &container.HostConfig{
		PortBindings: nat.PortMap{registryRegContainerPort: []nat.PortBinding{binding}},
	}

	t.Cleanup(func() {
		removeOpts := container.RemoveOptions{Force: true, RemoveVolumes: true}

		_ = dockerClient.ContainerRemove(context.Background(), autoUpdateRegistryName, removeOpts)
	})

	created, err := dockerClient.ContainerCreate(t.Context(), cfg, hostCfg, nil, nil, autoUpdateRegistryName)
	require.NoError(t, err)

	err = dockerClient.ContainerStart(t.Context(), created.ID, container.StartOptions{})
	require.NoError(t, err)

	forwardDindPort(t, autoUpdateRegistryPort)
	awaitRegistryReady(t, autoUpdateRegistryPort)
}

func (s *AutoUpdateSuite) pushTags(t *testing.T, repo string, tags ...string) {
	t.Helper()

	for _, tag := range tags {
		ref := autoUpdateImage(repo, tag)

		err := s.dockerClient.ImageTag(t.Context(), registryRegImage, ref)
		require.NoError(t, err)

		pushOpts := newPushOptions(t, "", "")

		reader, err := s.dockerClient.ImagePush(t.Context(), ref, pushOpts)
		require.NoError(t, err)

		output, err := io.ReadAll(reader)
		require.NoError(t, err)

		err = reader.Close()
		require.NoError(t, err)

		require.NotContains(t, string(output), `"errorDetail"`, "docker push of %s failed: %s", ref, string(output))
	}
}

func lastRegistryTag(t *testing.T, repo string) string {
	t.Helper()

	registryRepo, err := name.NewRepository(autoUpdateRegistryHost() + "/" + repo)
	require.NoError(t, err)

	tags, err := remote.List(registryRepo, remote.WithContext(t.Context()), remote.WithAuth(authn.Anonymous))
	require.NoError(t, err)
	require.NotEmpty(t, tags)

	return tags[len(tags)-1]
}

func (s *AutoUpdateSuite) createService(
	t *testing.T, env *TestEnvironment, req *velez_api.CreateSmerd_Request,
) {
	t.Helper()

	created := env.CreateSmerd(t, req)
	require.Equal(t, velez_api.Smerd_running, created.GetStatus())
}

func (s *AutoUpdateSuite) setAutoUpdate(
	t *testing.T, env *TestEnvironment, req *velez_api.SetServiceAutoUpdate_Request,
) {
	t.Helper()

	resp, err := env.AutoUpdateClient().SetServiceAutoUpdate(t.Context(), req)
	require.NoError(t, err)

	task := env.AwaitTask(t, resp.GetEntityId(), resp.GetAction())
	require.Equal(t, velez_api.TaskStatus_DONE, task.GetStatus(), "set auto-update task error: %s", task.GetError())
}

func (s *AutoUpdateSuite) listAutoUpdateService(
	t *testing.T, env *TestEnvironment, serviceName string,
) *velez_api.ListAutoUpdateServices_Response_Service {
	t.Helper()

	resp, err := env.AutoUpdateClient().ListAutoUpdateServices(t.Context(), newListAutoUpdateServicesRequest())
	require.NoError(t, err)

	for _, service := range resp.GetServices() {
		if service.GetName() == serviceName {
			return service
		}
	}

	require.FailNow(t, "service missing from ListAutoUpdateServices", serviceName)

	return nil
}

func (s *AutoUpdateSuite) requireWorkerSettings(t *testing.T, env *TestEnvironment) {
	t.Helper()

	resp, err := env.AutoUpdateClient().ListAutoUpdateServices(t.Context(), newListAutoUpdateServicesRequest())
	require.NoError(t, err)

	require.True(t, resp.GetIsWorkerEnabled())
	require.EqualValues(t, autoUpdateCheckPeriodSecs, resp.GetCheckPeriodSeconds())
}

func (s *AutoUpdateSuite) requireAutoUpdateLabels(t *testing.T, serviceName, wantEnabled, wantBumpLevel string) {
	t.Helper()

	inspected, err := s.dockerClient.ContainerInspect(t.Context(), serviceName)
	require.NoError(t, err)

	require.Equal(t, wantEnabled, inspected.Config.Labels[labels.AutoUpgrade])
	require.Equal(t, wantBumpLevel, inspected.Config.Labels[labels.AutoUpgradeBumpLevel])
}

// currentImage is empty while the container is being recreated.
func (s *AutoUpdateSuite) currentImage(serviceName string) string {
	inspected, err := s.dockerClient.ContainerInspect(context.Background(), serviceName)
	if err != nil {
		return ""
	}

	return inspected.Config.Image
}

func (s *AutoUpdateSuite) requireMovesTo(t *testing.T, serviceName, wantImage string) {
	t.Helper()

	require.Eventually(t, func() bool {
		return s.currentImage(serviceName) == wantImage
	}, autoUpdateMoveTimeout, autoUpdateMoveTick, "%s never moved to %s", serviceName, wantImage)
}

func (s *AutoUpdateSuite) requireStaysOn(t *testing.T, serviceName, wantImage string) {
	t.Helper()

	require.Equal(t, wantImage, s.currentImage(serviceName))

	require.Never(t, func() bool {
		current := s.currentImage(serviceName)

		return current != "" && current != wantImage
	}, autoUpdateNoMoveWindow, autoUpdateMoveTick, "%s moved off %s", serviceName, wantImage)
}
