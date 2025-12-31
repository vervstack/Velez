package postgres

import (
	pg_queries "go.vervstack.ru/Velez/internal/storage/postgres/generated/services_queries"
)

type servicesStorage struct {
	pg_queries.Querier
}
