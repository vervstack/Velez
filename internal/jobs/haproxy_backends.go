package jobs

import (
	"context"
	"net"
	"net/netip"
	"strconv"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/google/uuid"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage"
)

const (
	pgBackendProbeTimeout    = 90 * time.Second
	pgBackendDefaultDatabase = "postgres"
)

type haproxyBackend struct {
	Name, Address, User, Password        string
	HasCredentials, IsBackup, IsBalanced bool

	host, database, containerId string
	port                        uint32
	network                     container_runtime.NetworkInfo
}

func resolveHaproxyBackend(
	ctx context.Context, dataStorage storage.Storage, runtimes container_runtime.RuntimeResolver,
	backend *velez_api.PgBackend, environment string,
) (haproxyBackend, error) {
	resolved := haproxyBackend{
		Name: backend.GetName(), User: backend.GetUser(), Password: backend.GetPassword(), database: pgBackendDefaultDatabase,
		IsBackup:   backend.GetRole() == velez_api.PgBackend_BACKUP,
		IsBalanced: backend.GetRole() == velez_api.PgBackend_BALANCED,
	}

	target := backend.GetConnectionTarget()
	if backend != nil && backend.Environment != nil {
		environment = backend.GetEnvironment()
	}

	var err error

	switch source := backend.GetSource().(type) {
	case *velez_api.PgBackend_VelezService:
		resolved, err = resolveManagedHaproxyBackend(
			ctx, dataStorage, runtimes, resolved, source.VelezService, environment, target,
		)
	case *velez_api.PgBackend_Endpoint:
		resolved, err = resolveOwnHaproxyBackend(ctx, runtimes, resolved, source.Endpoint, environment, target)
	default:
		err = user_errors.New("Select a PostgreSQL instance or enter a server address")
	}

	if err != nil {
		return haproxyBackend{}, rerrors.Wrap(err)
	}

	resolved.Address = net.JoinHostPort(resolved.host, strconv.FormatUint(uint64(resolved.port), 10))

	return resolved, nil
}

func resolveManagedHaproxyBackend(
	ctx context.Context, dataStorage storage.Storage, runtimes container_runtime.RuntimeResolver,
	backend haproxyBackend, serviceName, environment string, target *velez_api.HaproxyConnectionTarget,
) (haproxyBackend, error) {
	if target.GetPublicEndpoint() != nil || target.GetTailscaleEndpoint() != nil {
		return haproxyBackend{}, user_errors.New("Velez PostgreSQL instances use their Docker network")
	}

	pg, err := haproxyMasterInstance(ctx, dataStorage, serviceName)
	if err != nil {
		return haproxyBackend{}, rerrors.Wrap(err)
	}

	runtime, err := runtimes.Runtime(ctx, environment)
	if err != nil {
		return haproxyBackend{}, rerrors.Wrap(err)
	}

	shared, err := resolveHaproxyDockerTarget(ctx, runtime, serviceName, target.GetDockerNetwork())
	if err != nil {
		return haproxyBackend{}, rerrors.Wrap(err)
	}

	backend.network, err = runtime.InspectNetwork(ctx, shared.networkId)
	if err != nil {
		return haproxyBackend{}, rerrors.Wrap(err)
	}

	backend.host = shared.host
	backend.port = domain.HaproxyDefaultPort
	backend.containerId = shared.containerId
	backend.User = pg.Username

	if pg.DbName != "" {
		backend.database = pg.DbName
	}

	return backend, nil
}

func resolveOwnHaproxyBackend(
	ctx context.Context, runtimes container_runtime.RuntimeResolver, backend haproxyBackend,
	endpoint *velez_api.HaproxyConnectionTarget_Endpoint, environment string, target *velez_api.HaproxyConnectionTarget,
) (haproxyBackend, error) {
	host := endpoint.GetHost()
	ip, err := netip.ParseAddr(host)

	if (err != nil || ip.Zone() != "") && !haproxyMasterNamePattern.MatchString(host) {
		return haproxyBackend{}, user_errors.New("Enter a valid PostgreSQL host")
	}

	port := endpoint.GetPort()
	if port == 0 || port > domain.HaproxyMaxPort {
		return haproxyBackend{}, user_errors.New("Enter a PostgreSQL port between 1 and 65535")
	}

	if backend.User == "" || backend.Password == "" {
		return haproxyBackend{}, user_errors.New("Enter a PostgreSQL user and password")
	}

	if target.GetPublicEndpoint() != nil || target.GetTailscaleEndpoint() != nil {
		err = validateHaproxyBackendTarget(target, endpoint)
		if err != nil {
			return haproxyBackend{}, rerrors.Wrap(err)
		}
	}

	backend.host = host
	backend.port = port
	backend.HasCredentials = true

	if target != nil && target.DockerNetwork != nil {
		networkName := target.GetDockerNetwork()
		if networkName == "" {
			return haproxyBackend{}, user_errors.New("Select a Docker network")
		}

		runtime, resolveErr := runtimes.Runtime(ctx, environment)
		if resolveErr != nil {
			return haproxyBackend{}, rerrors.Wrap(resolveErr)
		}

		backend.network, err = runtime.InspectNetwork(ctx, networkName)
		if err != nil {
			return haproxyBackend{}, rerrors.Wrap(err)
		}

		if backend.network.IsSystem || !backend.network.IsIccEnabled {
			return haproxyBackend{}, user_errors.New("Select a Docker network that allows container communication")
		}
	}

	return backend, nil
}

func ValidatePgBackend(
	ctx context.Context, dataStorage storage.Storage, secretStore secrets.Store,
	runtimes container_runtime.RuntimeResolver, docker node_clients.Docker, req *velez_api.ValidatePgBackend_Request,
) error {
	ctx, cancel := context.WithTimeout(ctx, pgBackendProbeTimeout)
	defer cancel()

	backend := req.GetBackend()

	resolved, err := resolveHaproxyBackend(ctx, dataStorage, runtimes, backend, "")
	if err != nil {
		return rerrors.Wrap(err)
	}

	mode := container.NetworkMode("bridge")
	if resolved.network.DockerName != "" {
		mode = container.NetworkMode(resolved.network.DockerName)
	}

	if backend.GetConnectionTarget().GetTailscaleEndpoint() != nil {
		sidecar, inspectErr := inspectHaproxySidecar(ctx, runtimes, req.GetSidecarContainerId(), "")
		if inspectErr != nil {
			return rerrors.Wrap(inspectErr)
		}

		mode = container.NetworkMode(networkModeOfPrefix + sidecar.ID)
	}

	if backend.GetVelezService() != "" {
		err = loadHaproxyProbeCredentials(ctx, dataStorage, secretStore, docker, backend.GetVelezService(), &resolved)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	runtime, err := runtimes.Runtime(ctx, "")
	if err != nil {
		return rerrors.Wrap(err)
	}

	return runHaproxyProbe(ctx, runtime, docker, resolved, mode)
}

func loadHaproxyProbeCredentials(
	ctx context.Context, dataStorage storage.Storage, secretStore secrets.Store, docker node_clients.Docker,
	serviceName string, backend *haproxyBackend,
) error {
	credentials, err := readPgFailoverCredentialsById(ctx, docker, backend.containerId)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if credentials.Password != "" {
		backend.User, backend.Password, backend.database = credentials.Username, credentials.Password, credentials.DbName

		return nil
	}

	pg, err := haproxyMasterInstance(ctx, dataStorage, serviceName)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if pg.SecretRef == "" {
		return nil
	}

	ref, err := domain.ParseSecretRef(pg.SecretRef)
	if err != nil {
		return rerrors.Wrap(err)
	}

	backend.Password, err = secretStore.Get(ctx, ref)
	if err != nil && !rerrors.Is(err, user_errors.ErrSecretNotFound) {
		return rerrors.Wrap(err, "error reading PostgreSQL credentials")
	}

	return nil
}

func runHaproxyProbe(
	ctx context.Context, runtime container_runtime.ContainerRuntime, docker node_clients.Docker,
	backend haproxyBackend, mode container.NetworkMode,
) error {
	_, err := runtime.PullImage(ctx, replicationExecutorImage)
	if err != nil {
		return rerrors.Wrap(err, "error pulling PostgreSQL probe image")
	}

	config := &container.Config{
		Image: replicationExecutorImage,
		Cmd:   []string{containerSleepCommand, containerSleepDuration},
	}
	host := &container.HostConfig{NetworkMode: mode}
	req := container_runtime.ContainerCreateRequest{
		Config:        &container_runtime.ContainerConfig{Config: config},
		HostConfig:    &container_runtime.HostConfig{HostConfig: host},
		ContainerName: "velez-pg-probe-" + uuid.NewString(),
	}

	created, err := runtime.ContainerCreate(ctx, req)
	if err != nil {
		return rerrors.Wrap(err, "error creating PostgreSQL probe container")
	}

	defer removeHaproxyProbe(ctx, runtime, created.ID)

	start := container.StartOptions{}

	err = docker.Client().ContainerStart(ctx, created.ID, start)
	if err != nil {
		return rerrors.Wrap(err, "error starting PostgreSQL probe")
	}

	opts := container.ExecOptions{
		Cmd: []string{"psql", "-X", "-w", "-v", "ON_ERROR_STOP=1", "-c", "select 1"},
		Env: []string{
			"PGHOST=" + backend.host, "PGPORT=" + strconv.FormatUint(uint64(backend.port), 10),
			"PGUSER=" + backend.User, "PGPASSWORD=" + backend.Password, "PGDATABASE=" + backend.database,
			"PGCONNECT_TIMEOUT=5",
		},
		AttachStdout: true, AttachStderr: true,
	}

	if backend.Password == "" {
		opts.Cmd = []string{"pg_isready", "-t", "5"}
	}

	_, exit, err := runtime.Exec(ctx, created.ID, opts)
	if err != nil {
		return rerrors.Wrap(err, "could not run PostgreSQL probe")
	}

	if exit != 0 {
		return user_errors.New("PostgreSQL did not accept the connection. Check the network and login details.")
	}

	return nil
}

func removeHaproxyProbe(ctx context.Context, runtime container_runtime.ContainerRuntime, id string) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), executorRemovalTimeout)
	defer cancel()

	_ = runtime.Remove(ctx, id)
}

func validateHaproxyBackendTarget(
	target *velez_api.HaproxyConnectionTarget, endpoint *velez_api.HaproxyConnectionTarget_Endpoint,
) error {
	if target.DockerNetwork != nil || (target.GetPublicEndpoint() != nil && target.GetTailscaleEndpoint() != nil) {
		return user_errors.New("Choose one PostgreSQL connection target")
	}

	configured := target.GetPublicEndpoint()
	if configured == nil {
		configured = target.GetTailscaleEndpoint()
	}

	if configured.GetHost() != endpoint.GetHost() || configured.GetPort() != endpoint.GetPort() {
		return user_errors.New("PostgreSQL source and connection target must use the same address")
	}

	return nil
}
