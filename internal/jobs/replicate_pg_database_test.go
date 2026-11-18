package jobs

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/tests/test_helper"
)

func Test_ReplicatePgDatabasePayload_JsonRoundTrip(t *testing.T) {
	t.Parallel()

	payload := &velez_api.ReplicatePgDatabaseTaskPayload{
		SourceInstanceId: "pgaas_source",
		SourceHost:       "source.example.test",
		SourcePort:       5432,
		SourceUser:       "source_user",
		SourcePassword:   "source_password",
		SourceDbName:     "source_db",
		TargetInstanceId: "pgaas_target",
		TargetHost:       "target.example.test",
		TargetPort:       6432,
		TargetUser:       "target_user",
		TargetPassword:   "target_password",
	}

	encoded, err := json.Marshal(payload)
	require.NoError(t, err)

	var restored velez_api.ReplicatePgDatabaseTaskPayload

	err = json.Unmarshal(encoded, &restored)
	require.NoError(t, err)
	require.Equal(t, payload, &restored)
}

func Test_CheckReplicationTargetJob_FailsWhenDatabaseExists(t *testing.T) {
	t.Parallel()

	cli := test_helper.NewRealDockerAPI(t)
	test_helper.EnsurePulled(t, cli, replicationExecutorImage)

	port := nat.Port("5432/tcp")
	config := &container.Config{
		Image: replicationExecutorImage,
		Env: []string{
			"POSTGRES_DB=source_db",
			"POSTGRES_USER=replication_user",
			"POSTGRES_PASSWORD=replication_password",
		},
		ExposedPorts: nat.PortSet{port: struct{}{}},
	}

	portBinding := nat.PortBinding{HostPort: ""}
	portBindings := []nat.PortBinding{portBinding}
	hostConfig := &container.HostConfig{PortBindings: nat.PortMap{port: portBindings}}
	containerName := test_helper.UniqueName(t, "replication_target")

	created, err := cli.ContainerCreate(t.Context(), config, hostConfig, nil, nil, containerName)
	require.NoError(t, err)

	t.Cleanup(func() { test_helper.RemoveContainer(t, cli, created.ID) })

	startOptions := container.StartOptions{}

	err = cli.ContainerStart(t.Context(), created.ID, startOptions)
	require.NoError(t, err)

	inspection, err := cli.ContainerInspect(t.Context(), created.ID)
	require.NoError(t, err)

	hostPort := inspection.NetworkSettings.Ports[port][0].HostPort
	require.NotEmpty(t, hostPort)

	payload := &velez_api.ReplicatePgDatabaseTaskPayload{
		SourceDbName:   "source_db",
		TargetHost:     "127.0.0.1",
		TargetPort:     mustUint32(t, hostPort),
		TargetUser:     "replication_user",
		TargetPassword: "replication_password",
	}

	job := &checkReplicationTargetJob{payload: payload}

	require.Eventually(t, func() bool {
		err := job.Do(context.Background())

		return err != nil && err.Error() == "target database already exists"
	}, 30*time.Second, 250*time.Millisecond)
}

func mustUint32(t *testing.T, value string) uint32 {
	t.Helper()

	parsed, err := strconv.ParseUint(value, 10, 32)
	require.NoError(t, err, "invalid published port %q", value)

	return uint32(parsed)
}
