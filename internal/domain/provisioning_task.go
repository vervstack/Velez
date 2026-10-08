package domain

import (
	"time"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

type ProvisioningTask struct {
	TaskId    int64
	EntityId  string
	Action    string
	Status    velez_api.ProvisioningTask_Status
	Error     string
	UpdatedAt time.Time
	Jobs      []ProvisioningJob
}

type ProvisioningJob struct {
	Name   string
	Status velez_api.ProvisioningTask_Status
}
