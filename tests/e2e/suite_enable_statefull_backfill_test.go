//go:build e2e_full

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage/container_derived"
)

const (
	// The Postgres environments seed keeps the default environment unsuffixed, so
	// only containers of an unsuffixed node are visible to the backfill.
	backfillSuffix = ""

	backfillPlainService   = "e2e_backfill_plain"
	backfillPgName         = "e2e_backfill_pg"
	backfillPgService      = "pgaas_" + backfillPgName
	backfillSidecarService = "e2e_backfill_sidecar"
	backfillSidecarName    = "e2e_backfill_sidecar_c"

	backfillPgTaskTimeout = 3 * time.Minute
	backfillPollEvery     = 2 * time.Second
)

type backfillContainer struct {
	id        string
	startedAt string
}

func Test_EnableStatefull_BackfillsLabeledContainers(t *testing.T) {
	t.Parallel()

	env, _ := newStatefullEnvironment(t, Planes[1], backfillSuffix)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	serviceNames := []string{backfillPlainService, backfillPgService, backfillSidecarService}

	removeBackfillContainers(dockerClient, serviceNames)
	t.Cleanup(func() { removeBackfillContainers(dockerClient, serviceNames) })

	plainReq := newBackfillPlainRequest()
	plain := env.CreateSmerd(t, plainReq)
	require.Equal(t, velez_api.Smerd_running, plain.GetStatus())

	pgInstance, pgPassword := createBackfillPgInstance(t, env)

	startBackfillSidecar(t, env)

	before := map[string]backfillContainer{
		backfillPlainService:   inspectBackfillContainer(t, dockerClient, backfillPlainService),
		backfillPgService:      inspectBackfillContainer(t, dockerClient, backfillPgService),
		backfillSidecarService: inspectBackfillContainer(t, dockerClient, backfillSidecarService),
	}

	enableStatefullPg(t, env)

	requireServiceListed(t, env, backfillPlainService)
	requireServiceListed(t, env, backfillPgService)
	requireBackfilledPg(t, env, pgInstance, pgPassword)
	requireSidecarNotBackfilled(t, env)
	requireBackfillUntouched(t, dockerClient, before)

	time.Sleep(registerDeployWatcherTicks)

	requireDeploymentRunning(t, env, backfillPlainService)
	requireDeploymentRunning(t, env, backfillPgService)
	requireBackfillUntouched(t, dockerClient, before)
}

func newBackfillPlainRequest() *velez_api.CreateSmerd_Request {
	return &velez_api.CreateSmerd_Request{
		Name:         backfillPlainService,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Labels:       map[string]string{labels.VervServiceLabel: backfillPlainService},
	}
}

func createBackfillPgInstance(t *testing.T, env *TestEnvironment) (*velez_api.PgInstance, string) {
	t.Helper()

	createReq := &velez_api.CreatePgInstance_Request{Name: backfillPgName}

	_, err := env.Custom.PgaasApiImpl.CreatePgInstance(t.Context(), createReq)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		listed := findPgInstance(t, env, backfillPgService)

		return listed != nil && listed.GetStatus() == "running"
	}, backfillPgTaskTimeout, backfillPollEvery, "the pg instance must reach running")

	instance := findPgInstance(t, env, backfillPgService)
	require.NotNil(t, instance)

	credsReq := &velez_api.GetPgInstanceCredentials_Request{Name: backfillPgService}

	creds, err := env.Custom.PgaasApiImpl.GetPgInstanceCredentials(t.Context(), credsReq)
	require.NoError(t, err)
	require.NotEmpty(t, creds.GetPassword())

	return instance, creds.GetPassword()
}

func startBackfillSidecar(t *testing.T, env *TestEnvironment) {
	t.Helper()

	dockerClient := env.Custom.NodeClients.Docker().Client()

	containerLabels := map[string]string{
		labels.VervServiceLabel: backfillSidecarService,
		labels.Sidecar:          "true",
	}
	addTestLabels(t, containerLabels)

	cfg := &container.Config{Image: HelloWorldAppImage, Labels: containerLabels}

	created, err := dockerClient.ContainerCreate(t.Context(), cfg, nil, nil, nil, backfillSidecarName)
	require.NoError(t, err)

	err = dockerClient.ContainerStart(t.Context(), created.ID, container.StartOptions{})
	require.NoError(t, err)
}

func inspectBackfillContainer(t *testing.T, dockerClient client.APIClient, serviceName string) backfillContainer {
	t.Helper()

	listOpts := container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("label", labels.VervServiceLabel+"="+serviceName)),
	}

	listed, err := dockerClient.ContainerList(t.Context(), listOpts)
	require.NoError(t, err)
	require.Len(t, listed, 1, "expected exactly one container for service %q", serviceName)

	inspected, err := dockerClient.ContainerInspect(t.Context(), listed[0].ID)
	require.NoError(t, err)
	require.True(t, inspected.State.Running, "container of service %q must be running", serviceName)

	return backfillContainer{id: inspected.ID, startedAt: inspected.State.StartedAt}
}

func requireBackfilledPg(t *testing.T, env *TestEnvironment, want *velez_api.PgInstance, wantPassword string) {
	t.Helper()

	listResp, err := env.Custom.PgaasApiImpl.ListPgInstances(t.Context(), &velez_api.ListPgInstances_Request{})
	require.NoError(t, err)

	var matching []*velez_api.PgInstance

	for _, instance := range listResp.GetInstances() {
		if instance.GetName() == backfillPgService {
			matching = append(matching, instance)
		}
	}

	require.Len(t, matching, 1, "the backfilled pg instance must be listed exactly once")
	require.Equal(t, want.GetUsername(), matching[0].GetUsername())
	require.Equal(t, want.GetDbName(), matching[0].GetDbName())

	secretRef := container_derived.PgInstanceSecretRef(backfillPgService)

	secret, err := env.Custom.Services.Secrets().Get(t.Context(), secretRef)
	require.NoError(t, err)
	require.Equal(t, wantPassword, secret)

	credsReq := &velez_api.GetPgInstanceCredentials_Request{Name: backfillPgService}

	creds, err := env.Custom.PgaasApiImpl.GetPgInstanceCredentials(t.Context(), credsReq)
	require.NoError(t, err)
	require.Equal(t, wantPassword, creds.GetPassword())
	require.Equal(t, want.GetUsername(), creds.GetUsername())
	require.Equal(t, want.GetDbName(), creds.GetDbName())

	requireDeploymentRunning(t, env, backfillPlainService)
	requireDeploymentRunning(t, env, backfillPgService)
}

func requireSidecarNotBackfilled(t *testing.T, env *TestEnvironment) {
	t.Helper()

	listResp, err := env.ServiceApiClient().ListServices(t.Context(), newListServicesRequest())
	require.NoError(t, err)

	for _, svc := range listResp.GetServices() {
		require.NotEqual(t, backfillSidecarService, svc.GetName(), "a sidecar container must not become a service")
	}

	deploymentsReq := &velez_api.ListDeployments_Request{ServiceName: toolbox.ToPtr(backfillSidecarService)}

	deploymentsResp, err := env.Custom.ServiceApiImpl.ListDeployments(t.Context(), deploymentsReq)
	require.NoError(t, err)
	require.Empty(t, deploymentsResp.GetDeployments(), "a sidecar container must not get a deployment")
}

func requireBackfillUntouched(t *testing.T, dockerClient client.APIClient, before map[string]backfillContainer) {
	t.Helper()

	for serviceName, want := range before {
		got := inspectBackfillContainer(t, dockerClient, serviceName)
		require.Equal(t, want, got, "the backfill must not restart or recreate %q", serviceName)
	}
}

func removeBackfillContainers(dockerClient client.APIClient, serviceNames []string) {
	ctx := context.Background()
	removeOpts := container.RemoveOptions{Force: true, RemoveVolumes: true}

	for _, serviceName := range serviceNames {
		listOpts := container.ListOptions{
			All:     true,
			Filters: filters.NewArgs(filters.Arg("label", labels.VervServiceLabel+"="+serviceName)),
		}

		listed, err := dockerClient.ContainerList(ctx, listOpts)
		if err != nil {
			continue
		}

		for _, listedContainer := range listed {
			_ = dockerClient.ContainerRemove(ctx, listedContainer.ID, removeOpts)
		}

		_ = dockerClient.VolumeRemove(ctx, serviceName+"-data", true)
	}
}
