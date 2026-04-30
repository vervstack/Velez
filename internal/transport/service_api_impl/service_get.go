package service_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func (impl *Impl) GetService(ctx context.Context, pbReq *pb.GetService_Request) (*pb.GetService_Response, error) {
	req := domain.GetServiceReq{}

	switch v := pbReq.Payload.(type) {
	case *pb.GetService_Request_Id:
		req.Id = &v.Id
	case *pb.GetService_Request_Name:
		req.Name = &v.Name
	}

	s, err := impl.servicesService.Get(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return &pb.GetService_Response{
		Payload: &pb.GetService_Response_VervService{
			VervService: &pb.VervAppService{
				Id:   s.Id,
				Name: s.ServiceBaseInfo.Name,
			},
		},
	}, nil
}
