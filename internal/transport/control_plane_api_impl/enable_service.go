package control_plane_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"

	pb "go.vervstack.ru/Velez/pkg/velez_api"
)

var (
	errUnsupportedService = rerrors.New("unsupported service", codes.InvalidArgument)
)

func (impl *Impl) EnableService(ctx context.Context, req *pb.EnableService_Request) (
	*pb.EnableService_Response, error) {

	var err error
	switch req.GetService() {
	case pb.VervServiceType_cluster_mode:
		runner := impl.pipeliner.EnableStateFull()
		err = runner.Run(ctx)
	default:
		return nil, rerrors.Wrap(errUnsupportedService)
	}

	if err != nil {
		return &pb.EnableService_Response{}, rerrors.Wrap(err)
	}

	return &pb.EnableService_Response{}, nil
}
