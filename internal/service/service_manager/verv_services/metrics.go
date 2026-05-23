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

	// single-mode fallback: find container by name when VERV_SERVICE label is absent
	if len(resp.Smerds) == 0 {
		req = &velez_api.ListSmerds_Request{Name: &serviceName}
		resp, err = v.containerService.ListSmerds(ctx, req)
		if err != nil {
			return domain.ServiceMetrics{}, rerrors.Wrap(err)
		}
	}

	metrics := domain.ServiceMetrics{
		ReplicasDesired: uint32(len(resp.Smerds)),
	}

	if len(resp.Smerds) == 0 {
		return metrics, nil
	}

	var (
		totalCPU            float64
		maxMemUsage         uint64
		maxMemLimit         uint64
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
		if stats.MemUsageMi > maxMemUsage {
			maxMemUsage = stats.MemUsageMi
		}
		if stats.MemLimitMi > maxMemLimit {
			maxMemLimit = stats.MemLimitMi
		}
		replicasRunning++

		if earliestStartedTime == nil || stats.StartedAt.Before(*earliestStartedTime) {
			earliestStartedTime = &stats.StartedAt
		}
	}

	metrics.CPUPercent = totalCPU
	metrics.MemMi = maxMemUsage
	metrics.MemMaxMi = maxMemLimit
	metrics.ReplicasRunning = replicasRunning

	if earliestStartedTime != nil {
		metrics.UptimeSeconds = uint64(time.Since(*earliestStartedTime).Seconds())
	}

	return metrics, nil
}
