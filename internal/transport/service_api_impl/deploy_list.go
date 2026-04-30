package service_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
	"go.vervstack.ru/Velez/internal/transport/common"
)

func (impl *Impl) ListDeployments(ctx context.Context, pbReq *pb.ListDeployments_Request) (*pb.ListDeployments_Response, error) {
	req := domain.ListDeploymentsReq{
		Paging: common.FromPaging(pbReq.GetPaging()),
	}

	if pbReq.ServiceId != nil {
		req.ServiceIds = []int64{int64(*pbReq.ServiceId)}
	}

	list, err := impl.servicesService.ListDeployments(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	resp := &pb.ListDeployments_Response{
		Total:       list.Total,
		Deployments: make([]*pb.DeploymentInfo, 0, len(list.Deployments)),
	}

	for _, d := range list.Deployments {
		resp.Deployments = append(resp.Deployments, &pb.DeploymentInfo{
			Id:        uint64(d.Id),
			Status:    toDeploymentStatus(d.Status),
			SpecId:    uint64(d.SpecId),
			CreatedAt: timestamppb.New(d.CreatedAt),
		})
	}

	return resp, nil
}

func toDeploymentStatus(s deployments_queries.VelezDeploymentStatus) pb.DeploymentStatus {
	switch s {
	case deployments_queries.VelezDeploymentStatusSCHEDULEDDEPLOYMENT:
		return pb.DeploymentStatus_SCHEDULED_DEPLOYMENT
	case deployments_queries.VelezDeploymentStatusSCHEDULEDDELETION:
		return pb.DeploymentStatus_SCHEDULED_DELETION
	case deployments_queries.VelezDeploymentStatusSCHEDULEDUPGRADE:
		return pb.DeploymentStatus_SCHEDULED_UPGRADE
	case deployments_queries.VelezDeploymentStatusRUNNING:
		return pb.DeploymentStatus_RUNNING
	case deployments_queries.VelezDeploymentStatusFAILED:
		return pb.DeploymentStatus_FAILED
	case deployments_queries.VelezDeploymentStatusDELETED:
		return pb.DeploymentStatus_DELETED
	default:
		return pb.DeploymentStatus_DEPLOYMENT_STATUS_UNKNOWN
	}
}
