package velez_api_impl

import (
	"context"
	"encoding/json"
	"time"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

// createSmerdWatchTimeout bounds how long the synchronous CreateSmerd RPC
// blocks waiting for the create_smerd task to reach a terminal status. It's
// a safety net against the task engine ever getting stuck (e.g. the
// unmarshal-error-before-FinishTask bug fixed in internal/jobs/worker.go),
// not something normal operation should ever hit - sized loosely above the
// slowest observed real run (~17s for postgres:16 under full parallel
// Docker contention in the e2e suite).
const createSmerdWatchTimeout = 60 * time.Second

func (impl *Impl) CreateSmerd(ctx context.Context, req *velez_api.CreateSmerd_Request) (*velez_api.Smerd, error) {
	initialContext := &velez_api.CreateSmerdTaskPayload{}
	initialContext.SetRequest(req)

	_, err := impl.jobsEngine.Enqueue(ctx, req.GetName(), jobs.CreateSmerdAction, initialContext)
	if err != nil {
		return nil, rerrors.Wrap(err, "error enqueuing create_smerd task")
	}

	watchCtx, cancel := context.WithTimeout(ctx, createSmerdWatchTimeout)
	defer cancel()

	var finalTask tasks_queries.VelezTask
	for task := range impl.jobsEngine.Watch(watchCtx, req.GetName(), jobs.CreateSmerdAction) {
		finalTask = task
	}

	isDone := finalTask.Status == tasks_queries.VelezTaskStatusDONE
	isFailed := finalTask.Status == tasks_queries.VelezTaskStatusFAILED

	if !isDone && !isFailed && watchCtx.Err() != nil {
		return nil, rerrors.Wrapf(
			watchCtx.Err(),
			"timed out waiting for create_smerd task, last status: %q",
			finalTask.Status,
		)
	}

	if finalTask.Status == tasks_queries.VelezTaskStatusFAILED {
		return nil, rerrors.New(finalTask.Error.String)
	}

	// Unmarshal only container_id rather than the full CreateSmerdTaskPayload -
	// narrower and avoids depending on the rest of the payload's shape.
	var result struct {
		ContainerId *string `json:"container_id,omitempty"`
	}

	err = json.Unmarshal(finalTask.Context.RawMessage, &result)
	if err != nil {
		return nil, rerrors.Wrap(err, "error unmarshaling task context")
	}

	var containerID string
	if result.ContainerId != nil {
		containerID = *result.ContainerId
	}

	smerd, err := impl.smerdService.InspectSmerd(ctx, containerID)
	if err != nil {
		return nil, rerrors.Wrap(err, "error inspecting smerd")
	}

	return smerd, nil
}
