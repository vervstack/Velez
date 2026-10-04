package local_storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"maps"
	"sync"
	"time"

	"github.com/sqlc-dev/pqtype"
	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/parser"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

// deployments is a real in-memory Deployments store, backing single-node/dev
// mode the way postgres/deployments.go backs cluster mode. There is no
// persisted services table in this backend (see dockerServices), so a spec's
// ServiceID is always whatever dockerServices.GetByName handed back, which is
// the zero value for every container-derived service - ListDeploymentsReq's
// ServiceName filter is therefore matched against the service name embedded
// in the deployment's spec payload instead.
//
// deployments also doubles as this backend's storage.Transactor (see
// Execute below) - its mu is the one lock shared between plain reads/writes
// and a composite write run through verv_services.executeDeployment, so the
// two can never interleave.
type deployments struct {
	docker node_clients.Docker

	mu sync.Mutex

	nextSpecId       int64
	nextDeploymentId int64

	specs       map[int64]deployments_queries.CreateSpecificationParams
	deployments []domain.Deployment
}

func newDeploymentsStorage(docker node_clients.Docker) *deployments {
	return &deployments{
		docker: docker,
		specs:  make(map[int64]deployments_queries.CreateSpecificationParams),
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

// getSpecificationByIdLocked prefers deriving the spec straight from the
// service's live container over the specs map entry: the map only records
// what was requested at deploy/upgrade time, which is what UpgradeDeploy
// would otherwise diff its new spec against - the container is what
// actually decides upgrade correctness, and it also has no relation to
// what predates a Velez restart wiping this map. The map entry only
// resolves the pre-container bridge window (name resolves but no
// container exists yet).
func (d *deployments) getSpecificationByIdLocked(
	ctx context.Context,
	id int64,
) (deployments_queries.GetSpecificationByIdRow, error) {
	spec, ok := d.specs[id]
	if !ok {
		return deployments_queries.GetSpecificationByIdRow{}, user_errors.ErrStorageNotFound
	}

	name := specServiceName(spec)
	if name != "" {
		row, found, err := d.resolveSpecFromContainer(ctx, name)
		if err != nil {
			return deployments_queries.GetSpecificationByIdRow{}, err
		}

		if found {
			return overlayRequestedSpec(row, spec)
		}
	}

	row := deployments_queries.GetSpecificationByIdRow{
		ID:             id,
		Name:           spec.Name,
		VervPayload:    spec.VervPayload,
		VervDescriptor: spec.VervDescriptor,
	}

	return row, nil
}

// overlayRequestedSpec merges the env and extra networks the stored spec
// requested over the container-derived row: the container is what is running,
// but a network without aliases never shows up in a request derived from it,
// and an upgrade overlay (UpgradeDeployReq.EnvOverrides/ExtraNetworks) lives
// only in the stored spec.
func overlayRequestedSpec(
	derived deployments_queries.GetSpecificationByIdRow,
	requested deployments_queries.CreateSpecificationParams,
) (deployments_queries.GetSpecificationByIdRow, error) {
	if !requested.VervPayload.Valid {
		return derived, nil
	}

	requestedReq := &pb.CreateSmerd_Request{}

	err := json.Unmarshal(requested.VervPayload.RawMessage, requestedReq)
	if err != nil {
		return derived, rerrors.Wrap(err, "error unmarshaling stored spec")
	}

	derivedReq := &pb.CreateSmerd_Request{}

	err = json.Unmarshal(derived.VervPayload.RawMessage, derivedReq)
	if err != nil {
		return derived, rerrors.Wrap(err, "error unmarshaling container spec")
	}

	mergedEnv := make(map[string]string, len(derivedReq.GetEnv())+len(requestedReq.GetEnv()))
	maps.Copy(mergedEnv, derivedReq.GetEnv())
	maps.Copy(mergedEnv, requestedReq.GetEnv())

	derivedReq.Env = mergedEnv

	if derivedReq.GetSettings() == nil {
		derivedReq.Settings = &pb.Container_Settings{}
	}

	derivedReq.Settings.Network = domain.MergeNetworkBinds(
		derivedReq.GetSettings().GetNetwork(), requestedReq.GetSettings().GetNetwork(),
	)

	payload, err := json.Marshal(derivedReq)
	if err != nil {
		return derived, rerrors.Wrap(err, "error marshaling merged spec")
	}

	derived.VervPayload = pqtype.NullRawMessage{RawMessage: payload, Valid: true}

	return derived, nil
}

// specServiceName extracts the service/container name embedded in a spec's
// VervPayload - CreateNewDeploy and UpgradeDeploy always set the marshaled
// CreateSmerd_Request's Name to the target service. "" means the payload is
// absent or didn't decode, so the caller has no name to resolve a container
// by and falls back to the map entry as-is.
func specServiceName(spec deployments_queries.CreateSpecificationParams) string {
	if !spec.VervPayload.Valid {
		return ""
	}

	var partial struct {
		Name string `json:"name"`
	}

	err := json.Unmarshal(spec.VervPayload.RawMessage, &partial)
	if err != nil {
		return ""
	}

	return partial.Name
}

// resolveSpecFromContainer finds name's live container and derives the spec
// currently running from it directly via ContainerInspect, the same
// container-is-system-of-record idiom as dockerServices.GetByName and
// dockerPgInstances.listFromContainers. found is false when there's no live
// container for name yet.
func (d *deployments) resolveSpecFromContainer(
	ctx context.Context, name string,
) (deployments_queries.GetSpecificationByIdRow, bool, error) {
	listReq := &pb.ListSmerds_Request{Name: &name}

	containers, err := d.docker.ListContainers(ctx, listReq, allEnvironments)
	if err != nil {
		return deployments_queries.GetSpecificationByIdRow{}, false, rerrors.Wrap(err, "error listing containers")
	}

	if len(containers) == 0 {
		return deployments_queries.GetSpecificationByIdRow{}, false, nil
	}

	info, err := d.docker.Client().ContainerInspect(ctx, containers[0].ID)
	if err != nil {
		return deployments_queries.GetSpecificationByIdRow{}, false, rerrors.Wrap(err, "error inspecting container")
	}

	smerdReq := parser.ToCreateRequest(name, info)

	payload, err := json.Marshal(smerdReq)
	if err != nil {
		return deployments_queries.GetSpecificationByIdRow{}, false, rerrors.Wrap(err, "error marshaling container spec")
	}

	row := deployments_queries.GetSpecificationByIdRow{
		Name:        name,
		CreatedAt:   time.Unix(containers[0].Created, 0),
		VervPayload: pqtype.NullRawMessage{RawMessage: payload, Valid: true},
	}

	return row, true, nil
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

	return user_errors.ErrStorageNotFound
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

		isOtherService := req.ServiceName != "" && specServiceName(d.specs[dep.SpecId]) != req.ServiceName
		if isOtherService {
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
