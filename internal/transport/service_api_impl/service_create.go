package service_api_impl

import (
	"context"
	"time"

	"go.redsock.ru/rerrors"
	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

// createServiceWatchTimeout bounds how long the synchronous CreateService RPC
// blocks waiting for the create_service task to reach a terminal status. It's
// a safety net against the task engine ever getting stuck, not something
// normal operation should ever hit - create_service is a 2-step, DB-only
// task (validate_name, upsert_service) so it should complete far faster than
// this.
const (
	createServiceWatchTimeout = 60 * time.Second
)

func (impl *Impl) CreateService(
	ctx context.Context,
	apiReq *pb.CreateService_Request,
) (*pb.CreateService_Response, error) {
	initialContext := &pb.CreateServiceTaskPayload{
		Name: apiReq.GetName(),
	}

	_, err := impl.jobsEngine.Enqueue(ctx, apiReq.GetName(), jobs.CreateServiceAction, initialContext)
	if err != nil {
		return nil, rerrors.Wrap(err, "error enqueuing create_service task")
	}

	watchCtx, cancel := context.WithTimeout(ctx, createServiceWatchTimeout)
	defer cancel()

	var finalTask tasks_queries.VelezTask

	for task := range impl.jobsEngine.Watch(watchCtx, apiReq.GetName(), jobs.CreateServiceAction) {
		finalTask = task
	}

	isDone := finalTask.Status == tasks_queries.VelezTaskStatusDONE
	isFailed := finalTask.Status == tasks_queries.VelezTaskStatusFAILED

	if !isDone && !isFailed && watchCtx.Err() != nil {
		return nil, rerrors.Wrapf(
			watchCtx.Err(),
			"timed out waiting for create_service task, last status: %q",
			finalTask.Status,
		)
	}

	if finalTask.Status == tasks_queries.VelezTaskStatusFAILED {
		return nil, rerrors.New(finalTask.Error.String)
	}

	return &pb.CreateService_Response{}, nil
}
