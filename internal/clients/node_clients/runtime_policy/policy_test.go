package runtime_policy

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	presetRuntimeName = "custom-runc"
	nginxImage        = "nginx:1"
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
			err := Apply(tc.hostConfig, tc.image, tc.settings)
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
	err := Apply(nil, "nginx", domain.Settings{IsSysboxEnabled: true})
	require.NoError(t, err)
}
