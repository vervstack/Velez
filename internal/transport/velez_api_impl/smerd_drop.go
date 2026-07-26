package velez_api_impl

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

// dropSmerdWatchTimeout bounds how long the synchronous DropSmerd RPC blocks
// waiting for the drop_smerd task to reach a terminal status. See
// createSmerdWatchTimeout's doc comment for why this exists - the same
// safety-net reasoning applies here.
const dropSmerdWatchTimeout = 60 * time.Second

func (impl *Impl) DropSmerd(
	ctx context.Context,
	req *velez_api.DropSmerd_Request,
) (*velez_api.DropSmerd_Response, error) {
	initialContext := &velez_api.DropSmerdTaskPayload{}
	initialContext.SetRequest(req)

	// DropSmerd has no single natural entity - it's a batch of containers,
	// not "the smerd named X" - so each call gets its own synthetic entity
	// ID and therefore its own task. There's no cross-request dedup, which
	// is fine since the underlying per-container removal is already
	// idempotent.
	entityID := uuid.New().String()

	_, err := impl.jobsEngine.Enqueue(ctx, entityID, jobs.DropSmerdAction, initialContext)
	if err != nil {
		return nil, rerrors.Wrap(err, "error enqueuing drop_smerd task")
	}

	watchCtx, cancel := context.WithTimeout(ctx, dropSmerdWatchTimeout)
	defer cancel()

	var finalTask tasks_queries.VelezTask
	for task := range impl.jobsEngine.Watch(watchCtx, entityID, jobs.DropSmerdAction) {
		finalTask = task
	}

	isDone := finalTask.Status == tasks_queries.VelezTaskStatusDONE
	isFailed := finalTask.Status == tasks_queries.VelezTaskStatusFAILED

	if !isDone && !isFailed && watchCtx.Err() != nil {
		return nil, rerrors.Wrapf(
			watchCtx.Err(),
			"timed out waiting for drop_smerd task, last status: %q",
			finalTask.Status,
		)
	}

	// Per dropContainerJob's design, this task should basically never reach
	// FAILED under normal operation - every job swallows its own per-item
	// Docker errors. This branch only guards against engine-level failures
	// (e.g. an enqueue-adjacent bug), handled defensively exactly like
	// smerd_create.go does.
	if finalTask.Status == tasks_queries.VelezTaskStatusFAILED {
		return nil, rerrors.New(finalTask.Error.String)
	}

	// Unmarshal only failed/successful rather than the full
	// DropSmerdTaskPayload - narrower and avoids depending on the rest of
	// the payload's shape.
	var result struct {
		Failed     []*velez_api.DropSmerd_Response_Error `json:"failed,omitempty"`
		Successful []string                              `json:"successful,omitempty"`
	}

	err = json.Unmarshal(finalTask.Context.RawMessage, &result)
	if err != nil {
		return nil, rerrors.Wrap(err, "error unmarshaling task context")
	}

	resp := &velez_api.DropSmerd_Response{
		Failed:     result.Failed,
		Successful: result.Successful,
	}

	return resp, nil
}
