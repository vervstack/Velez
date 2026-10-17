//go:build e2e_full

package e2e

import (
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/clients/node_clients/local_state"
	"go.vervstack.ru/Velez/internal/clients/node_clients/ports"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage/container_derived"
	"go.vervstack.ru/Velez/tests/test_helper"
)

// restartedEnvironment stops env and starts a fresh app over the same docker
// daemon, container suffix and local state file: a restarted Velez. Containers
// created by env survive; everything env kept in memory is gone.
func restartedEnvironment(t *testing.T, env *TestEnvironment, opts ...TestEnvOpt) *TestEnvironment {
	t.Helper()

	err := env.stop()
	require.NoError(t, err)

	restartOpts := append([]TestEnvOpt{asRestartOf(env)}, opts...)

	return NewEnvironment(t, restartOpts...)
}

func asRestartOf(prev *TestEnvironment) TestEnvOpt {
	return func(a *TestEnvironment) {
		a.isRestart = true
		a.dockerHost = prev.dockerHost
		a.velezKey = prev.velezKey

		a.Cfg.Environment.LocalStatePath = prev.Cfg.Environment.LocalStatePath
		a.Cfg.Environment.ContainerSuffix = prev.Cfg.Environment.ContainerSuffix
		a.Cfg.Environment.ClusterPgDsn = prev.Cfg.Environment.ClusterPgDsn
		a.Cfg.Environment.Environments = prev.Cfg.Environment.Environments
	}
}

func (e *TestEnvironment) isClusterMode() bool {
	return e.Cfg.Environment.ClusterPgDsn != ""
}

// SecretValue reads an S3 instance secret from the real store: the instance's
// containers in single-node mode. Cluster mode (velez.secrets) is not wired.
func (e *TestEnvironment) SecretValue(t *testing.T, ref domain.SecretRef) string {
	t.Helper()

	require.False(t, e.isClusterMode(), "SecretValue: cluster-mode secret reads are not wired")

	var (
		containerLabel string
		envKey         string
	)

	switch ref {
	case domain.S3AdminTokenSecretRef(ref.Owner):
		containerLabel, envKey = labels.S3InstanceLabel, container_derived.S3GarageEnvAdminToken
	case domain.S3RpcSecretRef(ref.Owner):
		containerLabel, envKey = labels.S3InstanceLabel, container_derived.S3GarageEnvRpcSecret
	case domain.S3WebUiPasswordSecretRef(ref.Owner):
		containerLabel, envKey = labels.S3WebUiLabel, domain.S3WebUiPasswordEnv
	default:
		require.Failf(t, "unsupported secret", "SecretValue does not read %s", ref.String())
	}

	return e.containerEnvValue(t, containerLabel+"="+ref.Owner, envKey)
}

func (e *TestEnvironment) containerEnvValue(t *testing.T, labelFilter, envKey string) string {
	t.Helper()

	dockerClient := e.DockerClient()

	listOpts := container.ListOptions{All: true, Filters: filters.NewArgs(filters.Arg("label", labelFilter))}

	listed, err := dockerClient.ContainerList(t.Context(), listOpts)
	require.NoError(t, err)
	require.NotEmpty(t, listed, "no container labelled %s", labelFilter)

	chosen := listed[0]

	for _, candidate := range listed {
		if candidate.State == "running" {
			chosen = candidate

			break
		}
	}

	inspected, err := dockerClient.ContainerInspect(t.Context(), chosen.ID)
	require.NoError(t, err)

	return container_derived.EnvValue(inspected.Config.Env, envKey)
}

// poolPortManager is the port pool every environment of the test process
// allocates from.
func poolPortManager(t *testing.T) ports.PortManager {
	t.Helper()

	return test_helper.GetSharedPortManager(t, dindAvailablePorts())
}

// clusterDb opens Velez's own cluster database, the DSN it persisted in the
// local state file.
func (e *TestEnvironment) clusterDb(t *testing.T) *sql.DB {
	t.Helper()

	require.True(t, e.isClusterMode(), "clusterDb: the environment has no cluster postgres")

	stateFile, err := os.Open(e.Cfg.Environment.LocalStatePath)
	require.NoError(t, err)

	defer func() {
		closeErr := stateFile.Close()
		require.NoError(t, closeErr)
	}()

	var persisted local_state.State

	err = json.NewDecoder(stateFile).Decode(&persisted)
	require.NoError(t, err)
	require.NotEmpty(t, persisted.ClusterState.PgNodeDsn, "statefull is not enabled on this environment")

	db, err := sql.Open("postgres", persisted.ClusterState.PgNodeDsn)
	require.NoError(t, err)

	t.Cleanup(func() {
		closeErr := db.Close()
		require.NoError(t, closeErr)
	})

	return db
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
