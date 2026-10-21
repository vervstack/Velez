//go:build e2e_full

package e2e

import (
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/runtime_policy"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage/environments"
)

const (
	accessUpgradeImage = "nginx:e2e-access-upgrade"
	accessClusterSfx   = "e2e-reg-access"

	keptBindOptions     = "ro,private"
	keptBindPropagation = mount.PropagationPrivate

	systemRunPath    = "/var/run"
	systemSocketPath = "/tmp/e2e_reg_syssock/e2e.sock"
	systemTargetDir  = "/e2e-system"
)

func newKeptBindHostConfig(hostDir string) *container.HostConfig {
	return &container.HostConfig{Binds: []string{hostDir + ":" + linkTargetDir + ":" + keptBindOptions}}
}

func newSystemBindHostConfig(source string) *container.HostConfig {
	return &container.HostConfig{Binds: []string{source + ":" + systemTargetDir}}
}

func newBindLinkRegisterRequest(
	containerId, serviceName string, link *velez_api.RegisterContainer_Request_BindMountLink,
) *velez_api.RegisterContainer_Request {
	req := newRegisterContainerRequest(containerId, serviceName)
	req.BindMountLinks = []*velez_api.RegisterContainer_Request_BindMountLink{link}

	return req
}

func newKeptBindLink(source string) *velez_api.RegisterContainer_Request_BindMountLink {
	return &velez_api.RegisterContainer_Request_BindMountLink{Source: source, IsKeptAsBind: true}
}

func newKeptBindLinkWithVolumeName(source, volumeName string) *velez_api.RegisterContainer_Request_BindMountLink {
	link := newKeptBindLink(source)
	link.VolumeName = &volumeName

	return link
}

func newLinkedBindLink(source string) *velez_api.RegisterContainer_Request_BindMountLink {
	return &velez_api.RegisterContainer_Request_BindMountLink{Source: source}
}

func newElevatedRegisterRequest(containerId, serviceName string, isAllowed bool) *velez_api.RegisterContainer_Request {
	req := newRegisterContainerRequest(containerId, serviceName)
	req.IsElevatedAccessAllowed = isAllowed

	return req
}

func newGetContainerRequest(containerId string) *velez_api.GetContainer_Request {
	return &velez_api.GetContainer_Request{
		Id:          containerId,
		Environment: environments.DefaultEnvironmentName,
	}
}

func newUpgradeToImageRequest(containerName, image string) *velez_api.UpgradeSmerd_Request {
	return &velez_api.UpgradeSmerd_Request{
		Name:        containerName,
		Image:       image,
		Environment: environments.DefaultEnvironmentName,
	}
}

func newCreateSmerdWithBindMountRequest(name string) *velez_api.CreateSmerd_Request {
	bind := &velez_api.BindMount{Source: "/tmp", Target: linkTargetDir}
	settings := &velez_api.Container_Settings{BindMounts: []*velez_api.BindMount{bind}}

	return &velez_api.CreateSmerd_Request{
		Name:         name,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Settings:     settings,
	}
}

func newCreateSmerdWithElevatedAccessRequest(name string) *velez_api.CreateSmerd_Request {
	access := &velez_api.ElevatedAccess{CapAdd: []string{policyNetAdminCap}}
	settings := &velez_api.Container_Settings{ElevatedAccess: access}

	return &velez_api.CreateSmerd_Request{
		Name:         name,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Settings:     settings,
	}
}

func newCreateSmerdWithAllowanceLabelRequest(name string) *velez_api.CreateSmerd_Request {
	allowance := map[string]string{labels.ElevatedAccessLabel: labels.ElevatedAccessAllowedValue}

	return &velez_api.CreateSmerd_Request{
		Name:         name,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Labels:       allowance,
	}
}

// bareCapNames drops the CAP_ prefix some daemons add when they report capabilities.
func bareCapNames(caps []string) []string {
	bare := make([]string, 0, len(caps))

	for _, capName := range caps {
		bare = append(bare, strings.TrimPrefix(capName, "CAP_"))
	}

	return bare
}

func requireKeptBindMount(t *testing.T, inspected container.InspectResponse, source, target string) {
	t.Helper()

	for _, mnt := range inspected.Mounts {
		if mnt.Destination != target {
			continue
		}

		require.Equal(t, mount.TypeBind, mnt.Type, "mount at %s: %+v", target, mnt)
		require.Equal(t, source, mnt.Source)
		require.False(t, mnt.RW, "the read-only flag must be kept")
		require.Equal(t, keptBindPropagation, mnt.Propagation)

		return
	}

	require.Failf(t, "mount missing", "no mount at %q: %+v", target, inspected.Mounts)
}

func requireNoVolumeNamedLike(t *testing.T, dockerClient client.APIClient, namePart string) {
	t.Helper()

	nameFilter := filters.NewArgs(filters.Arg("name", namePart))

	listOpts := volume.ListOptions{Filters: nameFilter}

	listed, err := dockerClient.VolumeList(t.Context(), listOpts)
	require.NoError(t, err)
	require.Empty(t, listed.Volumes, "no volume may be created for a kept bind mount")
}

func requireRegisterFailedWith(
	t *testing.T, env *TestEnvironment, req *velez_api.RegisterContainer_Request, wantErr error,
) {
	t.Helper()

	task := requireRegisterFailed(t, env, req)
	require.Contains(t, task.GetError(), wantErr.Error())
}

func requireRefusedAndUntouched(
	t *testing.T,
	env *TestEnvironment,
	foreign foreignContainer,
	serviceName string,
	req *velez_api.RegisterContainer_Request,
	wantErr error,
) {
	t.Helper()

	requireRegisterFailedWith(t, env, req, wantErr)

	requireContainerUntouched(t, env.DockerClient(), foreign)
	requireSingleContainerNamed(t, env.DockerClient(), foreign.name)
	requireServiceNotListed(t, env, serviceName)
}

func requireElevatedAccessKept(t *testing.T, inspected container.InspectResponse, hostCfg *container.HostConfig) {
	t.Helper()

	require.True(t, inspected.State.Running)
	require.Equal(t, hostCfg.Privileged, inspected.HostConfig.Privileged)
	require.ElementsMatch(t, bareCapNames(hostCfg.CapAdd), bareCapNames(inspected.HostConfig.CapAdd))
}

func requireAllowanceLabel(t *testing.T, inspected container.InspectResponse) {
	t.Helper()

	require.Equal(t, labels.ElevatedAccessAllowedValue, inspected.Config.Labels[labels.ElevatedAccessLabel])
}

func getContainerById(t *testing.T, env *TestEnvironment, containerId string) *velez_api.DockerContainer {
	t.Helper()

	got, err := env.VelezApiClient().GetContainer(t.Context(), newGetContainerRequest(containerId))
	require.NoError(t, err)

	return got
}

func requireReportedElevatedAccess(t *testing.T, got *velez_api.DockerContainer, hostCfg *container.HostConfig) {
	t.Helper()

	access := got.GetElevatedAccess()
	require.NotNil(t, access, "GetContainer must report the elevated access")
	require.Equal(t, hostCfg.Privileged, access.GetIsPrivileged())
	require.ElementsMatch(t, bareCapNames(hostCfg.CapAdd), bareCapNames(access.GetCapAdd()))
}

func requireUpgradeKeepsAccess(
	t *testing.T, env *TestEnvironment, containerName string, hostCfg *container.HostConfig,
) {
	t.Helper()

	err := env.DockerClient().ImageTag(t.Context(), NginxAlpineImage, accessUpgradeImage)
	require.NoError(t, err)

	before, err := env.DockerClient().ContainerInspect(t.Context(), containerName)
	require.NoError(t, err)

	_, err = env.VelezApiClient().UpgradeSmerd(t.Context(), newUpgradeToImageRequest(containerName, accessUpgradeImage))
	require.NoError(t, err)

	after, err := env.DockerClient().ContainerInspect(t.Context(), containerName)
	require.NoError(t, err)
	require.NotEqual(t, before.ID, after.ID, "upgrade recreates the container")
	require.Equal(t, accessUpgradeImage, after.Config.Image)

	requireElevatedAccessKept(t, after, hostCfg)
	requireAllowanceLabel(t, after)
}

func listWhitelistedImages(t *testing.T, env *TestEnvironment) []*velez_api.WhitelistedImage {
	t.Helper()

	resp, err := env.SettingsClient().ListWhitelistedImages(t.Context(), &velez_api.ListWhitelistedImages_Request{})
	require.NoError(t, err)

	return resp.GetImages()
}

func findWhitelistedImage(t *testing.T, env *TestEnvironment, repository string) *velez_api.WhitelistedImage {
	t.Helper()

	for _, image := range listWhitelistedImages(t, env) {
		if !image.GetIsBuiltin() && image.GetRepository() == repository {
			return image
		}
	}

	return nil
}

func removeWhitelistedImage(t *testing.T, env *TestEnvironment, imageId string) {
	t.Helper()

	req := &velez_api.RemoveWhitelistedImage_Request{Id: imageId}

	_, err := env.SettingsClient().RemoveWhitelistedImage(t.Context(), req)
	require.NoError(t, err)
}

func Test_RegisterContainer_SingleMode_KeepsBindMountAsBind(t *testing.T) {
	t.Parallel()

	containerName, serviceName := registerContainerNames("bindkept")
	hostDir := linkBindDir("bindkept")

	env := Planes[0].NewEnvironment(t)
	dockerClient := env.DockerClient()

	foreign := startContainerWith(t, dockerClient, containerName, newKeptBindHostConfig(hostDir))

	requireRegisterDone(t, env, newBindLinkRegisterRequest(foreign.id, serviceName, newKeptBindLink(hostDir)))

	inspected, err := dockerClient.ContainerInspect(t.Context(), containerName)
	require.NoError(t, err)
	require.NotEqual(t, foreign.id, inspected.ID)
	require.True(t, inspected.State.Running)
	require.Equal(t, serviceName, inspected.Config.Labels[labels.VervServiceLabel])

	requireKeptBindMount(t, inspected, hostDir, linkTargetDir)
	requireNoVolumeNamedLike(t, dockerClient, serviceName)
	requireSingleContainerNamed(t, dockerClient, containerName)
}

func Test_RegisterContainer_SingleMode_RefusesBadBindLinks(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)
	dockerClient := env.DockerClient()

	t.Run("kept bind with a volume name", func(t *testing.T) {
		containerName, serviceName := registerContainerNames("keptvolume")
		hostDir := linkBindDir("keptvolume")

		foreign := startContainerWith(t, dockerClient, containerName, newKeptBindHostConfig(hostDir))

		link := newKeptBindLinkWithVolumeName(hostDir, "e2e_reg_keptvolume_named")
		req := newBindLinkRegisterRequest(foreign.id, serviceName, link)

		requireRefusedAndUntouched(t, env, foreign, serviceName, req, user_errors.ErrBindMountKeptWithVolumeName)
	})

	t.Run("system directory linked to a volume", func(t *testing.T) {
		containerName, serviceName := registerContainerNames("sysdir")

		foreign := startContainerWith(t, dockerClient, containerName, newSystemBindHostConfig(systemRunPath))

		req := newBindLinkRegisterRequest(foreign.id, serviceName, newLinkedBindLink(systemRunPath))

		requireRefusedAndUntouched(t, env, foreign, serviceName, req, user_errors.ErrBindMountSystemPathLinked)
	})

	t.Run("socket path linked to a volume", func(t *testing.T) {
		containerName, serviceName := registerContainerNames("syssock")

		foreign := startContainerWith(t, dockerClient, containerName, newSystemBindHostConfig(systemSocketPath))

		req := newBindLinkRegisterRequest(foreign.id, serviceName, newLinkedBindLink(systemSocketPath))

		requireRefusedAndUntouched(t, env, foreign, serviceName, req, user_errors.ErrBindMountSystemPathLinked)
	})
}

func Test_RegisterContainer_SingleMode_ElevatedAccessAllowed_KeepsAccessAndSurvivesUpgrade(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)
	dockerClient := env.DockerClient()

	cases := map[string]func() *container.HostConfig{
		"cap add":    newCapAddHostConfig,
		"privileged": newPrivilegedHostConfig,
	}

	for caseName, newHostConfig := range cases {
		t.Run(caseName, func(t *testing.T) {
			containerName, serviceName := registerContainerNames("elev" + dockerSafeToken(caseName))
			hostCfg := newHostConfig()

			foreign := startContainerWith(t, dockerClient, containerName, hostCfg)

			requireRegisterDone(t, env, newElevatedRegisterRequest(foreign.id, serviceName, true))

			inspected, err := dockerClient.ContainerInspect(t.Context(), containerName)
			require.NoError(t, err)
			require.NotEqual(t, foreign.id, inspected.ID)
			require.Equal(t, serviceName, inspected.Config.Labels[labels.VervServiceLabel])

			requireElevatedAccessKept(t, inspected, hostCfg)
			requireAllowanceLabel(t, inspected)

			got := getContainerById(t, env, inspected.ID)
			requireReportedElevatedAccess(t, got, hostCfg)
			require.True(t, got.GetIsImageWhitelisted())

			requireUpgradeKeepsAccess(t, env, containerName, hostCfg)
		})
	}
}

func Test_RegisterContainer_SingleMode_ElevatedAccessNotAllowed_TaskFailsUntouched(t *testing.T) {
	t.Parallel()

	containerName, serviceName := registerContainerNames("elevrefuse")

	env := Planes[0].NewEnvironment(t)

	foreign := startContainerWith(t, env.DockerClient(), containerName, newCapAddHostConfig())

	req := newElevatedRegisterRequest(foreign.id, serviceName, false)

	requireRefusedAndUntouched(t, env, foreign, serviceName, req, user_errors.ErrElevatedAccessImageNotWhitelisted)

	requireNoContainerNamed(t, env, containerName+replacedContainerSuffix)
}

func Test_RegisterContainer_SingleMode_PlainContainerHasNoElevatedAccess(t *testing.T) {
	t.Parallel()

	containerName, serviceName := registerContainerNames("elevplain")

	env := Planes[0].NewEnvironment(t)
	dockerClient := env.DockerClient()

	foreign := startForeignContainer(t, env, containerName, false)

	requireRegisterDone(t, env, newElevatedRegisterRequest(foreign.id, serviceName, true))

	inspected, err := dockerClient.ContainerInspect(t.Context(), containerName)
	require.NoError(t, err)
	require.True(t, inspected.State.Running)
	require.Empty(t, inspected.HostConfig.CapAdd)
	require.False(t, inspected.HostConfig.Privileged)

	got := getContainerById(t, env, inspected.ID)
	require.Nil(t, got.GetElevatedAccess(), "a plain container reports no elevated access")
}

func Test_CreateSmerd_RejectsReservedSettings(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)

	cases := map[string]*velez_api.CreateSmerd_Request{
		"bind mounts":     newCreateSmerdWithBindMountRequest(GetServiceName(t) + "_bind"),
		"elevated access": newCreateSmerdWithElevatedAccessRequest(GetServiceName(t) + "_access"),
		"allowance label": newCreateSmerdWithAllowanceLabelRequest(GetServiceName(t) + "_label"),
	}

	for caseName, req := range cases {
		t.Run(caseName, func(t *testing.T) {
			_, err := env.VelezApiClient().CreateSmerd(t.Context(), req)
			require.Equal(t, codes.InvalidArgument, status.Code(err))
			require.ErrorContains(t, err, user_errors.ErrReservedSettingsNotAllowed.Error())
		})
	}
}

func Test_Settings_SingleMode_ListWhitelistedImages_OnlyBuiltIns(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)

	images := listWhitelistedImages(t, env)
	require.NotEmpty(t, images)

	for _, image := range images {
		require.True(t, image.GetIsBuiltin(), "single-node lists built-ins only: %+v", image)
		require.Empty(t, image.GetId())
	}
}

func Test_Settings_SingleMode_RemoveWhitelistedImage_NotSupported(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)

	req := &velez_api.RemoveWhitelistedImage_Request{Id: uuid.NewString()}

	_, err := env.SettingsClient().RemoveWhitelistedImage(t.Context(), req)
	require.Equal(t, codes.FailedPrecondition, status.Code(err))
}

func Test_RegisterContainer_ClusterMode_ElevatedAccessWhitelist(t *testing.T) {
	t.Parallel()

	env, _ := enableStatefullPgUnderDind(t, Planes[1], accessClusterSfx)
	dockerClient := env.DockerClient()

	repository := runtime_policy.RepositoryOf(NginxAlpineImage)

	allowedName, allowedService := registerContainerNames("accwl")
	refusedName, refusedService := registerContainerNames("accwl2")

	allowed := startContainerWith(t, dockerClient, allowedName, newCapAddHostConfig())

	require.Nil(t, findWhitelistedImage(t, env, repository), "the image starts out not whitelisted")

	requireRegisterDone(t, env, newElevatedRegisterRequest(allowed.id, allowedService, true))

	kept, err := dockerClient.ContainerInspect(t.Context(), allowedName)
	require.NoError(t, err)
	require.Equal(t, allowed.id, kept.ID, "cluster mode does not recreate the container")
	require.Equal(t, []string{policyNetAdminCap}, bareCapNames(kept.HostConfig.CapAdd))

	row := findWhitelistedImage(t, env, repository)
	require.NotNil(t, row, "allowing elevated access whitelists the image repository")
	require.NotEmpty(t, row.GetId())

	got := getContainerById(t, env, allowed.id)
	requireReportedElevatedAccess(t, got, newCapAddHostConfig())
	require.True(t, got.GetIsImageWhitelisted())

	removeWhitelistedImage(t, env, row.GetId())

	require.Nil(t, findWhitelistedImage(t, env, repository), "the removed row must be gone")
	require.False(t, getContainerById(t, env, allowed.id).GetIsImageWhitelisted())

	refused := startContainerWith(t, dockerClient, refusedName, newCapAddHostConfig())

	req := newElevatedRegisterRequest(refused.id, refusedService, false)

	requireRefusedAndUntouched(t, env, refused, refusedService, req, user_errors.ErrElevatedAccessImageNotWhitelisted)
}
