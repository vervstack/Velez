package pg_pattern

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	rtb "go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	postgresImage = "postgres:18"
)

type PostgresConstructor struct {
	InstanceName  string
	DatabaseName  string
	User          string
	Password      string
	ExposedToPort *uint32
}

type PostgresPattern struct {
	PostgresConstructor

	Pattern container.CreateRequest
}

func Postgres(opts ...opt) PostgresPattern {
	c := basicPostgresConstructor()

	for _, o := range opts {
		o(&c)
	}

	return PostgresPattern{
		Pattern: container.CreateRequest{
			Config: &container.Config{
				Image: postgresImage,
				Env: []string{
					"POSTGRES_DB=" + c.DatabaseName,
					"POSTGRES_USER=" + c.User,
					"POSTGRES_PASSWORD=" + c.Password,
				},
				Volumes: map[string]struct{}{
					c.InstanceName: {},
				},
				Labels: map[string]string{
					labels.ComposeGroupLabel: c.InstanceName,
				},
			},
			HostConfig: &container.HostConfig{
				Mounts: []mount.Mount{
					{
						Type:   mount.TypeVolume,
						Source: c.InstanceName,
						Target: "/var/lib/postgresql",
					},
				},
			},
		},
	}
}

func basicPostgresConstructor() PostgresConstructor {
	return PostgresConstructor{
		InstanceName:  "postgres",
		DatabaseName:  "postgres",
		User:          "postgres",
		Password:      string(rtb.RandomBase64(16)),
		ExposedToPort: nil,
	}
}
