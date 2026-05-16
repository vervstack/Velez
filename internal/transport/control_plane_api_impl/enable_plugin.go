package control_plane_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

var (
	errUnsupportedService = rerrors.New("unsupported service", codes.InvalidArgument)
)

func (impl *Impl) EnablePlugin(ctx context.Context, req *pb.EnablePlugin_Request) (
	*pb.EnablePlugin_Response, error) {

	var err error
	switch req.GetPlugin() {
	case pb.VervPluginType_statefull_pg:
		payload, ok := req.Payload.(*pb.EnablePlugin_Request_StatefullCluster)
		if !ok {
			return nil, rerrors.New("invalid payload", codes.InvalidArgument)
		}

		r := domain.EnableStatefullClusterRequest{
			ExposePort:   payload.StatefullCluster.GetIsExposePort(),
			ExposeToPort: payload.StatefullCluster.GetExposeToPort(),
		}
		runner := impl.pipeliner.EnableStatefullMode(r)
		err = runner.Run(ctx)
	default:
		return nil, rerrors.Wrap(errUnsupportedService)
	}

	if err != nil {
		return &pb.EnablePlugin_Response{}, rerrors.Wrap(err)
	}

	return &pb.EnablePlugin_Response{}, nil
}
