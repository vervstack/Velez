package postgres

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/environments"
	service_resources_queries "go.vervstack.ru/Velez/internal/storage/postgres/generated/service_resources_queries"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/services_queries"
	"go.vervstack.ru/Velez/internal/storage/registries"
)

type Storage struct {
	nodeStorage                *nodeStorage
	servicesStorage            *servicesStorage
	deploymentsStorage         *deploymentsStorage
	pluginsStorage             *pluginsStorage
	serviceDependenciesStorage *serviceDependenciesStorage
	serviceResourcesStorage    *serviceResourcesStorage
	tasksStorage               *tasksStorage
	jobsStorage                *jobsStorage
	environmentsStorage        storage.EnvironmentsStorage
	registriesStorage          storage.RegistriesStorage

	txManager *sqldb.TxManager
}

func New(db *sql.DB) storage.Storage {
	return &Storage{
		nodeStorage: newNodeStorage(db),
		servicesStorage: &servicesStorage{
			conn:            db,
			querier:         services_queries.New(db),
			resourceQuerier: service_resources_queries.New(db),
		},

		deploymentsStorage:         newDeploymentsStorage(db),
		pluginsStorage:             newPluginsStorage(db),
		serviceDependenciesStorage: newServiceDependenciesStorage(db),
		serviceResourcesStorage:    newServiceResourcesStorage(db),
		tasksStorage:               newTasksStorage(db),
		jobsStorage:                newJobsStorage(db),
		environmentsStorage:        environments.NewPg(db),
		registriesStorage:          registries.NewPg(db),
		txManager:                  sqldb.NewTxManager(db),
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

func (s *Storage) Plugins() storage.PluginsStorage {
	return s.pluginsStorage
}

func (s *Storage) ServiceDependencies() storage.ServiceDependenciesStorage {
	return s.serviceDependenciesStorage
}

func (s *Storage) ServiceResources() storage.ServiceResourcesStorage {
	return s.serviceResourcesStorage
}

func (s *Storage) Tasks() storage.TasksStorage {
	return s.tasksStorage
}

func (s *Storage) Jobs() storage.JobsStorage {
	return s.jobsStorage
}

func (s *Storage) Environments() storage.EnvironmentsStorage {
	return s.environmentsStorage
}

func (s *Storage) Registries() storage.RegistriesStorage {
	return s.registriesStorage
}

func (s *Storage) TxManager() *sqldb.TxManager {
	return s.txManager
}

func wrapPgErr(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return rerrors.Wrap(storage.ErrNotFound)
	}

	var pgErr *pq.Error

	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" { // unique_violation
			return errors.Join(storage.ErrAlreadyExists, err)
		}
	}

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
