package workers

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.redsock.ru/rerrors"
	"golang.org/x/sync/errgroup"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/pipelines"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

type deployWatcher struct {
	services           service.Services
	pipeliner          pipelines.Pipeliner
	deploymentsStorage storage.DeploymentsStorage

	nodeId int64

	starter sync.Once
	ticker  *time.Ticker
}

func NewDeployWatcher(
	services service.Services,
	runner pipelines.Pipeliner,
	clusterClients cluster_clients.ClusterClients,

	interval time.Duration) Worker {
	return &deployWatcher{
		services:           services,
		pipeliner:          runner,
		deploymentsStorage: clusterClients.StateManager().Deployments(),

		nodeId: 1,

		starter: sync.Once{},
		ticker:  time.NewTicker(interval),
	}
}

func (d *deployWatcher) Start(ctx context.Context) {
	d.starter.Do(func() {
		for range d.ticker.C {
			list, err := d.listDeployments(ctx)
			if err != nil {
				logrus.Error("error listing deployments in deploy watcher: ", err)
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
	})
}

func (d *deployWatcher) Stop() error {
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
			list.active = append(list.scheduled, dep)
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
				return rerrors.Wrap(err, "")
			}
		case deployments_queries.VelezDeploymentStatusSCHEDULEDDELETION:
		case deployments_queries.VelezDeploymentStatusSCHEDULEDUPGRADE:

		}
	}
	return nil
}

func (d *deployWatcher) syncRunningBatch(ctx context.Context, active []domain.Deployment) error {
	return nil
}

func (d *deployWatcher) deleteBatch(ctx context.Context, deletion []domain.Deployment) error {
	return nil
}
