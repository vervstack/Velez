package jobs

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service"
)

const (
	DropHaproxyInstanceAction = "drop_haproxy_instance"

	stepRemoveHaproxyService = "remove_service"
	stepRemoveHaproxyVolume  = "remove_config_volume"
)

type dropHaproxyInstanceHandler struct {
	vervServices service.VervServicesService
	runtimes     container_runtime.RuntimeResolver
}

func NewDropHaproxyInstanceHandler(
	vervServices service.VervServicesService,
	runtimes container_runtime.RuntimeResolver,
) TaskHandler {
	return &dropHaproxyInstanceHandler{vervServices: vervServices, runtimes: runtimes}
}

func (h *dropHaproxyInstanceHandler) Action() string {
	return DropHaproxyInstanceAction
}

func (h *dropHaproxyInstanceHandler) NewContext() TaskContext {
	return &velez_api.DropHaproxyInstanceTaskPayload{}
}

func (h *dropHaproxyInstanceHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.DropHaproxyInstanceTaskPayload)
	if !ok {
		panic("drop_haproxy_instance: BuildJobs called with mismatched TaskContext type")
	}

	return []NamedJob{
		{
			Name: stepRemoveHaproxyService,
			Job:  &removeHaproxyServiceJob{vervServices: h.vervServices, payload: payload},
		},
		{
			Name: stepRemoveHaproxyVolume,
			Job:  &removeHaproxyVolumeJob{runtimes: h.runtimes, payload: payload},
		},
	}
}

type removeHaproxyServiceJob struct {
	vervServices service.VervServicesService
	payload      *velez_api.DropHaproxyInstanceTaskPayload
}

func (j *removeHaproxyServiceJob) Do(ctx context.Context) error {
	req := domain.RemoveServiceReq{
		Name:                 domain.HaproxyServiceName(j.payload.GetName()),
		DropRunningInstances: true,
		Environment:          j.payload.GetEnvironment(),
	}

	err := j.vervServices.Remove(ctx, req)
	if err != nil {
		return rerrors.Wrap(err, "error removing HAProxy service")
	}

	return nil
}

type removeHaproxyVolumeJob struct {
	runtimes container_runtime.RuntimeResolver
	payload  *velez_api.DropHaproxyInstanceTaskPayload
}

func (j *removeHaproxyVolumeJob) Do(ctx context.Context) error {
	runtime, err := j.runtimes.Runtime(ctx, j.payload.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving HAProxy container runtime")
	}

	err = runtime.RemoveVolume(ctx, domain.HaproxyConfigVolumeName(j.payload.GetName()))
	if err != nil {
		return rerrors.Wrap(err, "error removing HAProxy config volume")
	}

	return nil
}
