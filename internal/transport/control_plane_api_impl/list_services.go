package control_plane_api_impl

import (
	"context"
	"sort"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients/state"
	"go.vervstack.ru/Velez/internal/cluster/configuration"
	"go.vervstack.ru/Velez/internal/cluster/service_discovery"
	"go.vervstack.ru/Velez/internal/cluster/verv_closed_network"
	"go.vervstack.ru/Velez/internal/patterns"
	pb "go.vervstack.ru/Velez/pkg/velez_api"
)

func (impl *Impl) ListServices(ctx context.Context, _ *pb.ListVervServices_Request) (
	*pb.ListVervServices_Response, error) {

	smerds, err := impl.smerdManager.ListSmerds(ctx, nil)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	resp := &pb.ListVervServices_Response{}

	for _, smerd := range smerds.Smerds {
		srv := &pb.VervService{
			Type:  pb.VervServiceType_unknown_service_type,
			State: getState(smerd),
		}

		switch smerd.Name {
		case service_discovery.Name:
			srv.Type = pb.VervServiceType_makosh
		case configuration.Name:
			srv.Type = pb.VervServiceType_matreshka
			srv.Port = getPort(smerd.Ports)
		case patterns.PortainerServiceName:
			srv.Type = pb.VervServiceType_portainer
			srv.Port = getPort(smerd.Ports)
		case verv_closed_network.Name:
			srv.Type = pb.VervServiceType_headscale
		case state.PgName:
			srv.Type = pb.VervServiceType_statefull_pg
			srv.Port = getPort(smerd.Ports)
		default:
			continue
		}

		resp.Services = append(resp.Services, srv)
	}

	resp.Services = append(resp.Services, listInactiveServices(resp.Services)...)

	sort.Slice(resp.Services, func(i, j int) bool {
		return resp.Services[i].Type < resp.Services[j].Type
	})

	return resp, nil
}

func getState(smerd *pb.Smerd) pb.VervService_State {
	switch smerd.Status {
	case pb.Smerd_created, pb.Smerd_restarting, pb.Smerd_removing, pb.Smerd_paused:
		return pb.VervService_warning
	case pb.Smerd_running:
		return pb.VervService_running
	case pb.Smerd_exited, pb.Smerd_dead:
		return pb.VervService_dead
	default:
		return pb.VervService_unknown
	}
}

func getPort(ports []*pb.Port) *uint32 {
	for _, port := range ports {
		if port.ExposedTo != nil && *port.ExposedTo != 0 {
			return port.ExposedTo
		}
	}

	return nil
}

func listInactiveServices(enabledServices []*pb.VervService) []*pb.VervService {
	enabledServicesMap := make(map[pb.VervServiceType]struct{})
	for _, s := range enabledServices {
		enabledServicesMap[s.Type] = struct{}{}
	}

	var disabledServices []*pb.VervService
	for vervService := range pb.VervServiceType_name {
		if vervService == 0 {
			continue
		}

		_, exists := enabledServicesMap[pb.VervServiceType(vervService)]
		if exists {
			continue
		}

		srv := &pb.VervService{
			Type: pb.VervServiceType(vervService),
		}

		disabledServices = append(disabledServices, srv)
	}

	sort.Slice(disabledServices, func(i, j int) bool {
		return disabledServices[i].Type < disabledServices[j].Type
	})

	return disabledServices
}
