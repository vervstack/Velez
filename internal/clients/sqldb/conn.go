package sqldb

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox/closer"
)

const dialect = "postgres"

func New(connectionString string) (*sql.DB, error) {
	conn, err := sql.Open(dialect, connectionString)
	if err != nil {
		return nil, rerrors.Wrap(err, "error checking connection to postgres")
	}

	closer.Add(func() error {
		return conn.Close()
	})

	return conn, nil
}

func RollMigration(rootDsn string) (err error) {
	conn, err := sql.Open(dialect, rootDsn)
	if err != nil {
		return rerrors.Wrap(err, "error checking connection to postgres")
	}

	defer func() {
		e := conn.Close()
		if e != nil {
			log.Error().Err(e).Msg("error closing connection to postgres")
		}
	}()

	goose.SetLogger(sqlLogger{})
	err = goose.SetDialect(dialect)
	if err != nil {
		return rerrors.Wrap(err, "error setting dialect")
	}

	err = goose.Up(conn, "./migrations")
	if err != nil {
		return rerrors.Wrap(err, "error performing up")
	}

	return nil
}

type DB interface {
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)

	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)

	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)

	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type sqlLogger struct{}

func (s sqlLogger) Fatalf(format string, v ...interface{}) {
	log.Fatal().Msgf(format, v...)
}

func (s sqlLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}
