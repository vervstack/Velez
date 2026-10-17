//go:build e2e_full

package e2e

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.vervstack.ru/Velez/internal/api/clients/matreshka/pkg/matreshka_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/local_state"
	"go.vervstack.ru/Velez/internal/clients/node_clients/ports"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage/container_derived"
	"go.vervstack.ru/Velez/tests/test_helper"
)

const taskContextTimeout = time.Minute

// restartedEnvironment restarts the Velez container in place: same state file,
// same managed containers, a fresh process.
func restartedEnvironment(t *testing.T, env *TestEnvironment, opts ...TestEnvOpt) *TestEnvironment {
	t.Helper()

	require.Empty(t, opts, "a container restart cannot change the options")

	return restartedContainerEnvironment(t, env)
}

func (e *TestEnvironment) isClusterMode() bool {
	return e.Cfg.Environment.ClusterPgDsn != ""
}

// SecretValue reads a secret from the real store: the instance's containers in
// single-node mode, velez.secrets in cluster mode.
func (e *TestEnvironment) SecretValue(t *testing.T, ref domain.SecretRef) string {
	t.Helper()

	value, isFound := e.lookupSecret(t, ref)
	require.True(t, isFound, "secret %s not found", ref.String())

	return value
}

func (e *TestEnvironment) lookupSecret(t *testing.T, ref domain.SecretRef) (value string, isFound bool) {
	t.Helper()

	if e.isClusterMode() {
		return e.clusterSecret(t, ref)
	}

	switch ref {
	case domain.S3AdminTokenSecretRef(ref.Owner):
		labelFilters := []string{labels.S3InstanceLabel + "=" + ref.Owner}

		return e.lookupContainerEnv(t, labelFilters, container_derived.S3GarageEnvAdminToken)
	case domain.S3RpcSecretRef(ref.Owner):
		labelFilters := []string{labels.S3InstanceLabel + "=" + ref.Owner}

		return e.lookupContainerEnv(t, labelFilters, container_derived.S3GarageEnvRpcSecret)
	case domain.S3WebUiPasswordSecretRef(ref.Owner):
		labelFilters := []string{labels.S3WebUiLabel + "=" + ref.Owner}

		return e.lookupContainerEnv(t, labelFilters, domain.S3WebUiPasswordEnv)
	case container_derived.RegistryInstanceSecretRef(ref.Owner):
		labelFilters := []string{labels.RegistryaasInstanceLabel, labels.VervServiceLabel + "=" + ref.Owner}

		return e.lookupContainerEnv(t, labelFilters, domain.RegistryPasswordEnv)
	}

	t.Skip("a secret held in Velez's memory has no black-box reader")

	return "", false
}

func (e *TestEnvironment) clusterSecret(t *testing.T, ref domain.SecretRef) (value string, isFound bool) {
	t.Helper()

	const secretQuery = `SELECT value FROM velez.secrets WHERE scope = $1 AND owner = $2 AND key = $3`

	row := e.clusterDb(t).QueryRowContext(t.Context(), secretQuery, ref.Scope, ref.Owner, ref.Key)

	err := row.Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false
	}

	require.NoError(t, err)

	return value, true
}

// TaskContext returns the persisted context JSON of a task. The jobs engine
// captures its task storage at startup, before statefull mode switches the
// store to Postgres, so a task is in velez.tasks only when Velez started
// already clustered; otherwise it lives in Velez's memory and has no
// black-box reader.
func (e *TestEnvironment) TaskContext(t *testing.T, entityId, action string) string {
	t.Helper()

	if e.isClusterMode() {
		const contextQuery = `SELECT context FROM velez.tasks WHERE entity_id = $1 AND action = $2`

		var payload []byte

		row := e.clusterDb(t).QueryRowContext(t.Context(), contextQuery, entityId, action)

		err := row.Scan(&payload)
		if err == nil {
			require.NotEmpty(t, payload, "task carries no context")

			return string(payload)
		}

		require.ErrorIs(t, err, sql.ErrNoRows)
	}

	t.Skip("a task held in Velez's memory has no black-box reader")

	return ""
}

// GitlabRunnerId is the id GitLab assigned the runner: velez.runners in
// cluster mode, the runner container's labels in single-node mode.
func (e *TestEnvironment) GitlabRunnerId(t *testing.T, containerName string) int64 {
	t.Helper()

	if e.isClusterMode() {
		const runnerIdQuery = `
SELECT r.gitlab_runner_id
FROM velez.runners r
         JOIN velez.services s ON s.id = r.service_id
WHERE s.name = $1`

		var runnerId int64

		row := e.clusterDb(t).QueryRowContext(t.Context(), runnerIdQuery, containerName)

		err := row.Scan(&runnerId)
		require.NoError(t, err)

		return runnerId
	}

	inspected, err := e.DockerClient().ContainerInspect(t.Context(), containerName)
	require.NoError(t, err)

	runnerId, err := strconv.ParseInt(inspected.Config.Labels[labels.RunnerGitlabIdLabel], 10, 64)
	require.NoError(t, err)

	return runnerId
}

// lookupContainerEnv reads an env var off the container carrying every label
// filter, preferring a running one. isFound is false when no container matches
// or the variable is unset.
func (e *TestEnvironment) lookupContainerEnv(
	t *testing.T, labelFilters []string, envKey string,
) (value string, isFound bool) {
	t.Helper()

	dockerClient := e.DockerClient()

	filterArgs := filters.NewArgs()
	for _, labelFilter := range labelFilters {
		filterArgs.Add("label", labelFilter)
	}

	listOpts := container.ListOptions{All: true, Filters: filterArgs}

	listed, err := dockerClient.ContainerList(t.Context(), listOpts)
	require.NoError(t, err)

	if len(listed) == 0 {
		return "", false
	}

	chosen := listed[0]

	for _, candidate := range listed {
		if candidate.State == "running" {
			chosen = candidate

			break
		}
	}

	inspected, err := dockerClient.ContainerInspect(t.Context(), chosen.ID)
	require.NoError(t, err)

	value = container_derived.EnvValue(inspected.Config.Env, envKey)

	return value, value != ""
}

// poolPortManager is the port pool every environment of the test process
// allocates from.
func poolPortManager(t *testing.T) ports.PortManager {
	t.Helper()

	return test_helper.GetSharedPortManager(t, dindReservedPorts())
}

// persistedLocalState reads the state file Velez keeps its own identity in,
// out of the Velez container.
func (e *TestEnvironment) persistedLocalState(t *testing.T) local_state.State {
	t.Helper()

	var persisted local_state.State

	err := json.Unmarshal(e.containerStateFile(t), &persisted)
	require.NoError(t, err)

	return persisted
}

// dindHttpClient is an http client for fixtures that outlive any one test; its
// addresses must be ip:port.
func dindHttpClient() *http.Client {
	transport := &http.Transport{DialContext: dindDialIp, DisableKeepAlives: true}

	return &http.Client{Transport: transport}
}

type dindDialFunc func(ctx context.Context, network, address string) (net.Conn, error)

// pqDindDialer connects lib/pq through a DinD dialer.
type pqDindDialer struct {
	dial dindDialFunc
}

func (d pqDindDialer) Dial(network, address string) (net.Conn, error) {
	return d.dial(context.Background(), network, address)
}

func (d pqDindDialer) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return d.dial(ctx, network, address)
}

func (d pqDindDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return d.dial(ctx, network, address)
}

// openPostgresViaDind opens a database whose connections go through dial.
func openPostgresViaDind(t *testing.T, dsn string, dial dindDialFunc) *sql.DB {
	t.Helper()

	connector, err := pq.NewConnector(dsn)
	require.NoError(t, err)

	connector.Dialer(pqDindDialer{dial: dial})

	db := sql.OpenDB(connector)

	t.Cleanup(func() {
		closeErr := db.Close()
		require.NoError(t, closeErr)
	})

	return db
}

// clusterDb opens Velez's own cluster database, the DSN it persisted in the
// local state file, through the DinD dialer.
func (e *TestEnvironment) clusterDb(t *testing.T) *sql.DB {
	t.Helper()

	require.True(t, e.isClusterMode(), "clusterDb: the environment has no cluster postgres")

	persisted := e.persistedLocalState(t)
	require.NotEmpty(t, persisted.ClusterState.PgNodeDsn, "statefull is not enabled on this environment")

	return openPostgresViaDind(t, persisted.ClusterState.PgNodeDsn, e.DindDialContext)
}

// BoundServiceName returns the service a container name is bound to under the
// environment on this node, or "" when it is unbound. Cluster mode only.
func (e *TestEnvironment) BoundServiceName(t *testing.T, containerName, environment string) string {
	t.Helper()

	const bindingQuery = `
SELECT s.name
FROM velez.container_bindings b
         JOIN velez.services s ON s.id = b.service_id
WHERE b.node_id = $1
  AND b.environment = $2
  AND b.container_name = $3`

	var serviceName string

	row := e.clusterDb(t).QueryRowContext(t.Context(), bindingQuery, domain.SelfNodeId, environment, containerName)

	err := row.Scan(&serviceName)
	if errors.Is(err, sql.ErrNoRows) {
		return ""
	}

	require.NoError(t, err)

	return serviceName
}

// SpecificationVervPayload returns the verv payload stored for a deployment
// specification, or nil when it has none. Cluster mode only.
func (e *TestEnvironment) SpecificationVervPayload(t *testing.T, specId uint64) []byte {
	t.Helper()

	const payloadQuery = `SELECT verv_payload FROM velez.deployment_specifications WHERE id = $1`

	var payload []byte

	row := e.clusterDb(t).QueryRowContext(t.Context(), payloadQuery, specId)

	err := row.Scan(&payload)
	require.NoError(t, err)

	return payload
}

// BindingIsSidecar reports whether the container's binding under the
// environment on this node is a sidecar one; isBound is false when the
// container has no binding. Cluster mode only.
func (e *TestEnvironment) BindingIsSidecar(
	t *testing.T, containerName, environment string,
) (isSidecar bool, isBound bool) {
	t.Helper()

	const sidecarQuery = `
SELECT is_sidecar
FROM velez.container_bindings
WHERE node_id = $1
  AND environment = $2
  AND container_name = $3`

	row := e.clusterDb(t).QueryRowContext(t.Context(), sidecarQuery, domain.SelfNodeId, environment, containerName)

	err := row.Scan(&isSidecar)
	if errors.Is(err, sql.ErrNoRows) {
		return false, false
	}

	require.NoError(t, err)

	return isSidecar, true
}

// MatreshkaClient talks to the shared matreshka by its in-DinD address.
func (e *TestEnvironment) MatreshkaClient(t *testing.T) matreshka_api.MatreshkaApiClient {
	t.Helper()

	target := "passthrough:///" + dindGatewayAddr(t, dindMatreshkaPort)

	conn, err := grpc.NewClient(target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(e.dindGrpcDialer()))
	require.NoError(t, err)

	t.Cleanup(func() { _ = conn.Close() })

	return matreshka_api.NewMatreshkaApiClient(conn)
}
