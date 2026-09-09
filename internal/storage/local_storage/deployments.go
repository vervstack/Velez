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
//
// deployments also doubles as this backend's storage.Transactor (see
// Execute below) - its mu is the one lock shared between plain reads/writes
// and a composite write run through verv_services.executeDeployment, so the
// two can never interleave.
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
	ctx context.Context,
	arg deployments_queries.CreateSpecificationParams,
) (int64, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.createSpecificationLocked(ctx, arg)
}

func (d *deployments) GetSpecificationById(
	ctx context.Context,
	id int64,
) (deployments_queries.GetSpecificationByIdRow, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.getSpecificationByIdLocked(ctx, id)
}

func (d *deployments) CreateDeployment(
	ctx context.Context,
	arg deployments_queries.CreateDeploymentParams,
) (any, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.createDeploymentLocked(ctx, arg)
}

func (d *deployments) UpdateDeploymentStatus(
	ctx context.Context,
	arg deployments_queries.UpdateDeploymentStatusParams,
) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.updateDeploymentStatusLocked(ctx, arg)
}

// WithTx returns a Querier that operates directly on d's state without
// taking d.mu, since it is only ever reached from inside Execute below -
// which already holds d.mu for the call's whole duration. The concrete
// generated *deployments_queries.Queries type postgres.deploymentsStorage
// returns from WithTx has no in-memory equivalent (it needs a real *sql.DB),
// which is why storage.DeploymentsStorage.WithTx is declared against the
// deployments_queries.Querier interface instead.
func (d *deployments) WithTx(_ *sql.Tx) deployments_queries.Querier {
	return (*deploymentsLockedQuerier)(d)
}

// Execute implements storage.Transactor for this backend: it locks d.mu for
// the duration of fn, then calls fn(nil) - there is no real *sql.Tx to hand
// back. Locking here (rather than an independent mutex) is what makes a
// composite write run through verv_services.executeDeployment genuinely
// exclude List/ListDeployments - see the deployments doc comment. fn always
// reaches d's state through WithTx's deploymentsLockedQuerier, whose methods
// assume the lock is already held, so this never deadlocks against the
// locking Create*/Update* methods above.
func (d *deployments) Execute(fn func(tx *sql.Tx) error) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return fn(nil)
}

func (d *deployments) createSpecificationLocked(
	_ context.Context,
	arg deployments_queries.CreateSpecificationParams,
) (int64, error) {
	d.nextSpecId++

	d.specs[d.nextSpecId] = arg

	return d.nextSpecId, nil
}

func (d *deployments) getSpecificationByIdLocked(
	_ context.Context,
	id int64,
) (deployments_queries.GetSpecificationByIdRow, error) {
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

func (d *deployments) createDeploymentLocked(
	_ context.Context,
	arg deployments_queries.CreateDeploymentParams,
) (any, error) {
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

func (d *deployments) updateDeploymentStatusLocked(
	_ context.Context,
	arg deployments_queries.UpdateDeploymentStatusParams,
) error {
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

// deploymentsLockedQuerier adapts *deployments to deployments_queries.Querier
// for use from inside Execute, where d.mu is already held - it calls the
// *Locked variants directly instead of the public, self-locking methods
// (sync.Mutex isn't reentrant, so calling those would deadlock).
type deploymentsLockedQuerier deployments

func (q *deploymentsLockedQuerier) CreateDeployment(
	ctx context.Context,
	arg deployments_queries.CreateDeploymentParams,
) (any, error) {
	return (*deployments)(q).createDeploymentLocked(ctx, arg)
}

func (q *deploymentsLockedQuerier) CreateSpecification(
	ctx context.Context,
	arg deployments_queries.CreateSpecificationParams,
) (int64, error) {
	return (*deployments)(q).createSpecificationLocked(ctx, arg)
}

func (q *deploymentsLockedQuerier) GetSpecificationById(
	ctx context.Context,
	id int64,
) (deployments_queries.GetSpecificationByIdRow, error) {
	return (*deployments)(q).getSpecificationByIdLocked(ctx, id)
}

func (q *deploymentsLockedQuerier) UpdateDeploymentStatus(
	ctx context.Context,
	arg deployments_queries.UpdateDeploymentStatusParams,
) error {
	return (*deployments)(q).updateDeploymentStatusLocked(ctx, arg)
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
