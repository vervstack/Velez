package postgres

import (
	"database/sql"

	"github.com/rs/zerolog/log"

	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/nodes_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/services_queries"
)

type Storage struct {
	nodeStorage        *nodeStorage
	servicesStorage    *servicesStorage
	deploymentsStorage *deploymentsStorage

	txManager *sqldb.TxManager
}

func New(db *sql.DB) storage.Storage {
	return &Storage{
		nodeStorage: &nodeStorage{
			querier: nodes_queries.New(db),
		},
		servicesStorage: &servicesStorage{
			conn:    db,
			Querier: services_queries.New(db),
		},

		deploymentsStorage: newDeploymentsStorage(db),
		txManager:          sqldb.NewTxManager(db),
	}
}

func (s *Storage) Nodes() storage.NodesStorage {
	return s.nodeStorage
}

func (s *Storage) Services() storage.ServicesStorage {
	return s.servicesStorage
}

func (s *Storage) Deployments() storage.DeploymentsStorage {
	return s.deploymentsStorage
}

func (s *Storage) TxManager() *sqldb.TxManager {
	return s.txManager
}

func wrapPgErr(err error) error {
	return err
}

func closeRows(rows *sql.Rows) {
	err := rows.Close()
	if err != nil {
		log.Error().
			Err(err).
			Msg("error closing rows")
	}
}
