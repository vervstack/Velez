package control_plane_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/transport/common"
)

func (impl *Impl) ListNodes(ctx context.Context, req *pb.ListNodes_Request) (*pb.ListNodes_Response, error) {
	r := domain.ListNodesReq{
		Paging: common.FromPaging(req.Paging),
	}

	nodesList, err := impl.nodeService.ListNodes(ctx, r)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing nodes")
	}

	return &pb.ListNodes_Response{
		Nodes: common.ToNodeList(nodesList.Nodes),
		Total: nodesList.Total,
	}, nil
}
