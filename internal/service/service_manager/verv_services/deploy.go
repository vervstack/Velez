package verv_services

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
)

//nolint:forbidigo // package-private sentinel, not shared/user-facing
var errNoRunningDeployment = rerrors.New("no running deployment found for service")

// CreateNewDeploy upserts request.ServiceName before looking it up so a
// service with no container yet (nothing else has created it) still
// resolves - both enable_registry.go's deployRegistryJob and
// pgaas.CreatePgInstance rely on this instead of upserting themselves.
func (v *VervService) CreateNewDeploy(ctx context.Context, request domain.CreateDeployReq) error {
	err := v.dataStorage.Services().UpsertService(ctx, request.ServiceName)
	if err != nil {
		return rerrors.Wrap(err, "error upserting service")
	}

	svc, err := v.dataStorage.Services().GetByName(ctx, request.ServiceName)
	if err != nil {
		return rerrors.Wrap(err, "error getting service")
	}

	spec := deployments_queries.CreateSpecificationParams{
		Name:      uuid.New().String(),
		ServiceID: sql.NullInt64{Int64: svc.ID, Valid: true},
		VervPayload: pqtype.NullRawMessage{
			RawMessage: nil,
			Valid:      true,
		},
	}

	spec.VervPayload.RawMessage, err = json.Marshal(request.LaunchSmerd)
	if err != nil {
		return rerrors.Wrap(err, "error marshaling specification")
	}

	if request.VervDescriptor != nil {
		// internal/domain/vervonomicon.Descriptor carries only yaml
		// tags (it's a DO-NOT-EDIT pure data package - see
		// docs/features/vervonomicon.md) - encoding/json falls back
		// to Go field names, which is fine here since this JSON is
		// never read by anything outside this codebase.
		spec.VervDescriptor.RawMessage, err = json.Marshal(request.VervDescriptor) //nolint:musttag
		if err != nil {
			return rerrors.Wrap(err, "error marshaling vervonomicon descriptor")
		}

		spec.VervDescriptor.Valid = true
	}

	deploymentFn := func(deploymentStorage deployments_queries.Querier) error {
		specId, specErr := deploymentStorage.CreateSpecification(ctx, spec)
		if specErr != nil {
			return rerrors.Wrap(specErr, "error creating specification")
		}

		deployment := deployments_queries.CreateDeploymentParams{
			NodeID: 1,
			Status: deployments_queries.VelezDeploymentStatusSCHEDULEDDEPLOYMENT,
			SpecID: specId,
		}

		_, specErr = deploymentStorage.CreateDeployment(ctx, deployment)
		if specErr != nil {
			return rerrors.Wrap(specErr, "error creating new deploy")
		}

		return nil
	}

	err = v.executeDeployment(ctx, deploymentFn)
	if err != nil {
		return rerrors.Wrap(err, "error creating new deploy")
	}

	return nil
}

// executeDeployment runs fn against the Deployments store, wrapped in one
// atomic unit of work via storage.Transactor - a real SQL transaction for
// postgres, a mutex critical section for local_storage (see
// local_storage.deployments.Execute). Both backends always return a
// non-nil Transactor, so this never needs to special-case one.
func (v *VervService) executeDeployment(ctx context.Context, fn func(deployments_queries.Querier) error) error {
	err := v.dataStorage.TxManager().Execute(func(tx *sql.Tx) error {
		return fn(v.dataStorage.Deployments().WithTx(tx))
	})
	if err != nil {
		return rerrors.Wrap(err, "error executing deployment transaction")
	}

	return nil
}

func (v *VervService) ListDeployments(
	ctx context.Context,
	req domain.ListDeploymentsReq,
) (domain.DeploymentList, error) {
	list, err := v.dataStorage.Deployments().ListDeployments(ctx, req)
	if err != nil {
		return domain.DeploymentList{}, rerrors.Wrap(err, "error listing deployments")
	}

	return list, nil
}

func (v *VervService) UpgradeDeploy(ctx context.Context, request domain.UpgradeDeployReq) error {
	listReq := domain.ListDeploymentsReq{
		ServiceName: request.ServiceName,
	}

	deployments, err := v.dataStorage.Deployments().List(ctx, listReq)
	if err != nil {
		return rerrors.Wrap(err, "error listing deployments")
	}

	var runningDep *domain.Deployment

	for i := range deployments {
		if deployments[i].Status == deployments_queries.VelezDeploymentStatusRUNNING {
			runningDep = &deployments[i]

			break
		}
	}

	if runningDep == nil {
		return errNoRunningDeployment
	}

	currentSpec, err := v.dataStorage.Deployments().GetSpecificationById(ctx, runningDep.SpecId)
	if err != nil {
		return rerrors.Wrap(err, "error getting current spec")
	}

	smerdReq := &velez_api.CreateSmerd_Request{}

	err = json.Unmarshal(currentSpec.VervPayload.RawMessage, smerdReq)
	if err != nil {
		return rerrors.Wrap(err, "error unmarshaling spec payload")
	}

	if request.NewImage != nil {
		smerdReq.ImageName = *request.NewImage
	}

	payload, err := json.Marshal(smerdReq)
	if err != nil {
		return rerrors.Wrap(err, "error marshaling updated spec")
	}

	svc, err := v.dataStorage.Services().GetByName(ctx, request.ServiceName)
	if err != nil {
		return rerrors.Wrap(err, "error getting service")
	}

	upgradeFn := func(q deployments_queries.Querier) error {
		createSpecsParams := deployments_queries.CreateSpecificationParams{
			Name:      uuid.New().String(),
			ServiceID: sql.NullInt64{Int64: svc.ID, Valid: true},
			VervPayload: pqtype.NullRawMessage{
				RawMessage: payload,
				Valid:      true,
			},
		}

		newSpecId, specErr := q.CreateSpecification(ctx, createSpecsParams)
		if specErr != nil {
			return rerrors.Wrap(specErr, "error creating updated spec")
		}

		createDeploymentParams := deployments_queries.CreateDeploymentParams{
			NodeID: int32(runningDep.NodeId),
			Status: deployments_queries.VelezDeploymentStatusSCHEDULEDUPGRADE,
			SpecID: newSpecId,
		}

		_, specErr = q.CreateDeployment(ctx, createDeploymentParams)
		if specErr != nil {
			return rerrors.Wrap(specErr, "error creating upgrade deployment")
		}

		return nil
	}

	err = v.executeDeployment(ctx, upgradeFn)
	if err != nil {
		return rerrors.Wrap(err, "error scheduling upgrade")
	}

	return nil
}
