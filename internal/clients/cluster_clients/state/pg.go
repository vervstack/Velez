package state

import (
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka/resources"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/sqldb"
)

type pgState struct {
	conn sqldb.DB
}

func NewPgStateManager(dsn string) (cluster_clients.ClusterStateManager, error) {
	res := resources.Postgres{}
	err := res.ParseFromDsn(dsn)
	if err != nil {
		return nil, rerrors.Wrap(err, "error parsing dsn")
	}

	conn, err := sqldb.New(&res)
	if err != nil {
		return nil, rerrors.Wrap(err, "error connecting to database")
	}

	return &pgState{
		conn,
	}, nil
}
