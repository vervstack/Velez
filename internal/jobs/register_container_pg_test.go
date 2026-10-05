package jobs

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/patterns/db_patterns/pg_pattern"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testPgUser     = "admin"
	testPgPassword = "s3cret"
	testPgDbName   = "appdb"

	testPgEnvUser     = "envuser"
	testPgEnvPassword = "envpwd"
)

func Test_ResolvePgLogin_EnvWinsRequestFillsGaps(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		env       map[string]string
		superuser string
		password  string
		want      pgLogin
	}{
		{
			name: "env carries everything",
			env: map[string]string{
				pg_pattern.UserEnvVariable:     testPgEnvUser,
				pg_pattern.PasswordEnvVariable: testPgEnvPassword,
				pg_pattern.DbEnvVariable:       testPgDbName,
			},
			superuser: testPgUser,
			password:  testPgPassword,
			want:      pgLogin{superuser: testPgEnvUser, password: testPgEnvPassword, dbName: testPgDbName},
		},
		{
			name:      "request supplies both, db defaults",
			env:       map[string]string{},
			superuser: testPgUser,
			password:  testPgPassword,
			want:      pgLogin{superuser: testPgUser, password: testPgPassword, dbName: registeredPgDefaultDbName},
		},
		{
			name:      "request fills only the missing password",
			env:       map[string]string{pg_pattern.UserEnvVariable: testPgEnvUser},
			superuser: testPgUser,
			password:  testPgPassword,
			want:      pgLogin{superuser: testPgEnvUser, password: testPgPassword, dbName: registeredPgDefaultDbName},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := resolvePgLogin(tc.env, tc.superuser, tc.password, "")
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func Test_ResolvePgLogin_MissingCredentialsAreRefused(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		env       map[string]string
		superuser string
		password  string
	}{
		{"nothing anywhere", map[string]string{}, "", ""},
		{"password missing", map[string]string{pg_pattern.UserEnvVariable: testPgUser}, "", ""},
		{"user missing", map[string]string{pg_pattern.PasswordEnvVariable: testPgPassword}, "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := resolvePgLogin(tc.env, tc.superuser, tc.password, "")
			require.ErrorIs(t, err, user_errors.ErrPgCredentialsRequired)
		})
	}
}

func Test_PgMissingEnv_StampsOnlyMissingKeys(t *testing.T) {
	t.Parallel()

	login := pgLogin{superuser: testPgUser, password: testPgPassword, dbName: registeredPgDefaultDbName}

	require.Equal(t, map[string]string{
		pg_pattern.UserEnvVariable:     testPgUser,
		pg_pattern.PasswordEnvVariable: testPgPassword,
	}, pgMissingEnv(map[string]string{}, login))

	require.Equal(t, map[string]string{
		pg_pattern.PasswordEnvVariable: testPgPassword,
	}, pgMissingEnv(map[string]string{pg_pattern.UserEnvVariable: testPgEnvUser}, login))

	require.Empty(t, pgMissingEnv(map[string]string{
		pg_pattern.UserEnvVariable: testPgEnvUser, pg_pattern.PasswordEnvVariable: testPgEnvPassword,
	}, login))
}

func Test_RegisteredLabels_PgPatternStampsPgaasLabel(t *testing.T) {
	t.Parallel()

	pgLabels := registeredLabels(testRegisterServiceName, velez_api.ServicePattern_SERVICE_PATTERN_POSTGRES)

	require.Equal(t, map[string]string{
		labels.VervServiceLabel:   testRegisterServiceName,
		labels.DisplayNameLabel:   testRegisterServiceName,
		labels.PgaasInstanceLabel: labelTrueValue,
	}, pgLabels)

	genericLabels := registeredLabels(testRegisterServiceName, velez_api.ServicePattern_SERVICE_PATTERN_UNSPECIFIED)

	require.Equal(t, registeredServiceLabels(testRegisterServiceName), genericLabels)
}

func Test_PgLoginDsn_UsesRuntimeAddress(t *testing.T) {
	t.Parallel()

	login := pgLogin{superuser: testPgUser, password: testPgPassword, dbName: testPgDbName}

	dsn, err := pgLoginDsn("172.17.0.5:5432", login)
	require.NoError(t, err)
	require.Contains(t, dsn, "172.17.0.5")
	require.Contains(t, dsn, "5432")

	publishedDsn, err := pgLoginDsn("sc:30019", login)
	require.NoError(t, err)
	require.Contains(t, publishedDsn, "sc")
	require.Contains(t, publishedDsn, "30019")

	_, err = pgLoginDsn("no-port", login)
	require.Error(t, err)
}

func Test_ResolvePgLogin_PasswordPrecedenceEnvThenPendingThenStored(t *testing.T) {
	t.Parallel()

	envWithPassword := map[string]string{pg_pattern.PasswordEnvVariable: testPgEnvPassword}

	got, err := resolvePgLogin(envWithPassword, testPgUser, "pending", "stored")
	require.NoError(t, err)
	require.Equal(t, testPgEnvPassword, got.password)

	got, err = resolvePgLogin(map[string]string{}, testPgUser, "pending", "stored")
	require.NoError(t, err)
	require.Equal(t, "pending", got.password)

	got, err = resolvePgLogin(map[string]string{}, testPgUser, "", "stored")
	require.NoError(t, err)
	require.Equal(t, "stored", got.password)

	_, err = resolvePgLogin(map[string]string{}, testPgUser, "", "")
	require.ErrorIs(t, err, user_errors.ErrPgCredentialsRequired)
}

func Test_RegisteredPgPendingSecretRef_IsScopedApartFromFinalRef(t *testing.T) {
	t.Parallel()

	pending := RegisteredPgPendingSecretRef("svc/task-1")
	final := registeredPgSecretRef("svc")

	require.Equal(t, "pgaas-pending", pending.Scope)
	require.Equal(t, "svc/task-1", pending.Owner)
	require.Equal(t, "pgaas", final.Scope)
	require.NotEqual(t, pending, final)
}
