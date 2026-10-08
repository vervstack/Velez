package provisioning

import (
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/jobs_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

func toDomainTask(entry jobs.ProvisioningEntry) domain.ProvisioningTask {
	task := entry.Task

	jobStatuses := make([]domain.ProvisioningJob, 0, len(entry.Jobs))
	for _, job := range entry.Jobs {
		jobStatuses = append(jobStatuses, domain.ProvisioningJob{
			Name:   job.Name,
			Status: toJobStatus(job.Status),
		})
	}

	return domain.ProvisioningTask{
		TaskId:    task.ID,
		EntityId:  task.EntityID,
		Action:    task.Action,
		Status:    toTaskStatus(task.Status),
		Error:     task.Error.String,
		UpdatedAt: task.UpdatedAt,
		Jobs:      jobStatuses,
	}
}

func toTaskStatus(status tasks_queries.VelezTaskStatus) velez_api.ProvisioningTask_Status {
	switch status {
	case tasks_queries.VelezTaskStatusPENDING:
		return velez_api.ProvisioningTask_PENDING
	case tasks_queries.VelezTaskStatusRUNNING:
		return velez_api.ProvisioningTask_RUNNING
	case tasks_queries.VelezTaskStatusDONE:
		return velez_api.ProvisioningTask_DONE
	case tasks_queries.VelezTaskStatusFAILED:
		return velez_api.ProvisioningTask_FAILED
	default:
		return velez_api.ProvisioningTask_STATUS_UNSPECIFIED
	}
}

func toJobStatus(status jobs_queries.VelezJobStatus) velez_api.ProvisioningTask_Status {
	switch status {
	case jobs_queries.VelezJobStatusRUNNING:
		return velez_api.ProvisioningTask_RUNNING
	case jobs_queries.VelezJobStatusDONE:
		return velez_api.ProvisioningTask_DONE
	case jobs_queries.VelezJobStatusFAILED:
		return velez_api.ProvisioningTask_FAILED
	default:
		return velez_api.ProvisioningTask_PENDING
	}
}
