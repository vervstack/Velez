package postgres

import (
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/jobs_queries"
)

type jobsStorage struct {
	db sqldb.DB

	*jobs_queries.Queries
}

func newJobsStorage(db sqldb.DB) *jobsStorage {
	return &jobsStorage{
		db:      db,
		Queries: jobs_queries.New(db),
	}
}
