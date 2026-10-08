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

	sidecars := make([]*pb.ServiceSidecar, 0, len(s.Sidecars))
	for _, sidecar := range s.Sidecars {
		sidecars = append(sidecars, &pb.ServiceSidecar{
			ContainerId:   sidecar.ContainerId,
			ContainerName: sidecar.ContainerName,
			ImageName:     sidecar.ImageName,
			Status:        sidecar.Status,
		})
	}

	var proxyUrl *string

	if s.ProxyUrl != "" {
		proxyUrl = &s.ProxyUrl
	}

	return &pb.GetService_Response{
		Payload: &pb.GetService_Response_VervService{
			VervService: &pb.VervAppService{
				Name:                s.Name,
				DisplayName:         displayNameOrName(s.DisplayName, s.Name),
				CurrentDeploymentId: s.CurrentDeploymentId,
				Status:              s.Status,
				Labels:              s.Labels,
			},
		},
		About:            about,
		Sidecars:         sidecars,
		ProxyUrl:         proxyUrl,
		ProxyBypassHosts: s.ProxyBypassHosts,
	}, nil
}
