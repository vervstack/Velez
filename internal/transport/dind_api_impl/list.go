package dind_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func (impl *Impl) ListDinds(
	ctx context.Context,
	_ *pb.ListDinds_Request,
) (*pb.ListDinds_Response, error) {
	views, err := impl.dindService.ListDinds(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing dinds")
	}

	out := make([]*pb.DindInfo, 0, len(views))
	for _, view := range views {
		out = append(out, dindToPb(view))
	}

	resp := &pb.ListDinds_Response{
		Dinds: out,
	}

	return resp, nil
}

func dindToPb(view domain.DindView) *pb.DindInfo {
	out := &pb.DindInfo{
		Name:            view.Name,
		Address:         view.Address,
		IsSysboxEnabled: view.IsSysboxEnabled,
	}

	if !view.CreatedAt.IsZero() {
		out.CreatedAt = timestamppb.New(view.CreatedAt)
	}

	return out
}
