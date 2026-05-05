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
	plugins storage.PluginsStorage
}

func NewPluginService(plugins storage.PluginsStorage) service.PluginService {
	return &pluginService{plugins: plugins}
}

func (p *pluginService) ListPlugins(ctx context.Context) (*pb.ListPlugins_Response, error) {
	rows, err := p.plugins.ListPlugins(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	var active []*pb.Plugin

	for _, row := range rows {
		plugin := &pb.Plugin{
			Type:  pb.VervServiceType(pb.VervServiceType_value[row.Name]),
			State: row.State,
		}

		//TODO get from service
		//if row.Port.Valid {
		//	port := uint32(row.Port.Int32)
		//	plugin.Port = &port
		//}

		active = append(active, plugin)
	}

	result := append(active,
		listInactivePlugins(active)...)

	sort.Slice(result, func(i, j int) bool {
		return result[i].Type < result[j].Type
	})

	resp := &pb.ListPlugins_Response{
		Plugins: result,
	}

	return resp, nil
}

func listInactivePlugins(activePlugins []*pb.Plugin) []*pb.Plugin {
	activeMap := make(map[pb.VervServiceType]struct{})
	for _, p := range activePlugins {
		activeMap[p.Type] = struct{}{}
	}

	var disabled []*pb.Plugin
	for vervService := range pb.VervServiceType_name {
		if vervService == 0 {
			continue
		}

		_, exists := activeMap[pb.VervServiceType(vervService)]
		if exists {
			continue
		}

		plugin := &pb.Plugin{
			Type: pb.VervServiceType(vervService),
		}

		disabled = append(disabled, plugin)
	}

	return disabled
}
