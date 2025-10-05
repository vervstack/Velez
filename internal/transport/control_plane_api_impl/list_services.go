package control_plane_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/backservice/configuration"
	"go.vervstack.ru/Velez/internal/backservice/service_discovery"
	"go.vervstack.ru/Velez/pkg/control_plane_api"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

func (impl *Impl) ListServices(ctx context.Context, _ *control_plane_api.ListServices_Request) (
	*control_plane_api.ListServices_Response, error) {

	smerds, err := impl.smerdManager.ListSmerds(ctx, nil)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	resp := &control_plane_api.ListServices_Response{}

	for _, smerd := range smerds.Smerds {
		srv := &control_plane_api.Service{
			Type: control_plane_api.ServiceType_unknown_service_type,
		}

		switch smerd.Name {
		case service_discovery.Name:
			srv.Type = control_plane_api.ServiceType_makosh
		case configuration.Name:
			srv.Type = control_plane_api.ServiceType_matreshka
			srv.Port = getPort(smerd.Ports)
		default:
			continue
		}

		resp.Services = append(resp.Services, srv)
	}

	return resp, nil
}

func getPort(ports []*velez_api.Port) *uint32 {
	if len(ports) == 0 {
		return nil
	}

	return ports[0].ExposedTo
}
