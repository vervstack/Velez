package jobs

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service"
)

const (
	DropHaproxyInstanceAction = "drop_haproxy_instance"

	stepRemoveHaproxyService = "remove_service"
	stepRemoveHaproxyVolume  = "remove_config_volume"
	stepRemoveHaproxySidecar = "remove_vcn_sidecar"
)

type dropHaproxyInstanceHandler struct {
	vervServices service.VervServicesService
	runtimes     container_runtime.RuntimeResolver
	portManager  node_clients.PortManager
}

func NewDropHaproxyInstanceHandler(
	vervServices service.VervServicesService,
	runtimes container_runtime.RuntimeResolver,
	portManager node_clients.PortManager,
) TaskHandler {
	return &dropHaproxyInstanceHandler{vervServices: vervServices, runtimes: runtimes, portManager: portManager}
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
			Name: stepRemoveHaproxySidecar,
			Job: &removeHaproxySidecarJob{
				runtimes: h.runtimes, name: payload.GetName(), portManager: h.portManager,
			},
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

type removeHaproxySidecarJob struct {
	runtimes    container_runtime.RuntimeResolver
	portManager node_clients.PortManager
	name        string
}

func (j *removeHaproxySidecarJob) Do(ctx context.Context) error {
	runtime, err := j.runtimes.Runtime(ctx, "")
	if err != nil {
		return rerrors.Wrap(err)
	}

	name := haproxySidecarName(j.name)

	info, isFound, err := runtime.Inspect(ctx, name)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if isFound && info.Config != nil && info.Config.Labels[labels.HaproxySidecarLabel] == j.name {
		ports := haproxyPublishedPorts(info.HostConfig.PortBindings)

		err = runtime.Remove(ctx, info.ID)
		if err != nil {
			return rerrors.Wrap(err)
		}

		j.portManager.UnlockPorts(ports)

		err = runtime.RemoveVolume(ctx, name)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	return nil
}
