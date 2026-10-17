//go:build e2e_full

package e2e

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients/local_state"
	"go.vervstack.ru/Velez/internal/clients/node_clients/ports"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage/container_derived"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
	"go.vervstack.ru/Velez/tests/test_helper"
)

const taskContextTimeout = time.Minute

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

	return e.inMemorySecret(t, ref)
}

// inMemorySecret reads a single-node secret Velez keeps in its process memory
// only (pending register secrets, container-less refs): no black-box reader
// exists, so it goes through the in-process store.
func (e *TestEnvironment) inMemorySecret(t *testing.T, ref domain.SecretRef) (value string, isFound bool) {
	t.Helper()

	value, err := e.Custom.Services.Secrets().Get(t.Context(), ref)
	if rerrors.Is(err, user_errors.ErrSecretNotFound) {
		return "", false
	}

	require.NoError(t, err)

	return value, true
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

// TaskContext returns the persisted context JSON of a task: velez.tasks in
// cluster mode. A single-node task lives in Velez's memory only, so it is read
// through the in-process jobs engine.
func (e *TestEnvironment) TaskContext(t *testing.T, entityId, action string) string {
	t.Helper()

	if e.isClusterMode() {
		const contextQuery = `SELECT context FROM velez.tasks WHERE entity_id = $1 AND action = $2`

		var payload []byte

		row := e.clusterDb(t).QueryRowContext(t.Context(), contextQuery, entityId, action)

		err := row.Scan(&payload)
		require.NoError(t, err)
		require.NotEmpty(t, payload, "task carries no context")

		return string(payload)
	}

	ctx, cancel := context.WithTimeout(t.Context(), taskContextTimeout)
	defer cancel()

	var finalTask tasks_queries.VelezTask

	for task := range e.Custom.JobsEngine.Watch(ctx, entityId, action) {
		finalTask = task
	}

	require.True(t, finalTask.Context.Valid, "task carries no context")

	return string(finalTask.Context.RawMessage)
}

// GitlabRunnerId is the id GitLab assigned the runner, kept on its container's
// labels in single-node mode.
func (e *TestEnvironment) GitlabRunnerId(t *testing.T, containerName string) int64 {
	t.Helper()

	require.False(t, e.isClusterMode(), "GitlabRunnerId: cluster-mode reads are not wired")

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
