package transfer_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func (impl *Impl) PlanTransfer(
	ctx context.Context,
	req *pb.PlanTransfer_Request,
) (*pb.PlanTransfer_Response, error) {
	plan, err := impl.transfersOut.Plan(ctx, req.GetContainerName())
	if err != nil {
		return nil, rerrors.Wrap(err, "error planning transfer")
	}

	return planToPb(plan), nil
}

func planToPb(plan domain.TransferPlan) *pb.PlanTransfer_Response {
	sidecars := make([]*pb.PlanTransfer_SidecarCandidate, 0, len(plan.Sidecars))
	for _, sidecar := range plan.Sidecars {
		sidecars = append(sidecars, &pb.PlanTransfer_SidecarCandidate{
			ContainerName: sidecar.ContainerName,
			ImageName:     sidecar.ImageName,
		})
	}

	consumers := make([]*pb.PlanTransfer_NetworkConsumer, 0, len(plan.Consumers))
	for _, consumer := range plan.Consumers {
		consumers = append(consumers, &pb.PlanTransfer_NetworkConsumer{
			ContainerName: consumer.ContainerName,
			NetworkName:   consumer.NetworkName,
		})
	}

	return &pb.PlanTransfer_Response{
		Sidecars:          sidecars,
		Consumers:         consumers,
		UnsupportedMounts: plan.UnsupportedMounts,
	}
}
