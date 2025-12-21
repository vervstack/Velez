package pipelines

import (
	"context"
	"time"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka/resources"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients/state"
	"go.vervstack.ru/Velez/internal/cluster/cluster_state"
	"go.vervstack.ru/Velez/internal/patterns/db_patterns/pg_pattern"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/cluster_steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/container_steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/smerd_steps"
)

func (p *pipeliner) EnableStatefullMode() Runner[any] {
	//region Pipeline Context

	const schema = "velez"
	const nodeName = "icy_raccoon"

	containerName := cluster_state.Name

	launchContainer := pg_pattern.Postgres(
		pg_pattern.WithInstanceName(containerName),
		pg_pattern.WithExposedPort(), // TODO make configurable via request
		pg_pattern.WithPort(25432),
		pg_pattern.WithPassword("d3hSejFkZnF0"), // TODO remove after debug
	)

	var containerId string
	var rootDsn string
	userPwd := "d3hSejFkZnF0" //string(toolbox.RandomBase64(12))
	//var nodeDsn string

	//endregion

	return &runner[any]{
		Steps: []steps.Step{
			container_steps.Create(
				p.nodeClients, &launchContainer.Pattern,
				&containerName, &containerId),
			smerd_steps.Start(p.nodeClients, &containerId),
			cluster_steps.GetRgRootDsn(p.nodeClients.Docker(), &containerId, &rootDsn),
			steps.SingleFunc(func(ctx context.Context) error {
				// TODO Wait for healthy
				time.Sleep(3 * time.Second)
				pgClusterState, err := state.NewPgStateManager(rootDsn)
				if err != nil {
					return rerrors.Wrap(err, "error initializing pgClusterState")
				}

				p.clusterClients.StateManager().Set(pgClusterState)
				return nil
			}),
			cluster_steps.CreatePgUserForNode(
				&rootDsn, schema, nodeName, userPwd),

			steps.SingleFunc(func(ctx context.Context) error {
				localStateManager := p.nodeClients.LocalStateManager()

				localState := localStateManager.GetForUpdate()
				localState.PgRootDsn = rootDsn

				var nodeConnection resources.Postgres

				err := nodeConnection.ParseFromDsn(rootDsn)
				if err != nil {
					return rerrors.Wrap(err, "error parsing root user database connection")
				}

				nodeConnection.User = nodeName
				nodeConnection.Pwd = userPwd
				localState.PgNodeDsn = nodeConnection.ConnectionString()

				localStateManager.SetAndRelease(localState)
				return nil
			}),

			// Enable integration with state
		},
	}
}
