package plugins

import (
	"context"
	"database/sql"
	"strings"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/plugins_queries"
)

const (
	makoshContainerName    = "makosh"
	matreshkaContainerName = "matreshka"
	portainerContainerName = "portainer"
	headscaleContainerName = "headscale"
	pgContainerName        = "verv-cluster-state"
)

type dockerPluginsStorage struct {
	docker node_clients.Docker
}

func NewDockerImpl(docker node_clients.Docker) storage.PluginsStorage {
	return &dockerPluginsStorage{docker: docker}
}

func (d *dockerPluginsStorage) ListPlugins(ctx context.Context) ([]plugins_queries.ListPluginsRow, error) {
	listReq := &pb.ListSmerds_Request{}
	containers, err := d.docker.ListContainers(ctx, listReq)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	var rows []plugins_queries.ListPluginsRow

	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		var pluginType int32

		switch name {
		case makoshContainerName:
			pluginType = int32(pb.VervServiceType_makosh)
		case matreshkaContainerName:
			pluginType = int32(pb.VervServiceType_matreshka)
		case portainerContainerName:
			pluginType = int32(pb.VervServiceType_portainer)
		case headscaleContainerName:
			pluginType = int32(pb.VervServiceType_headscale)
		case pgContainerName:
			pluginType = int32(pb.VervServiceType_statefull_pg)
		default:
			continue
		}

		var port sql.NullInt32
		for _, p := range c.Ports {
			if p.PublicPort != 0 {
				port = sql.NullInt32{Valid: true, Int32: int32(p.PublicPort)}
				break
			}
		}

		row := plugins_queries.ListPluginsRow{
			PluginType: pluginType,
			State:      int32(containerStateToPluginState(c.State)),
			Port:       port,
		}

		rows = append(rows, row)
	}

	return rows, nil
}

func (d *dockerPluginsStorage) UpsertPlugin(_ context.Context, _ plugins_queries.UpsertPluginParams) error {
	return nil
}

func containerStateToPluginState(state string) pb.VervService_State {
	switch state {
	case "running":
		return pb.VervService_running
	case "exited", "dead":
		return pb.VervService_dead
	case "paused", "restarting", "removing", "created":
		return pb.VervService_warning
	default:
		return pb.VervService_unknown
	}
}

var _ storage.PluginsStorage = (*dockerPluginsStorage)(nil)
