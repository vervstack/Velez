package jobs

import (
	"context"
	"fmt"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
)

const (
	DropSmerdAction = "drop_smerd"
)

// dropResultAccessor is the narrow slice of *velez_api.DropSmerdTaskPayload
// dropContainerJob needs to record its per-identifier outcome.
// *velez_api.DropSmerdTaskPayload satisfies it.
type dropResultAccessor interface {
	AppendFailed(err *velez_api.DropSmerd_Response_Error)
	AppendSuccessful(msg string)
}

// dropSmerdHandler resolves the request's environment into the
// ContainerRuntime that serves it (see docs/container_runtimes) and removes
// containers through it, rather than talking to node_clients.Docker
// directly - this is what lets each removal be scoped to the environment the
// DropSmerd request named, instead of ignoring it entirely (see
// docs/container_runtimes/roadmap.md's "DropSmerd ignores environment scope
// entirely" bug).
type dropSmerdHandler struct {
	runtimes container_runtime.RuntimeResolver
}

func NewDropSmerdHandler(runtimes container_runtime.RuntimeResolver) TaskHandler {
	return &dropSmerdHandler{
		runtimes: runtimes,
	}
}

func (h *dropSmerdHandler) Action() string {
	return DropSmerdAction
}

func (h *dropSmerdHandler) NewContext() TaskContext {
	return &velez_api.DropSmerdTaskPayload{}
}

// BuildJobs mirrors container_manager.DropSmerds' own
// append(req.Uuids, req.Name...) worklist - uuids first, then names, both
// treated as interchangeable Docker container identifiers. Unlike
// copy_to_volume.go's map-derived file paths, proto repeated fields already
// have a stable, deterministic order, so no sorting is needed here for
// resume-safety. Job names are index-based (drop_container_<n>) rather than
// identifier-based since raw identifiers (container names/uuids) aren't safe
// job-name material.
func (h *dropSmerdHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.DropSmerdTaskPayload)
	if !ok {
		panic("drop_smerd: BuildJobs called with mismatched TaskContext type")
	}

	req := payload.GetRequest()
	worklist := append(req.GetUuids(), req.GetName()...)
	environment := req.GetEnvironment()

	namedJobs := make([]NamedJob, 0, len(worklist))

	for i, identifier := range worklist {
		job := &dropContainerJob{
			runtimes:    h.runtimes,
			environment: environment,
			identifier:  identifier,
			ctx:         payload,
		}

		name := fmt.Sprintf("drop_container_%d", i)

		namedJobs = append(namedJobs, NamedJob{Name: name, Job: job})
	}

	return namedJobs
}

// dropContainerJob removes a single container by uuid or name, through the
// ContainerRuntime resolved for environment. It mirrors
// container_manager.DropSmerds' per-item behavior exactly: removal is
// idempotent (labelBasedRuntime.Remove treats "no such container" as
// success), and any other error - including a failure to resolve environment
// itself - is recorded on the task context as a per-item failure rather than
// propagated as a job error.
type dropContainerJob struct {
	runtimes container_runtime.RuntimeResolver

	environment string
	identifier  string
	ctx         dropResultAccessor
}

// Do intentionally always returns nil, even when removal fails. This is NOT a
// bug: the old container_manager.DropSmerds RPC always returned a nil
// top-level error and reported every per-item failure only inside its
// response body's Failed slice. Returning a real error here would make this
// job (and therefore the drop_smerd task) reach FAILED on any single
// container's removal error, which would change DropSmerd's response
// contract for existing callers - a backward-compatibility break this
// repo's CLAUDE.md forbids. Do not "fix" this into propagating errors.
func (j *dropContainerJob) Do(ctx context.Context) error {
	containerRuntime, err := j.runtimes.Runtime(ctx, j.environment)
	if err == nil {
		err = containerRuntime.Remove(ctx, j.identifier)
	}

	if err != nil {
		failure := &velez_api.DropSmerd_Response_Error{
			Uuid:  j.identifier,
			Cause: err.Error(),
		}
		j.ctx.AppendFailed(failure)
	} else {
		j.ctx.AppendSuccessful(j.identifier)
	}

	return nil
}
