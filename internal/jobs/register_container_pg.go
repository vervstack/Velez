package jobs

import (
	"context"
	"database/sql"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka/resources"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/parser"
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/patterns/db_patterns/pg_pattern"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	stepVerifyPgLogin    = "verify_pg_login"
	stepStorePgSecret    = "store_pg_secret"
	stepUpsertPgInstance = "upsert_pg_instance"

	labelTrueValue = "true"

	pgaasSecretScope        = "pgaas"
	pgaasSecretKey          = "password"
	pgaasPendingSecretScope = "pgaas-pending"

	pgSslModeDisable = "disable"
	pgLocalhost      = "localhost"

	registeredPgDefaultDbName = "postgres"
	registeredPgLoginTimeout  = 10 * time.Second
)

type pgLogin struct {
	superuser string
	password  string
	dbName    string
}

type registerPgLoginAccessor interface {
	GetPgSuperuser() string
	GetPgPendingSecretOwner() string
}

type registerPgLoginSetter interface {
	SetPgSuperuser(v string)
}

// resolvePgLogin reads the admin login from the container env; the request
// user and the passwords held in the secrets store fill only what the env
// lacks, the pending (request-supplied) password ahead of the stored one.
func resolvePgLogin(
	containerEnv map[string]string, superuser, pendingPassword, storedPassword string,
) (pgLogin, error) {
	login := pgLogin{
		superuser: containerEnv[pg_pattern.UserEnvVariable],
		password:  containerEnv[pg_pattern.PasswordEnvVariable],
		dbName:    containerEnv[pg_pattern.DbEnvVariable],
	}

	if login.superuser == "" {
		login.superuser = superuser
	}

	if login.password == "" {
		login.password = pendingPassword
	}

	if login.password == "" {
		login.password = storedPassword
	}

	if login.dbName == "" {
		login.dbName = registeredPgDefaultDbName
	}

	if login.superuser == "" || login.password == "" {
		return pgLogin{}, rerrors.Wrap(user_errors.ErrPgCredentialsRequired)
	}

	return login, nil
}

// pgMissingEnv are the login env keys the container lacks, valued from the
// effective login - what a recreated container needs so Velez can read the
// login back from its env.
func pgMissingEnv(containerEnv map[string]string, login pgLogin) map[string]string {
	missing := make(map[string]string)

	if containerEnv[pg_pattern.UserEnvVariable] == "" {
		missing[pg_pattern.UserEnvVariable] = login.superuser
	}

	if containerEnv[pg_pattern.PasswordEnvVariable] == "" {
		missing[pg_pattern.PasswordEnvVariable] = login.password
	}

	return missing
}

// registeredPatternLabels are the labels a pattern adds to the service labels.
func registeredPatternLabels(pattern velez_api.ServicePattern) map[string]string {
	if pattern != velez_api.ServicePattern_SERVICE_PATTERN_POSTGRES {
		return map[string]string{}
	}

	return map[string]string{labels.PgaasInstanceLabel: labelTrueValue}
}

func registeredLabels(serviceName string, pattern velez_api.ServicePattern) map[string]string {
	merged := registeredServiceLabels(serviceName)

	maps.Copy(merged, registeredPatternLabels(pattern))

	return merged
}

func registeredPgSecretRef(serviceName string) domain.SecretRef {
	return domain.SecretRef{Scope: pgaasSecretScope, Owner: serviceName, Key: pgaasSecretKey}
}

// RegisteredPgPendingSecretRef holds the password a register request supplied
// until store_pg_secret moves it to its final ref; owner is the task entity id.
func RegisteredPgPendingSecretRef(owner string) domain.SecretRef {
	return domain.SecretRef{Scope: pgaasPendingSecretScope, Owner: owner, Key: pgaasSecretKey}
}

func readPgSecret(ctx context.Context, store secrets.Store, ref domain.SecretRef) (string, error) {
	value, err := store.Get(ctx, ref)
	if rerrors.Is(err, user_errors.ErrSecretNotFound) {
		return "", nil
	}

	if err != nil {
		return "", rerrors.Wrap(err, "error reading pg secret")
	}

	return value, nil
}

func deletePendingPgSecret(ctx context.Context, store secrets.Store, owner string) error {
	if owner == "" {
		return nil
	}

	err := store.Delete(ctx, RegisteredPgPendingSecretRef(owner))
	if err != nil && !rerrors.Is(err, user_errors.ErrSecretNotFound) {
		return rerrors.Wrap(err, "error deleting pending pg secret")
	}

	return nil
}

// loadPgLogin resolves the login the way verify does; includeStored also
// falls back to the final secret, for the jobs that run after store_pg_secret.
func loadPgLogin(
	ctx context.Context,
	store secrets.Store,
	containerEnv map[string]string,
	req registerRequestAccessor,
	pg registerPgLoginAccessor,
	includeStored bool,
) (pgLogin, error) {
	var pending, stored string

	owner := pg.GetPgPendingSecretOwner()
	if owner != "" {
		value, err := readPgSecret(ctx, store, RegisteredPgPendingSecretRef(owner))
		if err != nil {
			return pgLogin{}, rerrors.Wrap(err)
		}

		pending = value
	}

	if includeStored {
		value, err := readPgSecret(ctx, store, registeredPgSecretRef(req.GetServiceName()))
		if err != nil {
			return pgLogin{}, rerrors.Wrap(err)
		}

		stored = value
	}

	login, err := resolvePgLogin(containerEnv, pg.GetPgSuperuser(), pending, stored)
	if err != nil {
		return pgLogin{}, rerrors.Wrap(err)
	}

	return login, nil
}

// pgContainerHost is the address Velez reaches the container on from inside a
// container: the first network's IP, falling back to the container name.
func pgContainerHost(info container.InspectResponse) string {
	if info.NetworkSettings != nil {
		names := make([]string, 0, len(info.NetworkSettings.Networks))
		for name := range info.NetworkSettings.Networks {
			names = append(names, name)
		}

		slices.Sort(names)

		for _, name := range names {
			endpoint := info.NetworkSettings.Networks[name]
			if endpoint != nil && endpoint.IPAddress != "" {
				return endpoint.IPAddress
			}
		}
	}

	return strings.TrimPrefix(info.Name, "/")
}

// pgLoginDsn addresses the container's Postgres the way getRootDsnJob does:
// in-network address when Velez itself runs in a container, the published
// host port otherwise.
func pgLoginDsn(info container.InspectResponse, login pgLogin, isInContainer bool) (string, error) {
	cfg := &resources.Postgres{
		Host:    pgContainerHost(info),
		Port:    pgDefaultPort,
		User:    login.superuser,
		Pwd:     login.password,
		DbName:  login.dbName,
		SslMode: pgSslModeDisable,
	}

	if !isInContainer {
		port, err := getExposedPgPort(info)
		if err != nil {
			return "", rerrors.Wrap(err, "error getting exposed pg port")
		}

		cfg.Host = pgLocalhost
		cfg.Port = port
	}

	return cfg.ConnectionString() + "&application_name=RegisterContainer", nil
}

func pingPgLogin(ctx context.Context, dsn string) error {
	conn, err := sql.Open(sqldb.Dialect, dsn)
	if err != nil {
		return rerrors.Wrap(err, "error opening connection to postgres")
	}

	defer func() {
		closeErr := conn.Close()
		if closeErr != nil {
			log.Error().Err(closeErr).Msg("error closing postgres connection when verifying login")
		}
	}()

	pingCtx, cancel := context.WithTimeout(ctx, registeredPgLoginTimeout)
	defer cancel()

	err = conn.PingContext(pingCtx)
	if err != nil {
		return rerrors.Wrap(rerrors.Join(user_errors.ErrPgLoginFailed, err))
	}

	return nil
}

// verifyPgLoginJob opens a real connection with the resolved login before
// anything is written or restarted. The password is never put into the
// payload; on failure the pending secret is dropped.
type verifyPgLoginJob struct {
	runtimes container_runtime.RuntimeResolver
	secrets  secrets.Store

	req registerRequestAccessor
	pg  registerPgLoginAccessor
	ctx registerPgLoginSetter
}

func (j *verifyPgLoginJob) Do(ctx context.Context) error {
	err := j.verify(ctx)
	if err == nil {
		return nil
	}

	deleteErr := deletePendingPgSecret(ctx, j.secrets, j.pg.GetPgPendingSecretOwner())
	if deleteErr != nil {
		log.Error().Err(deleteErr).Msg("error deleting pending pg secret after failed login verification")
	}

	return rerrors.Wrap(err)
}

func (j *verifyPgLoginJob) verify(ctx context.Context) error {
	info, err := inspectRegisteredContainer(ctx, j.runtimes, j.req)
	if err != nil {
		return rerrors.Wrap(err)
	}

	login, err := loadPgLogin(ctx, j.secrets, parser.ToDockerEnv(info.Config.Env), j.req, j.pg, false)
	if err != nil {
		return rerrors.Wrap(err)
	}

	dsn, err := pgLoginDsn(info, login, env.IsInContainer())
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = pingPgLogin(ctx, dsn)
	if err != nil {
		return rerrors.Wrap(err)
	}

	j.ctx.SetPgSuperuser(login.superuser)

	return nil
}

func inspectRegisteredContainer(
	ctx context.Context, runtimes container_runtime.RuntimeResolver, req registerRequestAccessor,
) (container.InspectResponse, error) {
	runtime, err := runtimes.Runtime(ctx, req.GetEnvironment())
	if err != nil {
		return container.InspectResponse{}, rerrors.Wrap(err, "error resolving container runtime")
	}

	info, found, err := runtime.InspectAny(ctx, req.GetContainerId())
	if err != nil {
		return container.InspectResponse{}, rerrors.Wrap(err, "error inspecting container")
	}

	if !found || info.Config == nil {
		return container.InspectResponse{}, rerrors.Wrap(user_errors.ErrRegisterContainerNotFound)
	}

	return info, nil
}

// storePgSecretJob writes the final pgaas secret from the resolved login and
// drops the pending one.
type storePgSecretJob struct {
	runtimes container_runtime.RuntimeResolver
	secrets  secrets.Store

	req registerRequestAccessor
	pg  registerPgLoginAccessor
}

func (j *storePgSecretJob) Do(ctx context.Context) error {
	info, err := inspectRegisteredContainer(ctx, j.runtimes, j.req)
	if err != nil {
		return rerrors.Wrap(err)
	}

	login, err := loadPgLogin(ctx, j.secrets, parser.ToDockerEnv(info.Config.Env), j.req, j.pg, false)
	if err != nil {
		return rerrors.Wrap(err)
	}

	ref := registeredPgSecretRef(j.req.GetServiceName())

	err = j.secrets.Put(ctx, ref, login.password)
	if err != nil {
		return rerrors.Wrap(err, "error storing pg password")
	}

	err = deletePendingPgSecret(ctx, j.secrets, j.pg.GetPgPendingSecretOwner())
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

// upsertPgInstanceJob runs right after the bind transaction: PgInstances has
// no WithTx, and the upsert is idempotent.
type upsertPgInstanceJob struct {
	dataStorage storage.Storage
	runtimes    container_runtime.RuntimeResolver
	secrets     secrets.Store

	req     registerRequestAccessor
	pg      registerPgLoginAccessor
	service registeredServiceAccessor
}

func (j *upsertPgInstanceJob) Do(ctx context.Context) error {
	info, err := inspectRegisteredContainer(ctx, j.runtimes, j.req)
	if err != nil {
		return rerrors.Wrap(err)
	}

	login, err := loadPgLogin(ctx, j.secrets, parser.ToDockerEnv(info.Config.Env), j.req, j.pg, true)
	if err != nil {
		return rerrors.Wrap(err)
	}

	upsertReq := domain.UpsertPgInstanceReq{
		ServiceId: j.service.GetServiceId(),
		DbName:    login.dbName,
		Username:  login.superuser,
		SecretRef: registeredPgSecretRef(j.req.GetServiceName()).String(),
		Port:      pgDefaultPort,
	}

	_, err = j.dataStorage.PgInstances().UpsertPgInstance(ctx, upsertReq)
	if err != nil {
		return rerrors.Wrap(err, "error upserting pg instance row")
	}

	return nil
}
