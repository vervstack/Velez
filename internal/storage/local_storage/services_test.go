package local_storage

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/user_errors"
	"go.vervstack.ru/Velez/tests/test_helper"
)

// GetByName("velez") must not 500 in single-node mode: the service list
// hands out a synthetic "velez" card even when Velez runs as a bare binary
// with no container, so the detail lookup needs the same fallback.
func Test_dockerServices_GetByName_VelezResolvesWithoutContainer(t *testing.T) {
	t.Parallel()

	s := newServicesStorage(test_helper.NewRealDocker(t))

	svc, err := s.GetByName(context.Background(), velezServiceName)
	require.NoError(t, err)
	require.Equal(t, velezServiceName, svc.Name)
	require.NotEmpty(t, svc.Labels)
}

func Test_dockerServices_syntheticVelezService(t *testing.T) {
	t.Parallel()

	s := newServicesStorage(test_helper.NewRealDocker(t))

	svc, err := s.syntheticVelezService(context.Background())
	require.NoError(t, err)
	require.Equal(t, velezServiceName, svc.Name)
	require.Equal(t, pb.DeploymentStatus_RUNNING, svc.Status)
	require.Equal(t, containerStateRunning, svc.ServiceBaseInfo.Status)
	require.Len(t, svc.Labels, 1)
}

func Test_dockerServices_GetByName_UnknownNameNotFound(t *testing.T) {
	t.Parallel()

	s := newServicesStorage(test_helper.NewRealDocker(t))

	name := test_helper.UniqueName(t, "no-such-service")

	_, err := s.GetByName(context.Background(), name)
	require.ErrorIs(t, err, user_errors.ErrStorageNotFound)
}

// A CreateNewDeploy-driven deploy (enable_registry's deployRegistryJob,
// pgaas.CreatePgInstance) upserts the service and immediately looks it up
// again, before the deploy watcher has created any container for it -
// GetByName must resolve that name instead of reporting ErrNotFound.
func Test_dockerServices_GetByName_ResolvesUpsertedServiceWithoutContainer(t *testing.T) {
	t.Parallel()

	s := newServicesStorage(test_helper.NewRealDocker(t))

	name := test_helper.UniqueName(t, "pending-service")

	err := s.UpsertService(context.Background(), name)
	require.NoError(t, err)

	svc, err := s.GetByName(context.Background(), name)
	require.NoError(t, err)
	require.Equal(t, name, svc.Name)
	require.Equal(t, pb.DeploymentStatus_SCHEDULED_DEPLOYMENT, svc.Status)
}

// Test_dockerServices_GetByName_MatchesSuffixedContainerByLabel proves
// GetByName resolves a service whose real Docker container name carries a
// ContainerSuffix ("<name>_<suffix>", labelSuffixResolver.ContainerName)
// while its VervServiceLabel stays the bare virtual name - the exact
// mismatch that made a service show up in List (label-scanned) but 404 on
// GetByName (used to be a Docker container-Name filter), which broke
// registerRunnerRowJob and every other GetByName-right-after-deploy flow.
func Test_dockerServices_GetByName_MatchesSuffixedContainerByLabel(t *testing.T) {
	t.Parallel()

	cli := test_helper.NewRealDockerAPI(t)
	test_helper.EnsurePulled(t, cli, test_helper.HelloWorldAppImage)

	virtualName := test_helper.UniqueName(t, "suffixed-service")
	realDockerName := virtualName + "_dev"

	cfg := &container.Config{
		Image: test_helper.HelloWorldAppImage,
		Labels: map[string]string{
			labels.VervServiceLabel: virtualName,
		},
	}

	created, err := cli.ContainerCreate(context.Background(), cfg, nil, nil, nil, realDockerName)
	require.NoError(t, err)

	t.Cleanup(func() {
		test_helper.RemoveContainer(t, cli, created.ID)
	})

	s := newServicesStorage(test_helper.NewRealDocker(t))

	svc, err := s.GetByName(context.Background(), virtualName)
	require.NoError(t, err)
	require.Equal(t, virtualName, svc.Name)
}

func Test_dockerServices_GetByName_DeleteClearsUpsertedOverlay(t *testing.T) {
	t.Parallel()

	s := newServicesStorage(test_helper.NewRealDocker(t))

	name := test_helper.UniqueName(t, "pending-service")

	err := s.UpsertService(context.Background(), name)
	require.NoError(t, err)

	err = s.Delete(context.Background(), name)
	require.NoError(t, err)

	_, err = s.GetByName(context.Background(), name)
	require.ErrorIs(t, err, user_errors.ErrStorageNotFound)
}
