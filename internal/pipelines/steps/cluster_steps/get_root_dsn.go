package cluster_steps

import (
	"context"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka/resources"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/cluster/cluster_state"
	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/patterns/db_patterns/pg_pattern"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
)

type getPgDbDsn struct {
	docker      node_clients.Docker
	containerId *string
	connResp    *string
}

func GetRgRootDsn(
	docker node_clients.Docker,
	containerId *string,
	dsnResp *string,
) steps.Step {
	return &getPgDbDsn{
		docker,
		containerId,
		dsnResp,
	}
}

func (p *getPgDbDsn) Do(ctx context.Context) error {
	pgCfg := &resources.Postgres{
		Host: cluster_state.Name,
		Port: 5432,

		User:    "postgres",
		Pwd:     "",
		DbName:  "",
		SslMode: "disable",
	}

	cont, err := p.docker.Client().ContainerInspect(ctx, *p.containerId)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container")
	}

	for _, v := range cont.Config.Env {
		envVarParts := strings.SplitN(v, "=", 2)
		if len(envVarParts) < 2 {
			continue
		}

		switch envVarParts[0] {
		case pg_pattern.DbEnvVariable:
			pgCfg.DbName = envVarParts[1]
		case pg_pattern.UserEnvVariable:
			pgCfg.User = envVarParts[1]
		case pg_pattern.PasswordEnvVariable:
			pgCfg.Pwd = envVarParts[1]
		default:
			continue
		}
	}

	if !env.IsInContainer() {
		pgCfg.Host = "localhost"
		pgCfg.Port, err = p.getExposedPort(cont)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	*p.connResp = pgCfg.ConnectionString()

	return nil
}

func (p *getPgDbDsn) getExposedPort(cont container.InspectResponse) (uint64, error) {
	if cont.NetworkSettings == nil {
		return 0, rerrors.New("no network settings found in container")
	}

	ports := cont.NetworkSettings.Ports[pg_pattern.TcpPort]
	if len(ports) == 0 {
		return 0, rerrors.New("no exposure for 5432 found")
	}

	hostPort := ports[0].HostPort
	port, err := strconv.ParseUint(hostPort, 10, 64)
	if err != nil {
		return 0, rerrors.Wrap(err, "error parsing exposed port for container to uint64")
	}

	return port, nil
}
