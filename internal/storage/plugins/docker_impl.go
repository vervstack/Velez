package plugins

import (
	"context"
	"strings"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
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

func (d *dockerPluginsStorage) ListPlugins(ctx context.Context) ([]domain.PluginBaseInfo, error) {
	listReq := &pb.ListSmerds_Request{}
	containers, err := d.docker.ListContainers(ctx, listReq)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	var rows []domain.PluginBaseInfo

	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		var pluginType pb.VervServiceType

		switch name {
		case makoshContainerName:
			pluginType = pb.VervServiceType_makosh
		case matreshkaContainerName:
			pluginType = pb.VervServiceType_matreshka
		case portainerContainerName:
			pluginType = pb.VervServiceType_portainer
		case headscaleContainerName:
			pluginType = pb.VervServiceType_headscale
		case pgContainerName:
			pluginType = pb.VervServiceType_statefull_pg
		default:
			continue
		}

		row := domain.PluginBaseInfo{
			Name: pluginType.String(),
			// Not in statfull mode - no service
			ServiceId: nil,
		}

		rows = append(rows, row)
	}

	return rows, nil
}

func (d *dockerPluginsStorage) UpsertPlugin(_ context.Context, _ plugins_queries.UpsertPluginParams) error {
	return nil
}
