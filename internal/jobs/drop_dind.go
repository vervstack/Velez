package jobs

import (
	"context"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/storage"
)

const (
	DropDindAction = "drop_dind"

	stepDeleteDindRow     = "delete_dind_row"
	stepRemoveDindService = "remove_dind_service"
	stepRemoveDindNetwork = "remove_dind_network"
	stepRemoveDindVolume  = "remove_dind_volume"
)

type dropDindHandler struct {
	dataStorage  storage.Storage
	vervServices service.VervServicesService
	docker       node_clients.Docker
}

func NewDropDindHandler(
	dataStorage storage.Storage,
	vervServices service.VervServicesService,
	docker node_clients.Docker,
) TaskHandler {
	return &dropDindHandler{
		dataStorage:  dataStorage,
		vervServices: vervServices,
		docker:       docker,
	}
}

func (h *dropDindHandler) Action() string {
	return DropDindAction
}

func (h *dropDindHandler) NewContext() TaskContext {
	return &velez_api.DropDindTaskPayload{}
}

// BuildJobs deletes the dind row before removing the service: the row is
// addressed by the service's id, which a re-run can no longer look up once the
// service is gone. The network and the data volume are keyed by name alone.
func (h *dropDindHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.DropDindTaskPayload)
	if !ok {
		panic("drop_dind: BuildJobs called with mismatched TaskContext type")
	}

	return []NamedJob{
		{
			Name: stepDeleteDindRow,
			Job: &deleteDindRowJob{
				services:      h.dataStorage.Services(),
				dindInstances: h.dataStorage.DindInstances(),
				req:           payload,
			},
		},
		{
			Name: stepRemoveDindService,
			Job: &removeDindServiceJob{
				vervServices: h.vervServices,
				req:          payload,
			},
		},
		{
			Name: stepRemoveDindNetwork,
			Job: &removeDindNetworkJob{
				docker: h.docker,
				req:    payload,
			},
		},
		{
			Name: stepRemoveDindVolume,
			Job: &removeDindVolumeJob{
				docker: h.docker,
				req:    payload,
			},
		},
	}
}

type deleteDindRowJob struct {
	services      storage.ServicesStorage
	dindInstances storage.DindInstancesStorage

	req serviceNameAccessor
}

func (j *deleteDindRowJob) Do(ctx context.Context) error {
	svc, isFound, err := findServiceByName(ctx, j.services, j.req.GetName())
	if err != nil {
		return rerrors.Wrap(err)
	}

	if !isFound {
		return nil
	}

	err = j.dindInstances.DeleteDindInstance(ctx, svc.ID)
	if err != nil {
		return rerrors.Wrap(err, "error deleting dind instance row")
	}

	return nil
}

type removeDindServiceJob struct {
	vervServices service.VervServicesService

	req serviceNameAccessor
}

func (j *removeDindServiceJob) Do(ctx context.Context) error {
	removeReq := domain.RemoveServiceReq{Name: j.req.GetName(), DropRunningInstances: true}

	err := j.vervServices.Remove(ctx, removeReq)
	if err != nil {
		return rerrors.Wrap(err, "error removing dind service")
	}

	return nil
}

type removeDindNetworkJob struct {
	docker node_clients.Docker

	req serviceNameAccessor
}

func (j *removeDindNetworkJob) Do(ctx context.Context) error {
	name := j.req.GetName()

	err := j.docker.Client().NetworkRemove(ctx, domain.DindNetworkName(name))
	if err != nil && !cerrdefs.IsNotFound(err) {
		log.Ctx(ctx).Warn().
			Str("dind_name", name).
			Err(err).
			Msg("error removing dind network")
	}

	return nil
}

type removeDindVolumeJob struct {
	docker node_clients.Docker

	req serviceNameAccessor
}

func (j *removeDindVolumeJob) Do(ctx context.Context) error {
	name := j.req.GetName()

	err := j.docker.Client().VolumeRemove(ctx, domain.DindDataVolumeName(name), false)
	if err != nil && !cerrdefs.IsNotFound(err) {
		log.Ctx(ctx).Warn().
			Str("dind_name", name).
			Err(err).
			Msg("error removing dind data volume")
	}

	return nil
}
