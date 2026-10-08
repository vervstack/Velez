package common

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func ProvisioningTasksToPb(tasks []domain.ProvisioningTask) []*pb.ProvisioningTask {
	out := make([]*pb.ProvisioningTask, 0, len(tasks))
	for _, task := range tasks {
		out = append(out, provisioningTaskToPb(task))
	}

	return out
}

func provisioningTaskToPb(task domain.ProvisioningTask) *pb.ProvisioningTask {
	jobs := make([]*pb.ProvisioningTask_Job, 0, len(task.Jobs))
	for _, job := range task.Jobs {
		pbJob := &pb.ProvisioningTask_Job{
			Name:   job.Name,
			Status: job.Status,
		}

		jobs = append(jobs, pbJob)
	}

	out := &pb.ProvisioningTask{
		TaskId:    task.TaskId,
		EntityId:  task.EntityId,
		Action:    task.Action,
		Status:    task.Status,
		UpdatedAt: timestamppb.New(task.UpdatedAt),
		Jobs:      jobs,
	}

	if task.Error != "" {
		out.Error = &task.Error
	}

	return out
}
