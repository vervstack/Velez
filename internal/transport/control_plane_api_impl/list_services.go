package control_plane_api_impl

import (
	"context"
	"sort"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/cluster/configuration"
	"go.vervstack.ru/Velez/internal/cluster/service_discovery"
	"go.vervstack.ru/Velez/internal/cluster/verv_closed_network"
	"go.vervstack.ru/Velez/internal/patterns"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

func (impl *Impl) ListServices(ctx context.Context, _ *velez_api.ListServices_Request) (
	*velez_api.ListServices_Response, error) {

	smerds, err := impl.smerdManager.ListSmerds(ctx, nil)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	resp := &velez_api.ListServices_Response{}

	for _, smerd := range smerds.Smerds {
		srv := &velez_api.Service{
			Type: velez_api.VervServiceType_unknown_service_type,
		}

		switch smerd.Name {
		case service_discovery.Name:
			srv.Type = velez_api.VervServiceType_makosh
		case configuration.Name:
			srv.Type = velez_api.VervServiceType_matreshka
			srv.Port = getPort(smerd.Ports)
		case patterns.PortainerServiceName:
			srv.Type = velez_api.VervServiceType_portainer
			srv.Port = getPort(smerd.Ports)
		case verv_closed_network.Name:
			srv.Type = velez_api.VervServiceType_headscale
		default:
			continue
		}
		_, srv.Togglable = togglableServices[srv.Type]

		resp.Services = append(resp.Services, srv)
	}

	resp.InactiveServices = listInactiveServices(resp.Services)

	return resp, nil
}

func getPort(ports []*velez_api.Port) *uint32 {
	for _, port := range ports {
		if port.ExposedTo != nil && *port.ExposedTo != 0 {
			return port.ExposedTo
		}
	}

	return nil
}

func listInactiveServices(enabledServices []*velez_api.Service) []*velez_api.Service {
	enabledServicesMap := make(map[velez_api.VervServiceType]struct{})
	for _, s := range enabledServices {
		enabledServicesMap[s.Type] = struct{}{}
	}

	var disabledServices []*velez_api.Service
	for tp, constructor := range supportedServicesMapConstructors {
		_, exists := enabledServicesMap[tp]
		if exists {
			continue
		}

		srv := &velez_api.Service{
			Type: tp,
		}

		_, srv.Togglable = togglableServices[srv.Type]

		if constructor != nil {
			srv.Constructor = constructor()
		}

		disabledServices = append(disabledServices, srv)
	}

	sort.Slice(disabledServices, func(i, j int) bool {
		return disabledServices[i].Type < disabledServices[j].Type
	})

	return disabledServices
}

var (
	supportedServicesMapConstructors = map[velez_api.VervServiceType]func() *velez_api.CreateSmerd_Request{
		velez_api.VervServiceType_matreshka: nil,
		velez_api.VervServiceType_makosh:    nil,

		velez_api.VervServiceType_headscale: nil,

		velez_api.VervServiceType_webserver: nil,
		velez_api.VervServiceType_portainer: nil,

		velez_api.VervServiceType_cluster_mode: nil,
	}

	togglableServices = map[velez_api.VervServiceType]struct{}{
		velez_api.VervServiceType_headscale: {},
	}
)
