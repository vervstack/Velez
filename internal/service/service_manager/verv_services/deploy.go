package verv_services

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/sqlc-dev/pqtype"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
)

func (v *VervService) CreateNewDeploy(ctx context.Context, request domain.CreateDeployReq) error {
	err := v.txManager.Execute(
		func(tx *sql.Tx) (err error) {
			q := v.deploymentsStorage.WithTx(tx)

			spec := deployments_queries.CreateSpecificationParams{
				Name: "service_id=" + strconv.FormatUint(request.ServiceId, 10),
				VervPayload: pqtype.NullRawMessage{
					RawMessage: nil,
					Valid:      true,
				},
			}
			spec.VervPayload.RawMessage, err = json.Marshal(request.LaunchSmerd)
			if err != nil {
				return rerrors.Wrap(err, "error marshaling specification")
			}

			specId, err := q.CreateSpecification(ctx, spec)
			if err != nil {
				return rerrors.Wrap(err, "error creating specification")
			}

			deployment := deployments_queries.CreateDeploymentParams{
				ServiceID: int64(request.ServiceId),
				NodeID:    1,
				Status:    deployments_queries.VelezDeploymentStatusSCHEDULEDDEPLOYMENT,
				SpecID:    specId,
			}

			_, err = q.CreateDeployment(ctx, deployment)
			if err != nil {
				return rerrors.Wrap(err, "error creating new deploy")
			}

			return nil
		})
	if err != nil {
		return rerrors.Wrap(err, "error creating new deploy")
	}

	return nil
}

func (v *VervService) UpgradeDeploy(ctx context.Context, request domain.UpgradeDeployReq) error {
	deployments, err := v.deploymentsStorage.List(ctx, domain.ListDeploymentsReq{
		ServiceIds: []int64{int64(request.ServiceId)},
	})
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
		return rerrors.New("no running deployment found for service")
	}

	currentSpec, err := v.deploymentsStorage.GetSpecificationById(ctx, runningDep.SpecId)
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

	err = v.txManager.Execute(func(tx *sql.Tx) error {
		q := v.deploymentsStorage.WithTx(tx)

		newSpecId, err := q.CreateSpecification(ctx, deployments_queries.CreateSpecificationParams{
			Name: currentSpec.Name,
			VervPayload: pqtype.NullRawMessage{
				RawMessage: payload,
				Valid:      true,
			},
		})
		if err != nil {
			return rerrors.Wrap(err, "error creating updated spec")
		}

		_, err = q.CreateDeployment(ctx, deployments_queries.CreateDeploymentParams{
			ServiceID: int64(request.ServiceId),
			NodeID:    int32(runningDep.NodeId),
			Status:    deployments_queries.VelezDeploymentStatusSCHEDULEDUPGRADE,
			SpecID:    newSpecId,
		})
		if err != nil {
			return rerrors.Wrap(err, "error creating upgrade deployment")
		}

		return nil
	})
	if err != nil {
		return rerrors.Wrap(err, "error scheduling upgrade")
	}

	return nil
}
