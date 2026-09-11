package velez_api_impl

import (
	"context"
	"time"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

// upgradeSmerdWatchTimeout bounds how long the synchronous UpgradeSmerd RPC
// blocks waiting for the upgrade_smerd task to reach a terminal status. Same
// safety-net role as createSmerdWatchTimeout in smerd_create.go, sized
// higher: upgrade_smerd's 15 jobs do roughly twice the Docker work of
// create_smerd's 9 (two container creates - the scratch config-fetcher
// container and the final one - plus pause/rename/drop of the old container
// on top of the same pull/config/healthcheck stages), so its slowest real
// run should be expected to run longer than create_smerd's ~17s baseline.
const (
	upgradeSmerdWatchTimeout = 120 * time.Second
)

func (impl *Impl) UpgradeSmerd(ctx context.Context,
	req *velez_api.UpgradeSmerd_Request,
) (*velez_api.UpgradeSmerd_Response, error) {
	suffix, err := impl.resolveEnvironment(ctx, req.GetEnvironment())
	if err != nil {
		return nil, err
	}

	// velez.tasks is UNIQUE (entity_id, action): fold the environment's suffix
	// into the entity id so the same service name upgraded in two environments
	// gets two rows instead of the second deduping onto the first. An empty
	// suffix (default single-environment node) leaves the id as the bare name.
	entityID := jobs.SmerdEntityID(suffix, req.GetName())

	initialContext := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: req,
	}

	_, err = impl.jobsEngine.Enqueue(ctx, entityID, jobs.UpgradeSmerdAction, initialContext)
	if err != nil {
		return nil, rerrors.Wrap(err, "error enqueuing upgrade_smerd task")
	}

	watchCtx, cancel := context.WithTimeout(ctx, upgradeSmerdWatchTimeout)
	defer cancel()

	var finalTask tasks_queries.VelezTask

	for task := range impl.jobsEngine.Watch(watchCtx, entityID, jobs.UpgradeSmerdAction) {
		finalTask = task
	}

	isDone := finalTask.Status == tasks_queries.VelezTaskStatusDONE
	isFailed := finalTask.Status == tasks_queries.VelezTaskStatusFAILED

	if !isDone && !isFailed && watchCtx.Err() != nil {
		return nil, rerrors.Wrapf(
			watchCtx.Err(),
			"timed out waiting for upgrade_smerd task, last status: %q",
			finalTask.Status,
		)
	}

	if finalTask.Status == tasks_queries.VelezTaskStatusFAILED {
		return nil, user_errors.New(finalTask.Error.String)
	}

	return &velez_api.UpgradeSmerd_Response{}, nil
}
