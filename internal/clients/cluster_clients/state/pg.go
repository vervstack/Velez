package state

import (
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres"
)

type pgState struct {
	storage.Storage
}

const PgName = "verv-cluster-state"

func NewPgStateManager(dsn string) (cluster_clients.ClusterStateManager, error) {
	conn, err := sqldb.New(dsn)
	if err != nil {
		return nil, rerrors.Wrap(err, "error connecting to database")
	}

	return &pgState{
		Storage: postgres.New(conn),
	}, nil
}
