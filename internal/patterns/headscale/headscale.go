package headscale

import (
	_ "embed"
	"path"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
	rtb "go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

//go:embed config.yaml
var headscaleBasicConfig []byte

const (
	ServiceName = "headscale"

	groupName         = "verv_closed_network"
	defaultImage      = "headscale/headscale:0.27.2-rc.1"
	ApiPort           = "8080"
	defaultConfigPath = "/etc/headscale/config.yaml"
)

func Headscale(r domain.SetupHeadscaleRequest) container.CreateRequest {
	name := ServiceName

	return container.CreateRequest{
		Config: &container.Config{
			Hostname: name,
			ExposedPorts: nat.PortSet{
				ApiPort: struct{}{},
			},
			Cmd: strslice.StrSlice{"serve"},
			Healthcheck: &container.HealthConfig{
				Test: []string{"CMD", "headscale", "health"},
			},

			Image: rtb.Coalesce(rtb.FromPtr(r.CustomImage), defaultImage),

			Labels: map[string]string{
				labels.VervServiceLabel:  "true",
				labels.ComposeGroupLabel: groupName,
			},
		},
		HostConfig: &container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: name,
					Target: path.Dir(defaultConfigPath),
				},
				{
					Type:   mount.TypeVolume,
					Source: name,
					Target: "/var/lib/headscale",
				},
			},
			PortBindings: map[nat.Port][]nat.PortBinding{
				ApiPort: {
					{
						HostPort: rtb.Coalesce(rtb.FromPtr(r.ExposeToPort), ApiPort),
					},
				},
			},
		},
	}
}

func BasicConfig() []byte {
	b := make([]byte, 0, len(headscaleBasicConfig))
	copy(b, headscaleBasicConfig)
	return b
}
