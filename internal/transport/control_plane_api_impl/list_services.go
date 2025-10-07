package control_plane_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/backservice/configuration"
	"go.vervstack.ru/Velez/internal/backservice/service_discovery"
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
			Type: velez_api.ServiceType_unknown_service_type,
		}

		switch smerd.Name {
		case service_discovery.Name:
			srv.Type = velez_api.ServiceType_makosh
		case configuration.Name:
			srv.Type = velez_api.ServiceType_matreshka
			srv.Port = getPort(smerd.Ports)
		default:
			continue
		}

		resp.Services = append(resp.Services, srv)
	}

	resp.InactiveServices = listInactiveServices(resp.Services)

	return resp, nil
}

func getPort(ports []*velez_api.Port) *uint32 {
	if len(ports) == 0 {
		return nil
	}

	return ports[0].ExposedTo
}

func listInactiveServices(enabledServices []*velez_api.Service) []*velez_api.Service {
	enabledServicesMap := make(map[velez_api.ServiceType]struct{})
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

		if constructor != nil {
			srv.Constructor = constructor()
		}

		disabledServices = append(disabledServices, srv)

	}

	return disabledServices
}

var supportedServicesMapConstructors = map[velez_api.ServiceType]func() *velez_api.CreateSmerd_Request{
	velez_api.ServiceType_makosh:    nil,
	velez_api.ServiceType_matreshka: nil,
	velez_api.ServiceType_svarog:    nil,

	velez_api.ServiceType_webserver: nil,
	velez_api.ServiceType_portainer: patterns.Portainer,
}
