package local_storage

import (
	"context"
	"strings"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
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

func newPluginsStorage(docker node_clients.Docker) *dockerPluginsStorage {
	return &dockerPluginsStorage{docker: docker}
}

var pluginContainerNames = map[string]pb.VervPluginType{
	makoshContainerName:    pb.VervPluginType_makosh,
	matreshkaContainerName: pb.VervPluginType_matreshka,
	portainerContainerName: pb.VervPluginType_portainer,
	headscaleContainerName: pb.VervPluginType_headscale,
	pgContainerName:        pb.VervPluginType_statefull_pg,
}

func (d *dockerPluginsStorage) ListPlugins(ctx context.Context) ([]domain.PluginBaseInfo, error) {
	listReq := &pb.ListSmerds_Request{}
	containers, err := d.docker.ListContainers(ctx, listReq)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	containerStates := make(map[string]string, len(containers))
	for _, c := range containers {
		if len(c.Names) == 0 {
			continue
		}
		name := strings.TrimPrefix(c.Names[0], "/")
		containerStates[name] = c.State
	}

	rows := make([]domain.PluginBaseInfo, 0, len(pluginContainerNames))
	for containerName, pluginType := range pluginContainerNames {
		dockerState, exists := containerStates[containerName]

		var state pb.VervPlugin_State
		switch {
		case !exists:
			state = pb.VervPlugin_disabled
		case dockerState == "running":
			state = pb.VervPlugin_running
		case dockerState == "restarting":
			state = pb.VervPlugin_warning
		default:
			state = pb.VervPlugin_dead
		}

		row := domain.PluginBaseInfo{
			Name:        pluginType.String(),
			State:       state,
			ServiceId:   nil,
			ServiceName: containerName,
		}
		rows = append(rows, row)
	}

	return rows, nil
}

func (d *dockerPluginsStorage) UpsertPlugin(_ context.Context, _ plugins_queries.UpsertPluginParams) error {
	return nil
}
