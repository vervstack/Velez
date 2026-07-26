package domain

import (
	"time"
)

type ContainerStats struct {
	CPUPercent float64
	MemUsageMi uint64
	MemLimitMi uint64
	StartedAt  time.Time
}

type ServiceMetrics struct {
	CPUPercent      float64
	MemMi           uint64
	MemMaxMi        uint64
	ReplicasRunning uint32
	ReplicasDesired uint32
	UptimeSeconds   uint64
}

type BoundResource struct {
	Name         string
	ResourceType string
	Status       string
}
