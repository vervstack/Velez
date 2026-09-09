package local_storage

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
)

// deployments is a real in-memory Deployments store, backing single-node/dev
// mode the way postgres/deployments.go backs cluster mode. There is no
// persisted services table in this backend (see dockerServices), so a spec's
// ServiceID is always whatever dockerServices.GetByName handed back, which is
// the zero value for every container-derived service - ListDeploymentsReq's
// ServiceName filter can't be resolved to that and is left unapplied here,
// same as before this store existed.
type deployments struct {
	mu sync.Mutex

	nextSpecId       int64
	nextDeploymentId int64

	specs       map[int64]deployments_queries.CreateSpecificationParams
	deployments []domain.Deployment
}

func newDeploymentsStorage() *deployments {
	return &deployments{
		specs: make(map[int64]deployments_queries.CreateSpecificationParams),
	}
}

func (d *deployments) List(_ context.Context, req domain.ListDeploymentsReq) ([]domain.Deployment, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	filtered := d.filterLocked(req)

	return paginate(filtered, req.Paging), nil
}

func (d *deployments) ListDeployments(
	_ context.Context,
	req domain.ListDeploymentsReq,
) (domain.DeploymentList, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	filtered := d.filterLocked(req)

	list := domain.DeploymentList{
		Deployments: paginate(filtered, req.Paging),
		Total:       uint64(len(filtered)),
	}

	return list, nil
}

func (d *deployments) CreateSpecification(
	_ context.Context,
	arg deployments_queries.CreateSpecificationParams,
) (int64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.nextSpecId++

	d.specs[d.nextSpecId] = arg

	return d.nextSpecId, nil
}

func (d *deployments) GetSpecificationById(
	_ context.Context,
	id int64,
) (deployments_queries.GetSpecificationByIdRow, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	spec, ok := d.specs[id]
	if !ok {
		return deployments_queries.GetSpecificationByIdRow{}, storage.ErrNotFound
	}

	row := deployments_queries.GetSpecificationByIdRow{
		ID:             id,
		Name:           spec.Name,
		VervPayload:    spec.VervPayload,
		VervDescriptor: spec.VervDescriptor,
	}

	return row, nil
}

func (d *deployments) CreateDeployment(
	_ context.Context,
	arg deployments_queries.CreateDeploymentParams,
) (any, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.nextDeploymentId++

	now := time.Now()

	dep := domain.Deployment{
		Id:        d.nextDeploymentId,
		SpecId:    arg.SpecID,
		NodeId:    int64(arg.NodeID),
		Status:    arg.Status,
		CreatedAt: now,
		UpdatedAt: now,
	}

	spec, ok := d.specs[arg.SpecID]
	if ok && spec.ServiceID.Valid {
		dep.ServiceId = spec.ServiceID.Int64
	}

	d.deployments = append(d.deployments, dep)

	return d.nextDeploymentId, nil
}

func (d *deployments) UpdateDeploymentStatus(
	_ context.Context,
	arg deployments_queries.UpdateDeploymentStatusParams,
) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for i := range d.deployments {
		if d.deployments[i].Id != arg.ID {
			continue
		}

		d.deployments[i].Status = arg.Status
		d.deployments[i].UpdatedAt = time.Now()

		return nil
	}

	return storage.ErrNotFound
}

// WithTx has no meaningful local-storage equivalent: verv_services.deploy.go
// only calls it once v.dataStorage.TxManager() has already returned non-nil,
// which never happens for this backend (see TxManager below), so this is
// unreachable in practice rather than silently wrong.
func (d *deployments) WithTx(_ *sql.Tx) *deployments_queries.Queries {
	return nil
}

func (d *deployments) filterLocked(req domain.ListDeploymentsReq) []domain.Deployment {
	notStatus := make(map[deployments_queries.VelezDeploymentStatus]struct{}, len(req.NotStatus))
	for _, s := range req.NotStatus {
		notStatus[s] = struct{}{}
	}

	nodeIds := make(map[int64]struct{}, len(req.NodeIds))
	for _, id := range req.NodeIds {
		nodeIds[id] = struct{}{}
	}

	out := make([]domain.Deployment, 0, len(d.deployments))

	for _, dep := range d.deployments {
		_, excluded := notStatus[dep.Status]
		if excluded {
			continue
		}

		if len(nodeIds) == 0 {
			out = append(out, dep)

			continue
		}

		_, ok := nodeIds[dep.NodeId]
		if ok {
			out = append(out, dep)
		}
	}

	return out
}

// paginate applies offset/limit after filtering and after Total has already
// been counted, mirroring postgres/deployments.go's ListDeployments (a
// separate countTotal query runs before the LIMIT/OFFSET clause).
func paginate(deployments []domain.Deployment, paging domain.Paging) []domain.Deployment {
	if paging.Offset < uint64(len(deployments)) {
		deployments = deployments[paging.Offset:]
	} else {
		deployments = nil
	}

	if paging.Limit > 0 && uint64(len(deployments)) > paging.Limit {
		deployments = deployments[:paging.Limit]
	}

	return deployments
}
