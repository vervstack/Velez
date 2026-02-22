package pipelines

import (
	"context"
	"time"

	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"
	"go.vervstack.ru/matreshka/pkg/matreshka/resources"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients/state"
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/patterns/db_patterns/pg_pattern"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/cluster_steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/container_steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/smerd_steps"
)

func (p *pipeliner) EnableStatefullMode(req domain.EnableStatefullClusterRequest) Runner[domain.StateClusterDefinition] {
	//region Pipeline Context

	const schema = "velez"
	const masterNodeDefaultName = "icy_raccoon"

	containerName := state.PgName

	var ops []pg_pattern.Opt
	ops = append(ops, pg_pattern.WithInstanceName(containerName))

	if req.ExposePort {
		if req.ExposeToPort != 0 {
			ops = append(ops, pg_pattern.WithPort(req.ExposeToPort))
		} else {
			ops = append(ops, pg_pattern.WithExposedPort())
		}
	}

	var containerId string

	var rootDsn string

	localState := p.nodeClients.LocalStateManager().Get()

	var rootPwd string
	rootPg := resources.Postgres{}
	if rootPg.ParseFromDsn(localState.ClusterState.PgRootDsn) == nil {
		rootPwd = rootPg.Pwd
	} else {
		rootPwd = string(toolbox.RandomBase64(12))
	}

	ops = append(ops, pg_pattern.WithPassword(rootPwd))

	launchContainer := pg_pattern.Postgres(ops...)

	var userPwd string
	userPg := resources.Postgres{}
	if userPg.ParseFromDsn(localState.ClusterState.PgNodeDsn) == nil {
		userPwd = userPg.Pwd
	} else {
		userPwd = string(toolbox.RandomBase64(12))
	}
	//endregion

	response := domain.StateClusterDefinition{
		CreateReq:    launchContainer.Pattern,
		RootPostgres: rootPg,
	}

	return &runner[domain.StateClusterDefinition]{
		Steps: []steps.Step{
			// TODO Conflicts decisions must be handled on client side.
			// There is already existing state? Allow user to choose to delete\start it
			container_steps.Create(p.nodeClients, &launchContainer.Pattern, &containerName, &containerId),
			smerd_steps.Start(p.nodeClients, &containerId),
			cluster_steps.GetRgRootDsn(p.nodeClients.Docker(), &containerId, req.ExposePort, &rootDsn),
			steps.SingleFunc(func(ctx context.Context) error {
				// TODO Wait for healthy
				time.Sleep(3 * time.Second)
				err := sqldb.RollMigration(rootDsn)
				if err != nil {
					return rerrors.Wrap(err, "error rolling migration")
				}

				return nil
			}),
			cluster_steps.CreatePgUserForNode(
				&rootDsn, schema, masterNodeDefaultName, userPwd),

			steps.SingleFunc(func(ctx context.Context) error {
				localStateManager := p.nodeClients.LocalStateManager()

				localState := localStateManager.GetForUpdate()

				localState.ClusterState.PgRootDsn = rootDsn

				var nodeConnection resources.Postgres

				err := nodeConnection.ParseFromDsn(rootDsn)
				if err != nil {
					return rerrors.Wrap(err, "error parsing root user database connection")
				}

				nodeConnection.User = masterNodeDefaultName
				nodeConnection.Pwd = userPwd

				localState.ClusterState.PgNodeDsn = nodeConnection.ConnectionString() + "&application_name=" + masterNodeDefaultName

				pgClusterState, err := state.NewPgStateManager(ctx, localState.ClusterState.PgNodeDsn)
				if err != nil {
					return rerrors.Wrap(err, "error initializing pgClusterState")
				}

				p.clusterClients.StateManager().Set(pgClusterState)

				localStateManager.SetAndRelease(localState)
				return nil
			}),

			steps.SingleFunc(func(ctx context.Context) error {
				nodeStorage := p.clusterClients.StateManager().Nodes()
				err := nodeStorage.InitNode(ctx)
				if err != nil {
					return rerrors.Wrap(err, "error initializing node storage")
				}

				return nil
			}),
			// Enable integration with state
		},

		getResult: func() (res *domain.StateClusterDefinition, err error) {
			return &response, nil
		},
	}
}
