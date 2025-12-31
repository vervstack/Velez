package patterns

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"

	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	tailscaleImage         = "tailscale/tailscale:v1.90.8"
	TailscaleSidecarSuffix = "ts-sidecar"
)

func TailScaleContainerSidecar(serviceName string) container.CreateRequest {
	volumeName := serviceName + "-" + TailscaleSidecarSuffix
	return container.CreateRequest{
		Config: &container.Config{
			Image: tailscaleImage,
			Env: []string{
				"TS_USERSPACE=false",
				"TS_STATE_DIR=/var/lib/tailscale",
				"TS_FORCE_LOGIN_SERVER=true",
			},
			Volumes: map[string]struct{}{
				volumeName: {},
			},
			Labels: map[string]string{
				labels.ComposeGroupLabel: serviceName,
				labels.Sidecar:           "true",
			},
		},
		HostConfig: &container.HostConfig{
			NetworkMode:   container.NetworkMode("container:" + serviceName),
			RestartPolicy: container.RestartPolicy{},
			CapAdd: strslice.StrSlice{
				"NET_ADMIN",
				"NET_RAW",
			},
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: volumeName,
					Target: "/var/lib/tailscale",
				},
			},
			Resources: container.Resources{
				Devices: []container.DeviceMapping{
					{
						PathOnHost:        "/dev/net/tun",
						PathInContainer:   "/dev/net/tun",
						CgroupPermissions: "rwm",
					},
				},
			},
		},
		NetworkingConfig: &network.NetworkingConfig{},
	}
}
