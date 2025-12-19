package patterns

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"

	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	postgresImage = "postgres:18"
)

func Postgres(name, password string) container.CreateRequest {
	volumeName := name
	return container.CreateRequest{
		Config: &container.Config{
			Image: postgresImage,
			Env: []string{
				"POSTGRES_PASSWORD=" + password,
			},
			Volumes: map[string]struct{}{
				volumeName: {},
			},
			Labels: map[string]string{
				labels.ComposeGroupLabel: name,
			},
		},
		HostConfig: &container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: volumeName,
					Target: "/var/lib/postgresql",
				},
			},
		},
	}
}
