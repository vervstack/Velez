package local_storage

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/tests/test_helper"
)

func createLabelledContainer(t *testing.T, cli client.APIClient, name string, containerLabels map[string]string) {
	t.Helper()

	cfg := &container.Config{
		Image:  test_helper.HelloWorldAppImage,
		Labels: containerLabels,
	}

	created, err := cli.ContainerCreate(context.Background(), cfg, nil, nil, nil, name)
	require.NoError(t, err)

	t.Cleanup(func() {
		test_helper.RemoveContainer(t, cli, created.ID)
	})
}

func Test_dockerServiceDepsStorage_GetDependencies_SkipsSidecars(t *testing.T) {
	t.Parallel()

	cli := test_helper.NewRealDockerAPI(t)
	test_helper.EnsurePulled(t, cli, test_helper.HelloWorldAppImage)

	serviceName := test_helper.UniqueName(t, "deps-service")
	sidecarName := serviceName + "_web_ui"
	dependsOnName := serviceName + "_caller"
	targetService := test_helper.UniqueName(t, "deps-target")

	rootLabels := map[string]string{labels.VervServiceLabel: serviceName}
	createLabelledContainer(t, cli, serviceName, rootLabels)

	sidecarLabels := map[string]string{
		labels.VervServiceLabel: serviceName,
		labels.Sidecar:          "true",
	}
	createLabelledContainer(t, cli, sidecarName, sidecarLabels)

	dependantLabels := map[string]string{
		labels.VervServiceLabel: serviceName,
		labels.DependsOnLabel:   targetService,
	}
	createLabelledContainer(t, cli, dependsOnName, dependantLabels)

	s := newServiceDepsStorage(test_helper.NewRealDocker(t))

	deps, err := s.GetDependencies(context.Background(), serviceName)
	require.NoError(t, err)

	targets := make(map[string]domain.NodeType, len(deps))
	for _, dep := range deps {
		targets[dep.TargetService] = dep.NodeType
	}

	require.NotContains(t, targets, sidecarName)
	require.NotContains(t, targets, serviceName)
	require.Equal(t, domain.NodeTypeService, targets[targetService])
}
