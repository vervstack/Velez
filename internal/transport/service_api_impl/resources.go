package service_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) GetServiceResources(
	ctx context.Context,
	pbReq *pb.GetServiceResources_Request,
) (*pb.GetServiceResources_Response, error) {
	resources, err := impl.servicesService.GetServiceResources(ctx, pbReq.GetServiceName())
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting service resources")
	}

	pbResources := make([]*pb.BoundResource, 0, len(resources))
	for _, r := range resources {
		pbResource := &pb.BoundResource{
			Name:         r.Name,
			ResourceType: r.ResourceType,
			Status:       r.Status,
		}

		pbResources = append(pbResources, pbResource)
	}

	resp := &pb.GetServiceResources_Response{
		Resources: pbResources,
	}

	return resp, nil
}
