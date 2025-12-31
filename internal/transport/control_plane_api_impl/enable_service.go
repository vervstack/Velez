package control_plane_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"

	"go.vervstack.ru/Velez/internal/domain"
	pb "go.vervstack.ru/Velez/pkg/velez_api"
)

var (
	errUnsupportedService = rerrors.New("unsupported service", codes.InvalidArgument)
)

func (impl *Impl) EnableService(ctx context.Context, req *pb.EnableService_Request) (
	*pb.EnableService_Response, error) {

	var err error
	switch req.GetService() {
	case pb.VervServiceType_statefull_pg:
		payload, ok := req.Payload.(*pb.EnableService_Request_StatefullCluster)
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
		return &pb.EnableService_Response{}, rerrors.Wrap(err)
	}

	return &pb.EnableService_Response{}, nil
}
