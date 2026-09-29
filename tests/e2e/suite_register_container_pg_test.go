//go:build e2e_full

package e2e

import (
	"context"
	"database/sql"
	"io"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"
	"go.vervstack.ru/matreshka/pkg/matreshka/resources"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	pgRegUser     = "e2e_admin"
	pgRegPassword = "e2e-secret-pw"
	pgRegDbName   = "e2e_db"

	pgRegAllInterfaces = "0.0.0.0"
	pgRegLoopback      = "127.0.0.1"

	pgRegWrongPassword = "e2e-wrong-pw"

	pgRegPort = nat.Port("5432/tcp")

	pgRegPortEnvSingle     = dindPortBandEnd - 5
	pgRegPortFailures      = dindPortBandEnd - 4
	pgRegPortRequestSingle = dindPortBandEnd - 3
	pgRegPortClusterEnv    = dindPortBandEnd - 7
	pgRegPortClusterReq    = dindPortBandEnd - 6

	pgRegReadyTimeout = 90 * time.Second

	pgRegCreateTable = "CREATE TABLE IF NOT EXISTS e2e_survive (id INT PRIMARY KEY, note TEXT NOT NULL)"
	pgRegInsertRow   = "INSERT INTO e2e_survive (id, note) VALUES (1, 'survived-register')"
	pgRegSelectNote  = "SELECT note FROM e2e_survive WHERE id = 1"
	pgRegSurvivedRow = "survived-register"

	pgRegSecretScope = "pgaas"
	pgRegSecretKey   = "password"

	// The hidden login is exported inside the container's shell, so the
	// login reaches Postgres but never appears in Config.Env.
	pgRegHiddenLoginScript = "export POSTGRES_USER=" + pgRegUser +
		" POSTGRES_PASSWORD=" + pgRegPassword + "; exec docker-entrypoint.sh postgres"
)

type foreignPg struct {
	foreignContainer

	port int
}

func pgRegNames(prefix string) (containerName, serviceName string) {
	return "e2e_regpg_" + prefix + "_c", "e2e_regpg_" + prefix + "_svc"
}

func pullPgImage(t *testing.T, dockerClient client.APIClient) {
	t.Helper()

	pullReader, err := dockerClient.ImagePull(t.Context(), PostgresImage, image.PullOptions{})
	require.NoError(t, err)

	_, err = io.Copy(io.Discard, pullReader)
	require.NoError(t, err)

	err = pullReader.Close()
	require.NoError(t, err)
}

func newForeignPgConfig(t *testing.T, isLoginInEnv bool) *container.Config {
	t.Helper()

	cfg := &container.Config{
		Image:  PostgresImage,
		Labels: map[string]string{testCaseNameLabel: t.Name()},
	}

	if isLoginInEnv {
		cfg.Env = []string{
			"POSTGRES_USER=" + pgRegUser,
			"POSTGRES_PASSWORD=" + pgRegPassword,
			"POSTGRES_DB=" + pgRegDbName,
		}

		return cfg
	}

	cfg.Cmd = []string{"sh", "-c", pgRegHiddenLoginScript}

	return cfg
}

// startForeignPg runs a Postgres container Velez did not create, publishing
// 5432 on the fixed DinD band port and forwarding that port on this host: the
// Velez under test runs as a bare binary, so it dials localhost:<port> as it
// would beside a local Docker daemon, while the daemon here lives in DinD.
func startForeignPg(t *testing.T, env *TestEnvironment, name string, port int, isLoginInEnv bool) foreignPg {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	pullPgImage(t, dockerClient)

	cfg := newForeignPgConfig(t, isLoginInEnv)

	binding := nat.PortBinding{HostIP: pgRegAllInterfaces, HostPort: strconv.Itoa(port)}
	hostCfg := &container.HostConfig{PortBindings: nat.PortMap{pgRegPort: []nat.PortBinding{binding}}}

	t.Cleanup(func() {
		removeOpts := container.RemoveOptions{Force: true, RemoveVolumes: true}

		_ = dockerClient.ContainerRemove(context.Background(), name, removeOpts)
	})

	created, err := dockerClient.ContainerCreate(t.Context(), cfg, hostCfg, nil, nil, name)
	require.NoError(t, err)

	err = dockerClient.ContainerStart(t.Context(), created.ID, container.StartOptions{})
	require.NoError(t, err)

	forwardDindPort(t, port)

	return foreignPg{foreignContainer: foreignContainer{id: created.ID, name: name}, port: port}
}

func forwardDindPort(t *testing.T, port int) {
	t.Helper()

	target, ok := sharedDind.Addr(port)
	require.True(t, ok, "dind did not publish port %d", port)

	listenCfg := &net.ListenConfig{}

	lis, err := listenCfg.Listen(t.Context(), "tcp", net.JoinHostPort(pgRegLoopback, strconv.Itoa(port)))
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = lis.Close()
	})

	go acceptForwarded(t.Context(), lis, target)
}

func acceptForwarded(ctx context.Context, lis net.Listener, target string) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			return
		}

		go pipeForwarded(ctx, conn, target)
	}
}

func pipeForwarded(ctx context.Context, conn net.Conn, target string) {
	dialer := &net.Dialer{}

	upstream, err := dialer.DialContext(ctx, "tcp", target)
	if err != nil {
		_ = conn.Close()

		return
	}

	go func() {
		_, _ = io.Copy(upstream, conn)
		_ = upstream.Close()
	}()

	_, _ = io.Copy(conn, upstream)
	_ = conn.Close()
}

func openDindPg(t *testing.T, port int, user, password, dbName string) *sql.DB {
	t.Helper()

	addr, ok := sharedDind.Addr(port)
	require.True(t, ok, "dind did not publish port %d", port)

	host, portStr, err := net.SplitHostPort(addr)
	require.NoError(t, err)

	hostPort, err := strconv.ParseUint(portStr, 10, 64)
	require.NoError(t, err)

	cfg := &resources.Postgres{
		Host:    host,
		Port:    hostPort,
		User:    user,
		Pwd:     password,
		DbName:  dbName,
		SslMode: "disable",
	}

	db, err := sql.Open(sqldb.Dialect, cfg.ConnectionString())
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func awaitPgReady(t *testing.T, db *sql.DB) {
	t.Helper()

	require.Eventually(t, func() bool {
		return db.PingContext(t.Context()) == nil
	}, pgRegReadyTimeout, time.Second, "postgres never became reachable")
}

func writeSurvivalRow(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.ExecContext(t.Context(), pgRegCreateTable)
	require.NoError(t, err)

	_, err = db.ExecContext(t.Context(), pgRegInsertRow)
	require.NoError(t, err)
}

func requireSurvivalRow(t *testing.T, db *sql.DB) {
	t.Helper()

	var note string

	err := db.QueryRowContext(t.Context(), pgRegSelectNote).Scan(&note)
	require.NoError(t, err, "data written before register is gone")
	require.Equal(t, pgRegSurvivedRow, note)
}

func publishedPgHostPort(t *testing.T, inspected container.InspectResponse) int {
	t.Helper()

	require.NotNil(t, inspected.NetworkSettings)

	bindings := inspected.NetworkSettings.Ports[pgRegPort]
	require.NotEmpty(t, bindings, "recreated container publishes no 5432")

	port, err := strconv.Atoi(bindings[0].HostPort)
	require.NoError(t, err)

	return port
}

func newPgRegisterRequest(
	containerId, serviceName string, superuser, password *string,
) *velez_api.RegisterContainer_Request {
	req := newRegisterContainerRequest(containerId, serviceName)

	pattern := &velez_api.RegisterContainer_Request_PgPattern{
		Superuser: superuser,
		Password:  password,
	}

	req.Pattern = &velez_api.RegisterContainer_Request_Pg{Pg: pattern}

	return req
}

func pgRegSecretRef(serviceName string) domain.SecretRef {
	return domain.SecretRef{Scope: pgRegSecretScope, Owner: serviceName, Key: pgRegSecretKey}
}

func requireNoPendingSecret(t *testing.T, env *TestEnvironment, entityId string) {
	t.Helper()

	ref := jobs.RegisteredPgPendingSecretRef(entityId)

	_, err := env.Custom.Services.Secrets().Get(t.Context(), ref)
	require.True(t, rerrors.Is(err, user_errors.ErrSecretNotFound), "pending secret must be gone, got %v", err)
}

func requireNoPasswordInPayload(t *testing.T, task tasks_queries.VelezTask, passwords ...string) {
	t.Helper()

	require.True(t, task.Context.Valid, "task carries no context")

	payload := string(task.Context.RawMessage)

	for _, password := range passwords {
		require.NotContains(t, payload, password, "task payload must never carry the password")
	}
}

func requireRegisterNothingChanged(
	t *testing.T,
	env *TestEnvironment,
	foreign foreignPg,
	serviceName string,
	resp *velez_api.RegisterContainer_Response,
) {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	inspected, err := dockerClient.ContainerInspect(t.Context(), foreign.name)
	require.NoError(t, err)
	require.Equal(t, foreign.id, inspected.ID, "a failed register must not recreate the container")
	require.True(t, inspected.State.Running)
	require.Empty(t, inspected.Config.Labels[labels.VervServiceLabel])

	listed := findDockerContainer(t, env, foreign.name)
	require.False(t, listed.GetIsRegistered())

	svcResp, err := env.ServiceApiClient().ListServices(t.Context(), newListServicesRequest())
	require.NoError(t, err)

	for _, svc := range svcResp.GetServices() {
		require.NotEqual(t, serviceName, svc.GetName(), "a failed register must leave no service row")
	}

	require.Nil(t, findPgInstance(t, env, serviceName), "a failed register must leave no pg instance")

	requireNoPendingSecret(t, env, resp.GetEntityId())
}

func Test_RegisterContainer_Pg_SingleMode_EnvCredentials(t *testing.T) {
	t.Parallel()

	containerName, serviceName := pgRegNames("single_env")

	env := Planes[0].NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	foreign := startForeignPg(t, env, containerName, pgRegPortEnvSingle, true)

	before := openDindPg(t, foreign.port, pgRegUser, pgRegPassword, pgRegDbName)
	awaitPgReady(t, before)
	writeSurvivalRow(t, before)

	suggested := findDockerContainer(t, env, containerName)
	require.Equal(t, velez_api.ServicePattern_SERVICE_PATTERN_POSTGRES, suggested.GetSuggestedPattern())
	require.False(t, suggested.GetIsRegistered())

	req := newPgRegisterRequest(foreign.id, serviceName, nil, nil)

	req.KeepPortMapping = true

	resp := registerContainer(t, env, req)

	task := awaitRegisterTask(t, env, resp)
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "register task error: %s", task.Error.String)

	inspected, err := dockerClient.ContainerInspect(t.Context(), containerName)
	require.NoError(t, err)
	require.True(t, inspected.State.Running)
	require.NotEqual(t, foreign.id, inspected.ID, "single mode recreates the container")
	require.Equal(t, serviceName, inspected.Config.Labels[labels.VervServiceLabel])
	require.Equal(t, labelValueTrue, inspected.Config.Labels[labels.PgaasInstanceLabel])

	after := openDindPg(t, publishedPgHostPort(t, inspected), pgRegUser, pgRegPassword, pgRegDbName)
	awaitPgReady(t, after)
	requireSurvivalRow(t, after)

	listed := findDockerContainer(t, env, containerName)
	require.True(t, listed.GetIsRegistered())
	require.Equal(t, serviceName, listed.GetLinkedServiceName())

	instance := findPgInstance(t, env, serviceName)
	require.NotNil(t, instance, "registered pg must be visible in ListPgInstances")
	require.Equal(t, pgRegUser, instance.GetUsername())
	require.Equal(t, pgRegDbName, instance.GetDbName())

	requireNoPendingSecret(t, env, resp.GetEntityId())
}

func Test_RegisterContainer_Pg_SingleMode_RequestCredentials(t *testing.T) {
	t.Parallel()

	containerName, serviceName := pgRegNames("single_req")

	env := Planes[0].NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	foreign := startForeignPg(t, env, containerName, pgRegPortRequestSingle, false)

	before := openDindPg(t, foreign.port, pgRegUser, pgRegPassword, "postgres")
	awaitPgReady(t, before)

	req := newPgRegisterRequest(foreign.id, serviceName, toolbox.ToPtr(pgRegUser), toolbox.ToPtr(pgRegPassword))
	resp := registerContainer(t, env, req)

	task := awaitRegisterTask(t, env, resp)
	require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "register task error: %s", task.Error.String)
	requireNoPasswordInPayload(t, task, pgRegPassword)

	inspected, err := dockerClient.ContainerInspect(t.Context(), containerName)
	require.NoError(t, err)
	require.True(t, inspected.State.Running)
	require.NotEqual(t, foreign.id, inspected.ID)
	require.Contains(t, inspected.Config.Env, "POSTGRES_USER="+pgRegUser)
	require.Contains(t, inspected.Config.Env, "POSTGRES_PASSWORD="+pgRegPassword)

	instance := findPgInstance(t, env, serviceName)
	require.NotNil(t, instance)
	require.Equal(t, pgRegUser, instance.GetUsername())

	requireNoPendingSecret(t, env, resp.GetEntityId())
}

func Test_RegisterContainer_Pg_LoginFailuresChangeNothing(t *testing.T) {
	t.Parallel()

	containerName, serviceName := pgRegNames("failures")

	env := Planes[0].NewEnvironment(t)

	foreign := startForeignPg(t, env, containerName, pgRegPortFailures, false)

	probe := openDindPg(t, foreign.port, pgRegUser, pgRegPassword, "postgres")
	awaitPgReady(t, probe)

	t.Run("wrong password", func(t *testing.T) {
		req := newPgRegisterRequest(
			foreign.id, serviceName, toolbox.ToPtr(pgRegUser), toolbox.ToPtr(pgRegWrongPassword))
		resp := registerContainer(t, env, req)

		task := awaitRegisterTask(t, env, resp)
		require.Equal(t, tasks_queries.VelezTaskStatusFAILED, task.Status)
		require.Contains(t, task.Error.String, user_errors.ErrPgLoginFailed.Error())
		requireNoPasswordInPayload(t, task, pgRegWrongPassword)

		requireRegisterNothingChanged(t, env, foreign, serviceName, resp)
	})

	t.Run("credentials missing", func(t *testing.T) {
		req := newPgRegisterRequest(foreign.id, serviceName, nil, nil)
		resp := registerContainer(t, env, req)

		task := awaitRegisterTask(t, env, resp)
		require.Equal(t, tasks_queries.VelezTaskStatusFAILED, task.Status)
		require.Contains(t, task.Error.String, user_errors.ErrPgCredentialsRequired.Error())

		requireRegisterNothingChanged(t, env, foreign, serviceName, resp)
	})
}

func Test_RegisterContainer_Pg_ClusterMode(t *testing.T) {
	env, _ := enableStatefullPgUnderDind(t, Planes[1], registerClusterSuffix+"-pg")
	dockerClient := env.Custom.NodeClients.Docker().Client()

	t.Run("env credentials", func(t *testing.T) {
		containerName, serviceName := pgRegNames("cluster_env")

		foreign := startForeignPg(t, env, containerName, pgRegPortClusterEnv, true)

		probe := openDindPg(t, foreign.port, pgRegUser, pgRegPassword, pgRegDbName)
		awaitPgReady(t, probe)

		req := newPgRegisterRequest(foreign.id, serviceName, nil, nil)
		resp := registerContainer(t, env, req)

		task := awaitRegisterTask(t, env, resp)
		require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "register task error: %s", task.Error.String)
		requireNoPasswordInPayload(t, task, pgRegPassword)

		requirePgClusterRegistered(t, env, dockerClient, foreign, serviceName, resp, pgRegPassword)
	})

	t.Run("request credentials", func(t *testing.T) {
		containerName, serviceName := pgRegNames("cluster_req")

		foreign := startForeignPg(t, env, containerName, pgRegPortClusterReq, false)

		probe := openDindPg(t, foreign.port, pgRegUser, pgRegPassword, "postgres")
		awaitPgReady(t, probe)

		req := newPgRegisterRequest(foreign.id, serviceName, toolbox.ToPtr(pgRegUser), toolbox.ToPtr(pgRegPassword))
		resp := registerContainer(t, env, req)

		task := awaitRegisterTask(t, env, resp)
		require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "register task error: %s", task.Error.String)
		requireNoPasswordInPayload(t, task, pgRegPassword)

		requirePgClusterRegistered(t, env, dockerClient, foreign, serviceName, resp, pgRegPassword)
	})
}

func requirePgClusterRegistered(
	t *testing.T,
	env *TestEnvironment,
	dockerClient client.APIClient,
	foreign foreignPg,
	serviceName string,
	resp *velez_api.RegisterContainer_Response,
	wantPassword string,
) {
	t.Helper()

	inspected, err := dockerClient.ContainerInspect(t.Context(), foreign.name)
	require.NoError(t, err)
	require.Equal(t, foreign.id, inspected.ID, "cluster mode must not recreate the container")
	require.True(t, inspected.State.Running)

	instance := findPgInstance(t, env, serviceName)
	require.NotNil(t, instance, "registered pg must be visible in ListPgInstances")
	require.Equal(t, pgRegUser, instance.GetUsername())

	secret, err := env.Custom.Services.Secrets().Get(t.Context(), pgRegSecretRef(serviceName))
	require.NoError(t, err)
	require.Equal(t, wantPassword, secret)

	credsReq := &velez_api.GetPgInstanceCredentials_Request{Name: serviceName}

	creds, err := env.Custom.PgaasApiImpl.GetPgInstanceCredentials(t.Context(), credsReq)
	require.NoError(t, err)
	require.Equal(t, wantPassword, creds.GetPassword())

	requireNoPendingSecret(t, env, resp.GetEntityId())
}

func Test_ListContainers_SuggestedPattern(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t)

	pgName, _ := pgRegNames("suggest")
	nginxName := "e2e_regpg_suggest_nginx"

	startForeignPgWithoutPort(t, env, pgName)
	startForeignContainer(t, env, nginxName, false)

	pgListed := findDockerContainer(t, env, pgName)
	require.Equal(t, velez_api.ServicePattern_SERVICE_PATTERN_POSTGRES, pgListed.GetSuggestedPattern())

	nginxListed := findDockerContainer(t, env, nginxName)
	require.Equal(t, velez_api.ServicePattern_SERVICE_PATTERN_UNSPECIFIED, nginxListed.GetSuggestedPattern())
}

func startForeignPgWithoutPort(t *testing.T, env *TestEnvironment, name string) {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	pullPgImage(t, dockerClient)

	cfg := newForeignPgConfig(t, true)

	t.Cleanup(func() {
		removeOpts := container.RemoveOptions{Force: true, RemoveVolumes: true}

		_ = dockerClient.ContainerRemove(context.Background(), name, removeOpts)
	})

	created, err := dockerClient.ContainerCreate(t.Context(), cfg, nil, nil, nil, name)
	require.NoError(t, err)

	err = dockerClient.ContainerStart(t.Context(), created.ID, container.StartOptions{})
	require.NoError(t, err)
}
