package postgres

import (
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

type tasksStorage struct {
	db sqldb.DB

	*tasks_queries.Queries
}

func newTasksStorage(db sqldb.DB) *tasksStorage {
	return &tasksStorage{
		db:      db,
		Queries: tasks_queries.New(db),
	}
}
