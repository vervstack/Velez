package workers

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.redsock.ru/rerrors"
	"golang.org/x/sync/errgroup"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/pipelines"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
)

type deployWatcher struct {
	services           service.Services
	pipeliner          pipelines.Pipeliner
	deploymentsStorage storage.DeploymentsStorage
	nodeClients        node_clients.NodeClients

	nodeId int64

	starter  sync.Once
	stopOnce sync.Once
	ticker   *time.Ticker
	done     chan struct{}
}

func NewDeployWatcher(
	services service.Services,
	runner pipelines.Pipeliner,
	clusterClients cluster_clients.ClusterClients,
	nodeClients node_clients.NodeClients,

	interval time.Duration) Worker {
	return &deployWatcher{
		services:           services,
		pipeliner:          runner,
		deploymentsStorage: clusterClients.StateManager().Deployments(),
		nodeClients:        nodeClients,

		nodeId: 1,

		starter: sync.Once{},
		ticker:  time.NewTicker(interval),
		done:    make(chan struct{}),
	}
}

func (d *deployWatcher) Start(ctx context.Context) {
	d.starter.Do(func() {
		for {
			select {
			case <-d.done:
				return
			case <-d.ticker.C:
				list, err := d.listDeployments(ctx)
				if err != nil {
					if !rerrors.Is(err, cluster_clients.ErrServiceIsDisabled) {
						logrus.Error("error listing deployments in deploy watcher: ", err)
						// TODO make it fail only when state is not available
						// retry via api handle
						return
					}
					continue
				}

				g, errCtx := errgroup.WithContext(ctx)

				g.Go(func() error { return d.processScheduledBatch(errCtx, list.scheduled) })
				g.Go(func() error { return d.syncRunningBatch(errCtx, list.active) })
				g.Go(func() error { return d.deleteBatch(errCtx, list.scheduledDeletion) })

				err = g.Wait()
				if err != nil {
					logrus.Error("error running deploy watcher: ", err)
					continue
				}
			}
		}
	})
}

func (d *deployWatcher) Stop() error {
	d.stopOnce.Do(func() {
		d.ticker.Stop()
		close(d.done)
	})
	return nil
}

type deploymentsList struct {
	scheduled         []domain.Deployment
	active            []domain.Deployment
	scheduledDeletion []domain.Deployment
}

func (d *deployWatcher) listDeployments(ctx context.Context) (deploymentsList, error) {
	listReq := domain.ListDeploymentsReq{
		NodeIds: []int64{d.nodeId},
		NotStatus: []deployments_queries.VelezDeploymentStatus{
			deployments_queries.VelezDeploymentStatusDELETED,
			deployments_queries.VelezDeploymentStatusFAILED,
		},
	}
	deployments, err := d.deploymentsStorage.List(ctx, listReq)
	if err != nil {
		return deploymentsList{}, rerrors.Wrap(err, "error listing deployments")
	}

	var list deploymentsList

	for _, dep := range deployments {
		switch dep.Status {
		case deployments_queries.VelezDeploymentStatusSCHEDULEDDEPLOYMENT,
			deployments_queries.VelezDeploymentStatusSCHEDULEDUPGRADE:
			list.scheduled = append(list.scheduled, dep)
		case deployments_queries.VelezDeploymentStatusRUNNING:
			list.active = append(list.active, dep)
		case deployments_queries.VelezDeploymentStatusSCHEDULEDDELETION:
			list.scheduledDeletion = append(list.scheduledDeletion, dep)

		default:
			logrus.Errorf("Error during listing deployments in deploy watcher: unknown deployment status: %v", dep.Status)
		}
	}

	return list, nil
}

func (d *deployWatcher) processScheduledBatch(ctx context.Context, scheduled []domain.Deployment) error {
	for _, dep := range scheduled {
		//	Step 1 - define what's needs to  be done - upgrade or new deployment
		switch dep.Status {
		case deployments_queries.VelezDeploymentStatusSCHEDULEDDEPLOYMENT:
			spec, err := d.deploymentsStorage.GetSpecificationById(ctx, dep.SpecId)
			if err != nil {
				return rerrors.Wrap(err, "")
			}

			r := domain.LaunchSmerd{
				CreateSmerd_Request: &velez_api.CreateSmerd_Request{},
			}

			err = json.Unmarshal(spec.VervPayload.RawMessage, &r.CreateSmerd_Request)
			if err != nil {
				return rerrors.Wrap(err, "")
			}

			updateStatusParams := deployments_queries.UpdateDeploymentStatusParams{
				Status: deployments_queries.VelezDeploymentStatusRUNNING,
				ID:     dep.Id,
			}

			runner := d.pipeliner.LaunchSmerd(r)
			err = runner.Run(ctx)
			if err != nil {
				logrus.Error("error deploying smerd", rerrors.Wrap(err, ""))
				updateStatusParams.Status = deployments_queries.VelezDeploymentStatusFAILED
			}

			err = d.deploymentsStorage.UpdateDeploymentStatus(ctx, updateStatusParams)
			if err != nil {
				return rerrors.Wrap(err, "UpdateDeploymentStatus")
			}
		case deployments_queries.VelezDeploymentStatusSCHEDULEDDELETION:
		case deployments_queries.VelezDeploymentStatusSCHEDULEDUPGRADE:
			spec, err := d.deploymentsStorage.GetSpecificationById(ctx, dep.SpecId)
			if err != nil {
				return rerrors.Wrap(err, "GetSpecificationById")
			}

			smerdReq := &velez_api.CreateSmerd_Request{}
			err = json.Unmarshal(spec.VervPayload.RawMessage, smerdReq)
			if err != nil {
				return rerrors.Wrap(err, "")
			}

			updateStatusParams := deployments_queries.UpdateDeploymentStatusParams{
				Status: deployments_queries.VelezDeploymentStatusRUNNING,
				ID:     dep.Id,
			}

			upgradeRunner := d.pipeliner.UpgradeSmerd(domain.UpgradeSmerd{
				Name:  smerdReq.GetName(),
				Image: smerdReq.GetImageName(),
			})
			err = upgradeRunner.Run(ctx)
			if err != nil {
				logrus.Error("error upgrading smerd: ", rerrors.Wrap(err, ""))
				updateStatusParams.Status = deployments_queries.VelezDeploymentStatusFAILED
			}

			err = d.deploymentsStorage.UpdateDeploymentStatus(ctx, updateStatusParams)
			if err != nil {
				return rerrors.Wrap(err, "UpdateDeploymentStatus")
			}
		}
	}
	return nil
}

func (d *deployWatcher) syncRunningBatch(ctx context.Context, active []domain.Deployment) error {
	for _, dep := range active {
		spec, err := d.deploymentsStorage.GetSpecificationById(ctx, dep.SpecId)
		if err != nil {
			return rerrors.Wrap(err, "error getting spec for running deployment")
		}

		smerdReq := &velez_api.CreateSmerd_Request{}
		err = json.Unmarshal(spec.VervPayload.RawMessage, smerdReq)
		if err != nil {
			return rerrors.Wrap(err, "error unmarshaling spec")
		}

		running, _, err := d.nodeClients.Docker().IsContainerRunning(ctx, smerdReq.GetName())
		if err != nil {
			logrus.Errorf("error inspecting container %s: %v", smerdReq.GetName(), err)
			continue
		}

		if running {
			continue
		}

		err = d.deploymentsStorage.UpdateDeploymentStatus(ctx, deployments_queries.UpdateDeploymentStatusParams{
			Status: deployments_queries.VelezDeploymentStatusFAILED,
			ID:     dep.Id,
		})
		if err != nil {
			return rerrors.Wrap(err, "error marking deployment as failed")
		}
	}
	return nil
}

func (d *deployWatcher) deleteBatch(ctx context.Context, deletion []domain.Deployment) error {
	for _, dep := range deletion {
		spec, err := d.deploymentsStorage.GetSpecificationById(ctx, dep.SpecId)
		if err != nil {
			return rerrors.Wrap(err, "error getting spec for deletion")
		}

		smerdReq := &velez_api.CreateSmerd_Request{}
		err = json.Unmarshal(spec.VervPayload.RawMessage, smerdReq)
		if err != nil {
			return rerrors.Wrap(err, "error unmarshaling spec")
		}

		err = d.nodeClients.Docker().Remove(ctx, smerdReq.GetName())
		if err != nil {
			logrus.Errorf("error removing container %s: %v", smerdReq.GetName(), err)
			continue
		}

		err = d.deploymentsStorage.UpdateDeploymentStatus(ctx, deployments_queries.UpdateDeploymentStatusParams{
			Status: deployments_queries.VelezDeploymentStatusDELETED,
			ID:     dep.Id,
		})
		if err != nil {
			return rerrors.Wrap(err, "error marking deployment as deleted")
		}
	}
	return nil
}
