package jobs

import (
	"context"
	"fmt"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
)

const (
	DropSmerdAction = "drop_smerd"
)

// dropResultAccessor is the narrow slice of *velez_api.DropSmerdTaskPayload
// dropContainerJob needs to record its per-identifier outcome.
// *velez_api.DropSmerdTaskPayload satisfies it.
type dropResultAccessor interface {
	AppendFailed(*velez_api.DropSmerd_Response_Error)
	AppendSuccessful(string)
}

// dropSmerdHandler is the leanest dependency footprint of any handler in
// this package - dropping containers only ever needs Docker access.
type dropSmerdHandler struct {
	nodeClients node_clients.NodeClients
}

func NewDropSmerdHandler(nodeClients node_clients.NodeClients) TaskHandler {
	return &dropSmerdHandler{
		nodeClients: nodeClients,
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

	namedJobs := make([]NamedJob, 0, len(worklist))

	for i, identifier := range worklist {
		job := &dropContainerJob{
			docker:     h.nodeClients.Docker(),
			identifier: identifier,
			ctx:        payload,
		}

		name := fmt.Sprintf("drop_container_%d", i)

		namedJobs = append(namedJobs, NamedJob{Name: name, Job: job})
	}

	return namedJobs
}

// dropContainerJob removes a single container by uuid or name. It mirrors
// container_manager.DropSmerds' per-item behavior exactly: Docker.Remove is
// already idempotent (treats "no such container" as success), and any other
// error is recorded on the task context as a per-item failure rather than
// propagated as a job error.
type dropContainerJob struct {
	docker     node_clients.Docker
	identifier string
	ctx        dropResultAccessor
}

// Do intentionally always returns nil, even when j.docker.Remove fails. This
// is NOT a bug: the old container_manager.DropSmerds RPC always returned a
// nil top-level error and reported every per-item failure only inside its
// response body's Failed slice. Returning a real error here would make this
// job (and therefore the drop_smerd task) reach FAILED on any single
// container's removal error, which would change DropSmerd's response
// contract for existing callers - a backward-compatibility break this
// repo's CLAUDE.md forbids. Do not "fix" this into propagating errors.
func (j *dropContainerJob) Do(ctx context.Context) error {
	err := j.docker.Remove(ctx, j.identifier)
	if err != nil {
		failure := &velez_api.DropSmerd_Response_Error{
			Uuid:  j.identifier,
			Cause: err.Error(),
		}
		j.ctx.AppendFailed(failure)

		return nil
	}

	j.ctx.AppendSuccessful(j.identifier)

	return nil
}
