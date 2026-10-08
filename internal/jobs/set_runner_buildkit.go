package jobs

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

const (
	SetRunnerBuildkitAction = "set_runner_buildkit"
)

type setRunnerBuildkitHandler struct{}

func NewSetRunnerBuildkitHandler() TaskHandler {
	return &setRunnerBuildkitHandler{}
}

func (h *setRunnerBuildkitHandler) Action() string {
	return SetRunnerBuildkitAction
}

func (h *setRunnerBuildkitHandler) NewContext() TaskContext {
	return &velez_api.SetRunnerBuildkitTaskPayload{}
}

func (h *setRunnerBuildkitHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	_, ok := taskCtx.(*velez_api.SetRunnerBuildkitTaskPayload)
	if !ok {
		panic("set_runner_buildkit: BuildJobs called with mismatched TaskContext type")
	}

	return []NamedJob{
		{
			Name: SetRunnerBuildkitAction,
			Job:  &setRunnerBuildkitJob{},
		},
	}
}

type setRunnerBuildkitJob struct{}

func (j *setRunnerBuildkitJob) Do(_ context.Context) error {
	return rerrors.Wrap(errSetRunnerBuildkitNotImplemented)
}
