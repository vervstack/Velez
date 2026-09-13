package dockerutils

import (
	"context"
	"encoding/json"
	"time"

	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

const (
	percentageMultiplier = 100.0
	bytesPerMB           = 1024 * 1024
)

// Stats fetches a single non-streaming stats snapshot for nameOrID and
// derives domain.ContainerStats from it - the exact computation
// docker.Docker.Stats has always done, extracted here so
// container_runtime.labelBasedRuntime.Stats can reuse it without a second,
// drifting copy of the CPU-percent/memory math.
func Stats(ctx context.Context, cli client.APIClient, nameOrID string) (domain.ContainerStats, error) {
	resp, err := cli.ContainerStats(ctx, nameOrID, false)
	if err != nil {
		return domain.ContainerStats{}, rerrors.Wrap(err, "error getting container stats")
	}

	defer func() { _ = resp.Body.Close() }()

	var stats struct {
		CPUStats struct {
			CPUUsage struct {
				TotalUsage  uint64   `json:"total_usage"`
				PercpuUsage []uint64 `json:"percpu_usage"`
			} `json:"cpu_usage"`
			SystemUsage uint64 `json:"system_cpu_usage"`
			OnlineCPUs  int64  `json:"online_cpus"`
		} `json:"cpu_stats"`
		PreCPUStats struct {
			CPUUsage struct {
				TotalUsage uint64 `json:"total_usage"`
			} `json:"cpu_usage"`
			SystemUsage uint64 `json:"system_cpu_usage"`
		} `json:"precpu_stats"`
		MemoryStats struct {
			Usage uint64 `json:"usage"`
			Limit uint64 `json:"limit"`
		} `json:"memory_stats"`
	}

	err = json.NewDecoder(resp.Body).Decode(&stats)
	if err != nil {
		return domain.ContainerStats{}, rerrors.Wrap(err, "error decoding stats")
	}

	cpuPercent := 0.0
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)

	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)
	if cpuDelta > 0 && systemDelta > 0 {
		numCPU := stats.CPUStats.OnlineCPUs
		if numCPU == 0 {
			numCPU = int64(len(stats.CPUStats.CPUUsage.PercpuUsage))
		}

		cpuPercent = (cpuDelta / systemDelta) * float64(numCPU) * percentageMultiplier
	}

	memUsageMi := stats.MemoryStats.Usage / bytesPerMB
	memLimitMi := stats.MemoryStats.Limit / bytesPerMB

	inspectResp, err := cli.ContainerInspect(ctx, nameOrID)
	if err != nil {
		return domain.ContainerStats{}, rerrors.Wrap(err, "error inspecting container")
	}

	var startedAt time.Time

	startedAtStr := inspectResp.State.StartedAt
	if startedAtStr != "" {
		t, err := time.Parse(time.RFC3339, startedAtStr)
		if err == nil {
			startedAt = t
		}
	}

	return domain.ContainerStats{
		CpuPercent: cpuPercent,
		MemUsageMi: memUsageMi,
		MemLimitMi: memLimitMi,
		StartedAt:  startedAt,
	}, nil
}
