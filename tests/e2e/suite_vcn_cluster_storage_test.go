//go:build e2e_full

package e2e

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"github.com/sqlc-dev/pqtype"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage/postgres"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/tests/test_helper"
)

const (
	vcnStorageNodeA    = "vcn_node_a"
	vcnStorageNodeB    = "vcn_node_b"
	vcnStoragePassword = "vcn-storage-test"
	vcnStorageProvider = "headscale"
)

func Test_VcnClusterStorage(t *testing.T) {
	t.Parallel()

	address := startVcnClusterStoragePostgres(t)
	rootDsn := vcnStorageDsn(address, "postgres")
	root := openPostgresViaDind(t, rootDsn, dindDialIp)
	require.Eventually(t, func() bool {
		return root.PingContext(t.Context()) == nil
	}, 30*time.Second, 250*time.Millisecond)

	migrations := os.DirFS(filepath.Join(repoRoot(t), "migrations"))
	migrator, err := goose.NewProvider(goose.DialectPostgres, root, migrations)
	require.NoError(t, err)

	_, err = migrator.Up(t.Context())
	require.NoError(t, err)

	createVcnStorageNodes(t, root)

	nodeA := openVcnStorageNode(t, address, vcnStorageNodeA)
	nodeB := openVcnStorageNode(t, address, vcnStorageNodeB)

	t.Run("persists network across nodes", func(t *testing.T) {
		t.Parallel()

		verifyVcnNetworkPersistence(t, nodeA, nodeB)
	})

	t.Run("routes claims to their node", func(t *testing.T) {
		t.Parallel()

		verifyVcnTaskClaims(t, nodeA, nodeB)
	})
}

func startVcnClusterStoragePostgres(t *testing.T) string {
	t.Helper()

	const pgPort nat.Port = "5432/tcp"

	port := acquireClusterPgPort(t)
	env := TestEnvironment{t: t}
	dockerClient := env.DockerClient()
	config := &container.Config{
		Image:        "postgres:18",
		Env:          []string{"POSTGRES_PASSWORD=" + vcnStoragePassword},
		ExposedPorts: nat.PortSet{pgPort: struct{}{}},
	}
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{pgPort: {{HostPort: strconv.Itoa(port)}}},
	}
	name := test_helper.UniqueName(t, "postgres")
	created, err := dockerClient.ContainerCreate(t.Context(), config, hostConfig, nil, nil, name)
	require.NoError(t, err)

	t.Cleanup(func() {
		removeOpts := container.RemoveOptions{Force: true, RemoveVolumes: true}
		removeErr := dockerClient.ContainerRemove(context.Background(), created.ID, removeOpts)
		require.NoError(t, removeErr)
	})

	startOpts := container.StartOptions{}

	err = dockerClient.ContainerStart(t.Context(), created.ID, startOpts)
	require.NoError(t, err)

	return dindGatewayAddr(t, port)
}

func vcnStorageDsn(address, username string) string {
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(username, vcnStoragePassword),
		Host:     address,
		Path:     "/postgres",
		RawQuery: "sslmode=disable&connect_timeout=2",
	}

	return dsn.String()
}

func openVcnStorageNode(t *testing.T, address, name string) *postgres.Storage {
	t.Helper()

	dsn := vcnStorageDsn(address, name)
	db := openPostgresViaDind(t, dsn, dindDialIp)

	return postgres.New(db, dsn)
}

func createVcnStorageNodes(t *testing.T, root *sql.DB) {
	t.Helper()

	const roles = `
		CREATE ROLE vcn_node_a LOGIN IN ROLE working_node PASSWORD 'vcn-storage-test';
		CREATE ROLE vcn_node_b LOGIN IN ROLE working_node PASSWORD 'vcn-storage-test';
	`

	_, err := root.ExecContext(t.Context(), roles)
	require.NoError(t, err)
}

func verifyVcnNetworkPersistence(t *testing.T, nodeA, nodeB *postgres.Storage) {
	t.Helper()

	ctx := t.Context()
	nameA, err := nodeA.Networks().CurrentNode(ctx)
	require.NoError(t, err)
	require.Equal(t, vcnStorageNodeA, nameA)

	nameB, err := nodeB.Networks().CurrentNode(ctx)
	require.NoError(t, err)
	require.Equal(t, vcnStorageNodeB, nameB)

	expected := newVcnStorageNetwork(nameA)

	err = nodeA.Networks().Create(ctx, expected)
	require.NoError(t, err)

	got, err := nodeB.Networks().Get(ctx, expected.Id)
	require.NoError(t, err)

	got.CreatedAt = got.CreatedAt.UTC()
	require.Equal(t, expected, got)

	networks, err := nodeB.Networks().List(ctx)
	require.NoError(t, err)
	require.Len(t, networks, 1)
	require.Equal(t, expected.Id, networks[0].Id)

	err = nodeB.Networks().Drop(ctx, expected.Id)
	require.NoError(t, err)

	_, err = nodeA.Networks().Get(ctx, expected.Id)
	require.ErrorIs(t, err, user_errors.ErrStorageNotFound)
}

func newVcnStorageNetwork(nodeName string) domain.Network {
	return domain.Network{
		Id:          uuid.NewString(),
		Name:        "vcn-repo-network",
		Provider:    vcnStorageProvider,
		Source:      "deployed",
		Url:         "https://headscale.example",
		Token:       "provider-admin-token",
		ServiceName: "vcn-repo-service",
		ContainerId: "vcn-repo-container",
		NodeName:    nodeName,
		CreatedAt:   time.Now().UTC().Truncate(time.Microsecond),
	}
}

func verifyVcnTaskClaims(t *testing.T, nodeA, nodeB *postgres.Storage) {
	t.Helper()

	ctx := t.Context()
	targeted := createVcnStorageTask(t, nodeA, "targeted", `{"node_name":"vcn_node_a"}`)
	untargeted := createVcnStorageTask(t, nodeA, "untargeted", `{}`)

	claimB := newVcnStorageClaim("worker-b", time.Time{})
	claimed, err := nodeB.Tasks().ClaimTask(ctx, claimB)
	require.NoError(t, err)
	require.Equal(t, untargeted.ID, claimed.ID)

	_, err = nodeB.Tasks().ClaimTask(ctx, claimB)
	require.ErrorIs(t, err, sql.ErrNoRows)

	claimA := newVcnStorageClaim("worker-a", time.Time{})

	claimed, err = nodeA.Tasks().ClaimTask(ctx, claimA)
	require.NoError(t, err)
	require.Equal(t, targeted.ID, claimed.ID)

	complete := tasks_queries.FinishTaskParams{ID: untargeted.ID, Status: tasks_queries.VelezTaskStatusDONE}

	err = nodeB.Tasks().FinishTask(ctx, complete)
	require.NoError(t, err)

	staleBefore := time.Now().Add(time.Minute)
	staleClaimB := newVcnStorageClaim("worker-b", staleBefore)

	_, err = nodeB.Tasks().ClaimTask(ctx, staleClaimB)
	require.ErrorIs(t, err, sql.ErrNoRows)

	staleClaimA := newVcnStorageClaim("worker-a", staleBefore)

	claimed, err = nodeA.Tasks().ClaimTask(ctx, staleClaimA)
	require.NoError(t, err)
	require.Equal(t, targeted.ID, claimed.ID)
}

func createVcnStorageTask(t *testing.T, node *postgres.Storage, entity, payload string) tasks_queries.VelezTask {
	t.Helper()

	params := tasks_queries.CreateTaskParams{
		EntityID: entity,
		Action:   "connect_service_to_vpn",
		Context:  pqtype.NullRawMessage{RawMessage: []byte(payload), Valid: true},
	}
	task, err := node.Tasks().CreateTask(t.Context(), params)
	require.NoError(t, err)

	return task
}

func newVcnStorageClaim(worker string, staleBefore time.Time) tasks_queries.ClaimTaskParams {
	return tasks_queries.ClaimTaskParams{
		ClaimedBy: sql.NullString{String: worker, Valid: true},
		ClaimedAt: sql.NullTime{Time: staleBefore, Valid: !staleBefore.IsZero()},
	}
}
