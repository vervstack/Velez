package plugins

import (
	"context"
	"sort"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/storage"
)

type pluginService struct {
	storage storage.Storage
}

func NewPluginService(stg storage.Storage) service.PluginService {
	return &pluginService{storage: stg}
}

func (p *pluginService) ListPlugins(ctx context.Context) (*pb.ListPlugins_Response, error) {
	rows, err := p.storage.Plugins().ListPlugins(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing plugins")
	}

	var active []*pb.Plugin

	for _, row := range rows {
		plugin := &pb.Plugin{
			Type:        pb.VervPluginType(pb.VervPluginType_value[row.Name]),
			State:       row.State,
			ServiceName: row.ServiceName,
		}

		active = append(active, plugin)
	}

	active = append(active,
		listInactivePlugins(active)...)

	sort.Slice(active, func(i, j int) bool {
		return active[i].GetType() < active[j].GetType()
	})

	resp := &pb.ListPlugins_Response{
		Plugins: active,
	}

	return resp, nil
}

func listInactivePlugins(activePlugins []*pb.Plugin) []*pb.Plugin {
	activeMap := make(map[pb.VervPluginType]struct{})
	for _, p := range activePlugins {
		activeMap[p.GetType()] = struct{}{}
	}

	var disabled []*pb.Plugin

	for vervService := range pb.VervPluginType_name {
		if vervService == 0 {
			continue
		}

		_, exists := activeMap[pb.VervPluginType(vervService)]
		if exists {
			continue
		}

		plugin := &pb.Plugin{
			Type: pb.VervPluginType(vervService),
		}

		disabled = append(disabled, plugin)
	}

	return disabled
}
