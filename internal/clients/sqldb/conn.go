package sqldb

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox/closer"
)

const (
	Dialect = "postgres"
)

func New(connectionString string) (*sql.DB, error) {
	conn, err := sql.Open(Dialect, connectionString)
	if err != nil {
		return nil, rerrors.Wrap(err, "error checking connection to postgres")
	}

	closer.Add(func() error {
		return conn.Close()
	})

	return conn, nil
}

// RollMigration applies the goose migrations under migrationsDir. Pass ""
// for the production default ("./migrations", relative to the process's
// working directory) - callers only need a non-default value in e2e tests,
// which run from a directory other than the repo root.
func RollMigration(rootDsn string, migrationsDir string) (err error) {
	conn, err := sql.Open(Dialect, rootDsn)
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

	err = goose.SetDialect(Dialect)
	if err != nil {
		return rerrors.Wrap(err, "error setting dialect")
	}

	goose.SetTableName("velez.__migrations")

	if migrationsDir == "" {
		migrationsDir = "./migrations"
	}

	err = goose.Up(conn, migrationsDir)
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

func (s sqlLogger) Fatalf(format string, v ...any) {
	log.Fatal().Msgf(format, v...)
}

func (s sqlLogger) Printf(format string, v ...any) {
	log.Printf(format, v...)
}

type Scannable interface {
	Scan(dest ...any) error
}
