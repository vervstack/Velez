package runtime_policy

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	sysboxRuntimeName = "sysbox-runc"
	dockerSocketPath  = "/var/run/docker.sock"
)

// SettingsProvider yields the node-wide settings. Nil means defaults.
type SettingsProvider interface {
	GetSettings(ctx context.Context) (domain.Settings, error)
}

func Resolve(
	ctx context.Context, provider SettingsProvider, hostConfig *container.HostConfig, image string,
) error {
	var settings domain.Settings

	if provider != nil {
		var err error

		settings, err = provider.GetSettings(ctx)
		if err != nil {
			return rerrors.Wrap(err, "error getting node settings")
		}
	}

	err := Apply(hostConfig, image, settings)
	if err != nil {
		return rerrors.Wrap(err, "error applying runtime policy")
	}

	return nil
}

func Apply(hostConfig *container.HostConfig, image string, settings domain.Settings) error {
	if hostConfig == nil {
		return nil
	}

	canUseSysbox := settings.IsSysboxEnabled && hostConfig.Runtime == ""

	if !isElevated(hostConfig) {
		if canUseSysbox {
			hostConfig.Runtime = sysboxRuntimeName
		}

		return nil
	}

	if !isImageWhitelisted(image) {
		return rerrors.Wrap(user_errors.ErrElevatedAccessImageNotWhitelisted, "error checking image "+image)
	}

	if settings.IsSysboxWhitelistIgnored && canUseSysbox {
		hostConfig.Runtime = sysboxRuntimeName
	}

	return nil
}

func isElevated(hostConfig *container.HostConfig) bool {
	return hostConfig.Privileged ||
		len(hostConfig.CapAdd) > 0 ||
		len(hostConfig.Devices) > 0 ||
		hostConfig.NetworkMode.IsHost() ||
		hostConfig.NetworkMode.IsContainer() ||
		hostConfig.PidMode != "" ||
		hostConfig.IpcMode.IsHost() ||
		hostConfig.UsernsMode.IsHost() ||
		hasDockerSocket(hostConfig)
}

func hasDockerSocket(hostConfig *container.HostConfig) bool {
	for _, m := range hostConfig.Mounts {
		if m.Type == mount.TypeBind && m.Source == dockerSocketPath {
			return true
		}
	}

	for _, bind := range hostConfig.Binds {
		if strings.HasPrefix(bind, dockerSocketPath) {
			return true
		}
	}

	return false
}
