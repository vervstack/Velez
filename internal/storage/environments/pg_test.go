package environments

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
)

// environmentColumns mirrors velez.environments' column order, which every
// generated query in environments_queries selects with `SELECT *`.
func environmentColumns() []string {
	return []string{"id", "name", "suffix", "created_at", "updated_at"}
}

func newPgTest(t *testing.T) (*sql.DB, sqlmock.Sqlmock, storage.EnvironmentsStorage) {
	t.Helper()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db, mock, NewPg(db)
}

func TestPgStorage_ListEnvironments(t *testing.T) {
	_, mock, s := newPgTest(t)

	now := time.Now()

	rows := sqlmock.NewRows(environmentColumns()).
		AddRow(int64(1), "PROD", "", now, now).
		AddRow(int64(2), "STAGE", "stage", now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, suffix, created_at, updated_at")).
		WillReturnRows(rows)

	got, err := s.ListEnvironments(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, int64(1), got[0].ID)
	require.Equal(t, "PROD", got[0].Name)
	require.Equal(t, "stage", got[1].Suffix)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPgStorage_GetEnvironmentByName(t *testing.T) {
	_, mock, s := newPgTest(t)

	now := time.Now()

	rows := sqlmock.NewRows(environmentColumns()).
		AddRow(int64(7), "STAGE", "stage", now, now)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE name = $1")).
		WithArgs("STAGE").
		WillReturnRows(rows)

	got, err := s.GetEnvironmentByName(context.Background(), "STAGE")
	require.NoError(t, err)
	require.Equal(t, int64(7), got.ID)
	require.Equal(t, "stage", got.Suffix)
	require.NoError(t, mock.ExpectationsWereMet())
}

// A missing row must surface as storage.ErrNotFound, not a raw sql.ErrNoRows -
// the suffix-resolution path relies on that to reject unknown environments.
func TestPgStorage_GetEnvironmentByName_NotFound(t *testing.T) {
	_, mock, s := newPgTest(t)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE name = $1")).
		WithArgs("ghost").
		WillReturnError(sql.ErrNoRows)

	_, err := s.GetEnvironmentByName(context.Background(), "ghost")
	require.Error(t, err)
	require.True(t, errors.Is(err, storage.ErrNotFound))
}

func TestPgStorage_GetEnvironmentByID(t *testing.T) {
	_, mock, s := newPgTest(t)

	now := time.Now()

	rows := sqlmock.NewRows(environmentColumns()).
		AddRow(int64(3), "DEV", "dev", now, now)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE id = $1")).
		WithArgs(int64(3)).
		WillReturnRows(rows)

	got, err := s.GetEnvironmentByID(context.Background(), 3)
	require.NoError(t, err)
	require.Equal(t, "DEV", got.Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPgStorage_CreateEnvironment(t *testing.T) {
	_, mock, s := newPgTest(t)

	now := time.Now()

	rows := sqlmock.NewRows(environmentColumns()).
		AddRow(int64(9), "QA", "qa", now, now)

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO velez.environments")).
		WithArgs("QA", "qa").
		WillReturnRows(rows)

	req := domain.CreateEnvironmentReq{Name: "QA", Suffix: "qa"}

	got, err := s.CreateEnvironment(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, int64(9), got.ID)
	require.Equal(t, "qa", got.Suffix)
	require.NoError(t, mock.ExpectationsWereMet())
}

// UpdateEnvironment's params are optional: whatever the caller omits must be
// carried over from the current row rather than blanked out.
func TestPgStorage_UpdateEnvironment_OmittedFieldsKeepCurrentValues(t *testing.T) {
	_, mock, s := newPgTest(t)

	now := time.Now()

	current := sqlmock.NewRows(environmentColumns()).
		AddRow(int64(4), "STAGE", "stage", now, now)

	mock.ExpectQuery(regexp.QuoteMeta("WHERE id = $1")).
		WithArgs(int64(4)).
		WillReturnRows(current)

	updated := sqlmock.NewRows(environmentColumns()).
		AddRow(int64(4), "STAGE", "stg", now, now)

	mock.ExpectQuery(regexp.QuoteMeta("UPDATE velez.environments")).
		WithArgs(int64(4), "STAGE", "stg").
		WillReturnRows(updated)

	newSuffix := "stg"
	req := domain.UpdateEnvironmentReq{ID: 4, Suffix: &newSuffix}

	got, err := s.UpdateEnvironment(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "STAGE", got.Name)
	require.Equal(t, "stg", got.Suffix)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPgStorage_DeleteEnvironment(t *testing.T) {
	_, mock, s := newPgTest(t)

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM velez.environments")).
		WithArgs(int64(5)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := s.DeleteEnvironment(context.Background(), 5)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
