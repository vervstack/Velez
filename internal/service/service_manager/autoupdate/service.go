package autoupdate

import (
	"context"

	"github.com/docker/docker/client"

	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

type TaskEnqueuer interface {
	Enqueue(ctx context.Context, entityID, action string, initialContext any) (tasks_queries.VelezTask, error)
}

type Service struct {
	dockerAPI  client.APIClient
	jobsEngine TaskEnqueuer

	isWorkerEnabled    bool
	checkPeriodSeconds int64
}

func New(
	dockerAPI client.APIClient,
	jobsEngine TaskEnqueuer,
	isWorkerEnabled bool,
	checkPeriodSeconds int,
) service.AutoUpdateService {
	return &Service{
		dockerAPI:  dockerAPI,
		jobsEngine: jobsEngine,

		isWorkerEnabled:    isWorkerEnabled,
		checkPeriodSeconds: int64(checkPeriodSeconds),
	}
}
