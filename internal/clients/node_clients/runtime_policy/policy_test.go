package runtime_policy

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

const (
	presetRuntimeName = "custom-runc"
	nginxImage        = "nginx:1"
	userRepository    = "acme/agent"
)

func newPlainHostConfig() *container.HostConfig {
	return &container.HostConfig{}
}

func newPresetRuntimeHostConfig() *container.HostConfig {
	return &container.HostConfig{Runtime: presetRuntimeName}
}

func newTailscaleHostConfig() *container.HostConfig {
	hostConfig := &container.HostConfig{
		CapAdd:      []string{"NET_ADMIN"},
		NetworkMode: "container:other",
	}

	hostConfig.Devices = []container.DeviceMapping{
		{PathOnHost: "/dev/net/tun", PathInContainer: "/dev/net/tun"},
	}

	return hostConfig
}

func newNetworkJoinHostConfig() *container.HostConfig {
	return &container.HostConfig{NetworkMode: "container:other"}
}

func newDockerSocketMountHostConfig() *container.HostConfig {
	return &container.HostConfig{
		Mounts: []mount.Mount{{Type: mount.TypeBind, Source: dockerSocketPath, Target: dockerSocketPath}},
	}
}

func newDockerSocketBindHostConfig() *container.HostConfig {
	return &container.HostConfig{Binds: []string{dockerSocketPath + ":" + dockerSocketPath}}
}

func newPlainBindHostConfig() *container.HostConfig {
	return &container.HostConfig{Binds: []string{"/data:/data"}}
}

func Test_Apply_Cases(t *testing.T) {
	sysboxOn := domain.Settings{IsSysboxEnabled: true}
	sysboxOnIgnoreWhitelist := domain.Settings{IsSysboxEnabled: true, IsSysboxWhitelistIgnored: true}

	cases := []struct {
		name        string
		hostConfig  *container.HostConfig
		image       string
		settings    domain.Settings
		wantRuntime string
		wantErr     bool
	}{
		{"default sysbox off", newPlainHostConfig(), nginxImage, domain.Settings{}, "", false},
		{
			"default sysbox on", newPlainHostConfig(), nginxImage, sysboxOn, sysboxRuntimeName,
			false,
		},
		{
			"plain bind mount sysbox on", newPlainBindHostConfig(), nginxImage, sysboxOn,
			sysboxRuntimeName, false,
		},
		{
			"preset runtime kept", newPresetRuntimeHostConfig(), nginxImage, sysboxOn,
			presetRuntimeName, false,
		},
		{
			"tailscale elevated stays runc", newTailscaleHostConfig(),
			"tailscale/tailscale:v1.90.8", sysboxOn, "", false,
		},
		{
			"network join alone is not elevated", newNetworkJoinHostConfig(), nginxImage, sysboxOn,
			sysboxRuntimeName, false,
		},
		{"unknown elevated rejected", newTailscaleHostConfig(), "evil/image:1", sysboxOn, "", true},
		{
			"unknown elevated rejected sysbox off", newTailscaleHostConfig(), "evil/image:1",
			domain.Settings{}, "", true,
		},
		{
			"whitelisted ignore whitelist uses sysbox", newTailscaleHostConfig(),
			"tailscale/tailscale", sysboxOnIgnoreWhitelist, sysboxRuntimeName, false,
		},
		{
			"whitelisted ignore whitelist keeps preset",
			&container.HostConfig{Privileged: true, Runtime: presetRuntimeName},
			"docker:dind", sysboxOnIgnoreWhitelist, presetRuntimeName, false,
		},
		{
			"socket mount unknown rejected", newDockerSocketMountHostConfig(), "alpine",
			domain.Settings{}, "", true,
		},
		{
			"socket bind unknown rejected", newDockerSocketBindHostConfig(), "alpine",
			domain.Settings{}, "", true,
		},
		{
			"socket bind runner allowed", newDockerSocketBindHostConfig(),
			"gitlab/gitlab-runner:latest", domain.Settings{}, "", false,
		},
		{
			"registry prefix match", newTailscaleHostConfig(),
			"registry.example.com:5000/docker:dind", domain.Settings{}, "", false,
		},
		{
			"docker hub library match", newTailscaleHostConfig(),
			"docker.io/library/docker:dind", domain.Settings{}, "", false,
		},
		{
			"digest match", newTailscaleHostConfig(), "docker@sha256:abc", domain.Settings{}, "",
			false,
		},
		{
			"full host whitelist match", newDockerSocketBindHostConfig(),
			"ghcr.io/actions/actions-runner:latest", domain.Settings{}, "", false,
		},
		{
			"garage web ui sidecar allowed", newTailscaleHostConfig(),
			"khairul169/garage-webui:1.1.0", domain.Settings{}, "", false,
		},
		{
			"buildkit sidecar allowed", newTailscaleHostConfig(),
			"moby/buildkit:v0.34.0", domain.Settings{}, "", false,
		},
		{
			"binfmt installer allowed", newTailscaleHostConfig(),
			"tonistiigi/binfmt:qemu-v10.2.3-68", domain.Settings{}, "", false,
		},
		{
			"namespace suffix not matched", newTailscaleHostConfig(), "evil/docker:dind",
			domain.Settings{}, "", true,
		},
		{
			"case insensitive", newTailscaleHostConfig(), "Tailscale/Tailscale:v1",
			domain.Settings{}, "", false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := Apply(tc.hostConfig, tc.image, tc.settings, nil)
			if tc.wantErr {
				require.ErrorIs(t, err, user_errors.ErrElevatedAccessImageNotWhitelisted)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantRuntime, tc.hostConfig.Runtime)
		})
	}
}

func Test_Apply_NilHostConfig(t *testing.T) {
	err := Apply(nil, "nginx", domain.Settings{IsSysboxEnabled: true}, nil)
	require.NoError(t, err)
}

func Test_Apply_UserWhitelistedImagePasses(t *testing.T) {
	settings := domain.Settings{WhitelistedImages: []string{userRepository}}

	images := []string{
		userRepository,
		userRepository + ":2",
		"registry.example.com:5000/" + userRepository + ":2",
		"ACME/Agent",
	}

	for _, image := range images {
		err := Apply(newTailscaleHostConfig(), image, settings, nil)
		require.NoError(t, err, image)
	}
}

func Test_Apply_UserWhitelistDoesNotMatchOtherRepository(t *testing.T) {
	settings := domain.Settings{WhitelistedImages: []string{userRepository}}

	err := Apply(newTailscaleHostConfig(), "evil/"+userRepository, settings, nil)
	require.ErrorIs(t, err, user_errors.ErrElevatedAccessImageNotWhitelisted)
}

func Test_Apply_AllowanceLabelPasses(t *testing.T) {
	containerLabels := map[string]string{labels.ElevatedAccessLabel: labels.ElevatedAccessAllowedValue}

	err := Apply(newTailscaleHostConfig(), "evil/image:1", domain.Settings{}, containerLabels)
	require.NoError(t, err)
}

func Test_Apply_AllowanceLabelWithOtherValueRefused(t *testing.T) {
	containerLabels := map[string]string{labels.ElevatedAccessLabel: "yes"}

	err := Apply(newTailscaleHostConfig(), "evil/image:1", domain.Settings{}, containerLabels)
	require.ErrorIs(t, err, user_errors.ErrElevatedAccessImageNotWhitelisted)
}

func Test_Apply_NeitherWhitelistNorLabelRefused(t *testing.T) {
	settings := domain.Settings{WhitelistedImages: []string{userRepository}}
	containerLabels := map[string]string{"other": "label"}

	err := Apply(newTailscaleHostConfig(), "evil/image:1", settings, containerLabels)
	require.ErrorIs(t, err, user_errors.ErrElevatedAccessImageNotWhitelisted)
}

func Test_Apply_NonElevatedUnaffectedByWhitelist(t *testing.T) {
	hostConfig := newPlainHostConfig()

	err := Apply(hostConfig, "evil/image:1", domain.Settings{IsSysboxEnabled: true}, nil)
	require.NoError(t, err)
	require.Equal(t, sysboxRuntimeName, hostConfig.Runtime)
}

func Test_IsElevatedAllowed_Cases(t *testing.T) {
	settings := domain.Settings{WhitelistedImages: []string{userRepository}}
	allowedLabels := map[string]string{labels.ElevatedAccessLabel: labels.ElevatedAccessAllowedValue}

	require.True(t, IsElevatedAllowed(newPlainHostConfig(), "evil/image", domain.Settings{}, nil))
	require.True(t, IsElevatedAllowed(nil, "evil/image", domain.Settings{}, nil))
	require.True(t, IsElevatedAllowed(newTailscaleHostConfig(), "docker:dind", domain.Settings{}, nil))
	require.True(t, IsElevatedAllowed(newTailscaleHostConfig(), userRepository+":1", settings, nil))
	require.True(t, IsElevatedAllowed(newTailscaleHostConfig(), "evil/image", domain.Settings{}, allowedLabels))
	require.False(t, IsElevatedAllowed(newTailscaleHostConfig(), "evil/image", settings, nil))
}

func Test_RepositoryOf_Cases(t *testing.T) {
	require.Equal(t, userRepository, RepositoryOf("Registry.Example.com:5000/ACME/Agent:2"))
	require.Equal(t, userRepository, RepositoryOf(userRepository+"@sha256:abc"))
	require.Equal(t, "library/docker", RepositoryOf("docker.io/library/docker:dind"))
	require.Equal(t, "nginx", RepositoryOf("nginx"))
}

func Test_BuiltinWhitelist_ReturnsCopy(t *testing.T) {
	list := BuiltinWhitelist()
	require.Contains(t, list, "tailscale/tailscale")

	list[0] = "changed"

	require.NotEqual(t, "changed", BuiltinWhitelist()[0])
}

func Test_IsImageAllowed_Table(t *testing.T) {
	settings := domain.Settings{WhitelistedImages: []string{userRepository}}
	allowedLabels := map[string]string{labels.ElevatedAccessLabel: labels.ElevatedAccessAllowedValue}

	cases := []struct {
		name            string
		image           string
		settings        domain.Settings
		containerLabels map[string]string
		want            bool
	}{
		{"built-in image", "docker:dind", domain.Settings{}, nil, true},
		{"user whitelisted image", userRepository + ":1", settings, nil, true},
		{"allowance label", "evil/image", domain.Settings{}, allowedLabels, true},
		{"unknown image", "evil/image", settings, nil, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, IsImageAllowed(tc.image, tc.settings, tc.containerLabels))
		})
	}
}
