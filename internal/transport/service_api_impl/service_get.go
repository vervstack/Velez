package service_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func (impl *Impl) GetService(ctx context.Context, pbReq *pb.GetService_Request) (*pb.GetService_Response, error) {
	req := domain.GetServiceReq{
		Name: pbReq.GetName(),
	}

	s, err := impl.servicesService.Get(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting service info")
	}

	about := &pb.AboutService{
		Description:  s.About.Description,
		OriginalName: s.About.OriginalName,
		Env:          s.About.Env,
		ServiceType:  s.About.ServiceType,
		Team:         s.About.Team,
		Repo:         s.About.Repo,
		Port:         s.About.Port,
	}

	return &pb.GetService_Response{
		Payload: &pb.GetService_Response_VervService{
			VervService: &pb.VervAppService{
				Name:                s.Name,
				CurrentDeploymentId: s.CurrentDeploymentId,
				Status:              s.Status,
			},
		},
		About: about,
	}, nil
}
