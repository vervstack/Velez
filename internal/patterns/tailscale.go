package patterns

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
)

const tailscaleImage = "tailscale/tailscale:latest"

func TailScaleSidecar(serviceName string) container.CreateRequest {
	volumeName := serviceName + "-ts-sidecar"
	return container.CreateRequest{
		Config: &container.Config{
			//Cmd: []string{"sleep", "infinity"},
			Cmd:   []string{"tailscaled", "--state=/var/lib/tailscale/tailscaled.state"},
			Image: tailscaleImage,
			Volumes: map[string]struct{}{
				volumeName: {},
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
