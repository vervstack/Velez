package verv_services

import (
	"context"
	"time"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

func (v *VervService) GetServiceMetrics(ctx context.Context, serviceName string) (domain.ServiceMetrics, error) {
	req := &velez_api.ListSmerds_Request{
		Label: map[string]string{labels.VervServiceLabel: serviceName},
	}
	resp, err := v.containerService.ListSmerds(ctx, req)
	if err != nil {
		return domain.ServiceMetrics{}, rerrors.Wrap(err)
	}

	metrics := domain.ServiceMetrics{
		ReplicasDesired: uint32(len(resp.Smerds)),
	}

	if len(resp.Smerds) == 0 {
		return metrics, nil
	}

	var (
		totalCPU            float64
		maxMemory           uint64
		replicasRunning     uint32
		earliestStartedTime *time.Time
	)

	for _, smerd := range resp.Smerds {
		if smerd.Status != velez_api.Smerd_running {
			continue
		}

		stats, err := v.docker.Stats(ctx, smerd.Name)
		if err != nil {
			continue
		}

		totalCPU += stats.CPUPercent
		if stats.MemUsageMi > maxMemory {
			maxMemory = stats.MemUsageMi
		}
		replicasRunning++

		if earliestStartedTime == nil || stats.StartedAt.Before(*earliestStartedTime) {
			earliestStartedTime = &stats.StartedAt
		}
	}

	metrics.CPUPercent = totalCPU
	metrics.MemMi = maxMemory
	metrics.MemMaxMi = maxMemory
	metrics.ReplicasRunning = replicasRunning

	if earliestStartedTime != nil {
		metrics.UptimeSeconds = uint64(time.Since(*earliestStartedTime).Seconds())
	}

	return metrics, nil
}
