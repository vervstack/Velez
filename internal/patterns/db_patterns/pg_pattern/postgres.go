// Package pg_pattern provides a Docker container pattern for PostgreSQL.
package pg_pattern

import (
	"strconv"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	rtb "go.redsock.ru/toolbox"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/matreshka/pkg/matreshka/resources"
)

const (
	postgresImage              = "postgres:18"
	postgresInstanceName       = "postgres"
	postgresDbName             = "postgres"
	postgresUser               = "postgres"
	postgresPort               = 5432
	postgresRandomPasswordSize = 16
	DbEnvVariable              = "POSTGRES_DB"
	UserEnvVariable            = "POSTGRES_USER"
	PasswordEnvVariable        = "POSTGRES_PASSWORD"
	TCPPort                    = "5432/tcp"

	healthCheckInterval    = 2 * time.Second
	healthCheckTimeout     = 3 * time.Second
	healthCheckRetries     = 3
	healthCheckStartPeriod = 10 * time.Second
)

// Constructor contains PostgreSQL container configuration options.
type Constructor struct {
	InstanceName  string
	IsPortExposed bool
	ExposedToPort *uint64

	MatreshkaPg resources.Postgres
}

// Pattern is a Docker container pattern for PostgreSQL.
type Pattern struct {
	Constructor

	Pattern container.CreateRequest
}

// Postgres creates a PostgreSQL container with the given options.
func Postgres(opts ...Opt) Pattern {
	ctor := basicPostgresConstructor()

	for _, o := range opts {
		o(&ctor)
	}

	portStr := strconv.Itoa(postgresPort)
	healthCheckCmd := "pg_isready -U \"$" + UserEnvVariable + "\" -d \"$" + DbEnvVariable +
		"\" -h 127.0.0.1 -p " + portStr

	createReq := container.CreateRequest{
		Config: &container.Config{
			Hostname: ctor.InstanceName,
			Image:    postgresImage,
			Env: []string{
				DbEnvVariable + "=" + ctor.MatreshkaPg.DbName,
				UserEnvVariable + "=" + ctor.MatreshkaPg.User,
				PasswordEnvVariable + "=" + ctor.MatreshkaPg.Pwd,
			},
			Volumes: map[string]struct{}{
				ctor.InstanceName: {},
			},
			Labels: map[string]string{
				labels.ComposeGroupLabel: ctor.InstanceName,
			},
			Healthcheck: &container.HealthConfig{
				Test: []string{
					"CMD-SHELL",
					healthCheckCmd,
				},
				Interval:    healthCheckInterval,
				Timeout:     healthCheckTimeout,
				Retries:     healthCheckRetries,
				StartPeriod: healthCheckStartPeriod,
			},
		},
		HostConfig: &container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: ctor.InstanceName,
					Target: "/var/lib/postgresql",
				},
			},
		},
	}

	if ctor.IsPortExposed {
		createReq.ExposedPorts = map[nat.Port]struct{}{
			TCPPort: {},
		}

		var hostPort string

		if ctor.ExposedToPort != nil {
			hostPort = strconv.FormatUint(*ctor.ExposedToPort, 10)
		}

		createReq.HostConfig.PortBindings = map[nat.Port][]nat.PortBinding{
			TCPPort: {
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
		InstanceName: postgresInstanceName,
		MatreshkaPg: resources.Postgres{
			DbName: postgresDbName,
			User:   postgresUser,
			Pwd:    string(rtb.RandomBase64(postgresRandomPasswordSize)),
			Port:   postgresPort,
		},
		ExposedToPort: nil,
	}
}
