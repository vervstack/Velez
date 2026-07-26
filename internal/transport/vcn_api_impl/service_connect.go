package vcn_api_impl

import (
	"context"
	"time"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

// connectServiceWatchTimeout bounds how long the synchronous ConnectService
// RPC blocks waiting for the connect_service_to_vpn task to reach a terminal
// status. Same safety-net rationale as createSmerdWatchTimeout/
// assembleConfigWatchTimeout - not something normal operation should ever
// hit.
const (
	connectServiceWatchTimeout = 60 * time.Second
)

func (impl *Impl) ConnectService(ctx context.Context, req *velez_api.ConnectService_Request) (
	*velez_api.ConnectService_Response, error,
) {
	initialContext := &velez_api.ConnectServiceToVpnTaskPayload{
		ServiceName: req.GetServiceName(),
	}

	_, err := impl.jobsEngine.Enqueue(ctx, req.GetServiceName(), jobs.ConnectServiceToVpnAction, initialContext)
	if err != nil {
		return nil, rerrors.Wrap(err, "error enqueuing connect_service_to_vpn task")
	}

	watchCtx, cancel := context.WithTimeout(ctx, connectServiceWatchTimeout)
	defer cancel()

	var finalTask tasks_queries.VelezTask

	for task := range impl.jobsEngine.Watch(watchCtx, req.GetServiceName(), jobs.ConnectServiceToVpnAction) {
		finalTask = task
	}

	isDone := finalTask.Status == tasks_queries.VelezTaskStatusDONE
	isFailed := finalTask.Status == tasks_queries.VelezTaskStatusFAILED

	if !isDone && !isFailed && watchCtx.Err() != nil {
		return nil, rerrors.Wrapf(
			watchCtx.Err(),
			"timed out waiting for connect_service_to_vpn task, last status: %q",
			finalTask.Status,
		)
	}

	if finalTask.Status == tasks_queries.VelezTaskStatusFAILED {
		return nil, rerrors.New(finalTask.Error.String)
	}

	return &velez_api.ConnectService_Response{}, nil
}
