package verv_services

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/sqlc-dev/pqtype"
	"go.redsock.ru/rerrors"

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
	return nil
}
