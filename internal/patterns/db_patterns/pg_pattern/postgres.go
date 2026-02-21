package pg_pattern

import (
	"strconv"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	rtb "go.redsock.ru/toolbox"
	"go.vervstack.ru/matreshka/pkg/matreshka/resources"

	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	postgresImage       = "postgres:18"
	DbEnvVariable       = "POSTGRES_DB"
	UserEnvVariable     = "POSTGRES_USER"
	PasswordEnvVariable = "POSTGRES_PASSWORD"

	TcpPort = "5432/tcp"
)

type Constructor struct {
	InstanceName  string
	IsPortExposed bool
	ExposedToPort *uint64

	MatreshkaPg resources.Postgres
}

type Pattern struct {
	Constructor

	Pattern container.CreateRequest
}

func Postgres(opts ...Opt) Pattern {
	c := basicPostgresConstructor()

	for _, o := range opts {
		o(&c)
	}

	createReq := container.CreateRequest{
		Config: &container.Config{
			Hostname: c.InstanceName,
			Image:    postgresImage,
			Env: []string{
				DbEnvVariable + "=" + c.MatreshkaPg.DbName,
				UserEnvVariable + "=" + c.MatreshkaPg.User,
				PasswordEnvVariable + "=" + c.MatreshkaPg.Pwd,
			},
			Volumes: map[string]struct{}{
				c.InstanceName: {},
			},
			Labels: map[string]string{
				labels.ComposeGroupLabel: c.InstanceName,
			},
			Healthcheck: &container.HealthConfig{
				Test:        []string{"CMD-SHELL", "pg_isready -U \"$" + UserEnvVariable + "\" -d \"$" + DbEnvVariable + "\" -h 127.0.0.1 -p 5432"},
				Interval:    2 * time.Second,
				Timeout:     3 * time.Second,
				Retries:     3,
				StartPeriod: 10 * time.Second,
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
	}

	if c.IsPortExposed {
		createReq.Config.ExposedPorts = map[nat.Port]struct{}{
			TcpPort: {},
		}

		var hostPort string

		if c.ExposedToPort != nil {
			hostPort = strconv.FormatUint(*c.ExposedToPort, 10)
		}

		createReq.HostConfig.PortBindings = map[nat.Port][]nat.PortBinding{
			TcpPort: {
				{
					HostPort: hostPort,
				},
			},
		}
	}

	return Pattern{
		Pattern: createReq,
	}
}

func basicPostgresConstructor() Constructor {
	return Constructor{
		InstanceName: "postgres",
		MatreshkaPg: resources.Postgres{
			DbName: "postgres",
			User:   "postgres",
			Pwd:    string(rtb.RandomBase64(16)),
			Port:   5432,
		},
		ExposedToPort: nil,
	}
}
