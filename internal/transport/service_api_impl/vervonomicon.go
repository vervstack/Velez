package service_api_impl

import (
	"context"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) GetVervonomicon(ctx context.Context, pbReq *pb.GetVervonomicon_Request) (*pb.GetVervonomicon_Response, error) {
	return &pb.GetVervonomicon_Response{}, nil
}
