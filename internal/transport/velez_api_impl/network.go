package velez_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	api "go.vervstack.ru/Velez/pkg/velez_api"
)

func (a *Impl) MakeConnections(ctx context.Context, req *api.MakeConnections_Request) (
	*api.MakeConnections_Response, error) {

	for _, conn := range req.Connections {
		err := a.smerdService.ConnectToNetwork(ctx, toConnection(conn))
		if err != nil {
			return nil, rerrors.Wrap(err, "error connecting to network")
		}
	}

	return &api.MakeConnections_Response{}, nil
}

func (a *Impl) BreakConnections(ctx context.Context, req *api.BreakConnections_Request) (
	*api.BreakConnections_Response, error) {

	for _, conn := range req.Connections {
		err := a.smerdService.DisconnectFromNetwork(ctx, toConnection(conn))
		if err != nil {
			return nil, rerrors.Wrap(err, "error connecting to network")
		}
	}

	return &api.BreakConnections_Response{}, nil
}

func toConnection(in *api.Connection) domain.Connection {
	return domain.Connection{
		SmerdName: in.ServiceName,
		Network:   in.TargetNetwork,
		Aliases:   in.Aliases,
	}
}
