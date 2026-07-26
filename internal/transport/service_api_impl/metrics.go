package service_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) GetServiceMetrics(
	ctx context.Context,
	pbReq *pb.GetServiceMetrics_Request,
) (*pb.GetServiceMetrics_Response, error) {
	metrics, err := impl.servicesService.GetServiceMetrics(ctx, pbReq.GetServiceName())
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	resp := &pb.GetServiceMetrics_Response{
		CpuPercent:      metrics.CPUPercent,
		MemMi:           metrics.MemMi,
		MemMaxMi:        metrics.MemMaxMi,
		ReplicasRunning: metrics.ReplicasRunning,
		ReplicasDesired: metrics.ReplicasDesired,
		UptimeSeconds:   metrics.UptimeSeconds,
	}

	return resp, nil
}
