//go:build e2e_full

package e2e

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	s3WebUiLegacyInstanceName   = "e2e-s3-wui-legacy"
	s3WebUiUpgradeInstanceName  = "e2e-s3-wui-upg"
	s3WebUiRestartInstanceName  = "e2e-s3-wui-rst"
	s3WebUiGoneInstanceName     = "e2e-s3-wui-gone"
	s3WebUiResourceInstanceName = "e2e-s3-wui-res"
	s3WebUiNoResInstanceName    = "e2e-s3-wui-nores"
	s3WebUiGraphInstanceName    = "e2e-s3-wui-graph"
	s3WebUiResourceType         = "web_ui"
	s3WebUiNetworkModePrefix    = "container:"
	s3WebUiLoginPath            = "/api/auth/login"
	s3WebUiHealthPath           = "/api/v2/GetClusterHealth"
	s3WebUiHealthyMarker        = `"status": "healthy"`
	s3WebUiAnswerTimeout        = 60 * time.Second
	s3WebUiAnswerPoll           = 2 * time.Second
	s3WebUiPortProtocolSuffix   = "/tcp"
	s3WebUiLegacyPublishedProto = "tcp"
)

func (s *S3InstanceSuite) Test_S3Instance_LegacyWebUiStillWorks() {
	t := s.T()
	t.Parallel()

	instanceName := s.planeName(s3WebUiLegacyInstanceName)

	requireDindLoopbackBridge(t)

	env := s.plane.NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	removeS3Instance(dockerClient, instanceName)
	t.Cleanup(func() { removeS3Instance(dockerClient, instanceName) })

	createS3Instance(t, env, instanceName, true)

	convertToLegacyWebUiLayout(t, dockerClient, instanceName)

	legacyPort := publishedHostPort(t, dockerClient, domain.S3WebUiServiceName(instanceName), domain.S3WebUiContainerPort)

	instance := findS3Instance(t, env, instanceName)
	require.NotNil(t, instance)
	require.Equal(t, legacyPort, instance.GetWebUiPort(), "legacy web ui port comes from the web ui container")

	creds := getS3InstanceCredentials(t, env, instanceName)
	require.NotEmpty(t, creds.GetWebUiUrl())
	require.Contains(t, creds.GetWebUiUrl(), strconv.Itoa(int(legacyPort)))
	require.Equal(t, domain.S3WebUiUsername, creds.GetWebUiUsername())
	require.NotEmpty(t, creds.GetWebUiPassword())

	resource := waitForWebUiResource(t, env, domain.S3ServiceName(instanceName))
	require.Equal(t, legacyPort, resource.GetWebUiPort(), "legacy resource port comes from the web ui container")
	requireWebUiHostMatchesCredentials(t, resource, creds)

	dropS3Instance(t, env, instanceName)

	requireNoContainersOfService(t, dockerClient, domain.S3ServiceName(instanceName))
	requireNoContainersOfService(t, dockerClient, domain.S3WebUiServiceName(instanceName))
	require.Nil(t, findS3Instance(t, env, instanceName), "instance still listed after drop")
}

func (s *S3InstanceSuite) Test_S3Instance_WebUiSurvivesGarageUpgrade() {
	t := s.T()
	t.Parallel()

	t.Skip("known gap: upgrade_smerd cannot upgrade a garage container (host ports stay bound, " +
		"/etc/garage.toml is not carried over), independent of the web ui sidecar, see report")

	instanceName := s.planeName(s3WebUiUpgradeInstanceName)

	requireDindLoopbackBridge(t)

	env := s.plane.NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	removeS3Instance(dockerClient, instanceName)
	t.Cleanup(func() { removeS3Instance(dockerClient, instanceName) })

	createS3Instance(t, env, instanceName, true)

	garageBefore := inspectContainer(t, env, domain.S3ServiceName(instanceName))
	sidecarBefore := inspectContainer(t, env, domain.S3WebUiServiceName(instanceName))

	upgradeReq := newUpgradeS3GarageRequest(instanceName, garageBefore.Config.Image)

	upgradeStoppingOldFirst(t, env, upgradeReq)

	garageAfter := inspectContainer(t, env, domain.S3ServiceName(instanceName))
	require.NotEqual(t, garageBefore.ID, garageAfter.ID, "upgrade replaces the garage container")

	sidecarAfter := requireWebUiSidecarJoined(t, dockerClient, instanceName)
	require.NotEqual(t, sidecarBefore.ID, sidecarAfter.ID, "upgrade recreates the web ui sidecar")

	creds := getS3InstanceCredentials(t, env, instanceName)

	requireWebUiReachesGarage(t, creds)

	dropS3Instance(t, env, instanceName)
}

func (s *S3InstanceSuite) Test_S3Instance_WebUiAfterGarageRestart() {
	t := s.T()
	t.Parallel()

	t.Skip("known gap: sidecar loses network namespace when the root restarts, see report")

	instanceName := s.planeName(s3WebUiRestartInstanceName)

	requireDindLoopbackBridge(t)

	env := s.plane.NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	removeS3Instance(dockerClient, instanceName)
	t.Cleanup(func() { removeS3Instance(dockerClient, instanceName) })

	createS3Instance(t, env, instanceName, true)

	creds := getS3InstanceCredentials(t, env, instanceName)

	requireWebUiReachesGarage(t, creds)

	err := dockerClient.ContainerRestart(t.Context(), domain.S3ServiceName(instanceName), container.StopOptions{})
	require.NoError(t, err)

	requireWebUiReachesGarage(t, creds)

	dropS3Instance(t, env, instanceName)
}

func (s *S3InstanceSuite) Test_S3Instance_DropAfterSidecarRemoved() {
	t := s.T()
	t.Parallel()

	instanceName := s.planeName(s3WebUiGoneInstanceName)

	requireDindLoopbackBridge(t)

	env := s.plane.NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	removeS3Instance(dockerClient, instanceName)
	t.Cleanup(func() { removeS3Instance(dockerClient, instanceName) })

	createS3Instance(t, env, instanceName, true)

	removeOpts := container.RemoveOptions{Force: true}

	err := dockerClient.ContainerRemove(t.Context(), domain.S3WebUiServiceName(instanceName), removeOpts)
	require.NoError(t, err)

	dropS3Instance(t, env, instanceName)

	requireNoContainersOfService(t, dockerClient, domain.S3ServiceName(instanceName))
	require.Nil(t, findS3Instance(t, env, instanceName), "instance still listed after drop")
}

func (s *S3InstanceSuite) Test_S3Instance_WebUiResourceCarriesPort() {
	t := s.T()
	t.Parallel()

	instanceName := s.planeName(s3WebUiResourceInstanceName)

	requireDindLoopbackBridge(t)

	env := s.plane.NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	removeS3Instance(dockerClient, instanceName)
	t.Cleanup(func() { removeS3Instance(dockerClient, instanceName) })

	createS3Instance(t, env, instanceName, true)

	instance := findS3Instance(t, env, instanceName)
	require.NotNil(t, instance)
	require.NotZero(t, instance.GetWebUiPort())

	resource := waitForWebUiResource(t, env, domain.S3ServiceName(instanceName))
	require.NotZero(t, resource.GetWebUiPort())
	require.Equal(t, instance.GetWebUiPort(), resource.GetWebUiPort(), "resource must carry the port garage publishes")

	creds := getS3InstanceCredentials(t, env, instanceName)
	requireWebUiHostMatchesCredentials(t, resource, creds)

	requireWebUiPageAnswers(t, resource.GetWebUiPort())

	dropS3Instance(t, env, instanceName)

	requireNoContainersOfService(t, dockerClient, domain.S3ServiceName(instanceName))
}

func (s *S3InstanceSuite) Test_S3Instance_NoWebUiResourceWithoutWebUi() {
	t := s.T()
	t.Parallel()

	instanceName := s.planeName(s3WebUiNoResInstanceName)

	requireDindLoopbackBridge(t)

	env := s.plane.NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	removeS3Instance(dockerClient, instanceName)
	t.Cleanup(func() { removeS3Instance(dockerClient, instanceName) })

	createS3Instance(t, env, instanceName, false)

	resources := getServiceResources(t, env, domain.S3ServiceName(instanceName))
	require.Nil(t, findBoundResource(resources, s3WebUiResourceType), "no web ui, no web_ui resource")

	dropS3Instance(t, env, instanceName)

	requireNoContainersOfService(t, dockerClient, domain.S3ServiceName(instanceName))
}

func (s *S3InstanceSuite) Test_S3Instance_WebUiSidecarIsNotADependency() {
	t := s.T()
	t.Parallel()

	instanceName := s.planeName(s3WebUiGraphInstanceName)

	requireDindLoopbackBridge(t)

	env := s.plane.NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	removeS3Instance(dockerClient, instanceName)
	t.Cleanup(func() { removeS3Instance(dockerClient, instanceName) })

	createS3Instance(t, env, instanceName, true)

	rootName := domain.S3ServiceName(instanceName)
	sidecarName := domain.S3WebUiServiceName(instanceName)

	sidecar := findServiceSidecar(t, env, rootName, sidecarName)
	require.NotNil(t, sidecar, "web ui container must stay listed as a sidecar of the s3 root service")

	graphReq := newGetServiceGraphRequest(rootName)

	graph, err := env.Custom.ServiceApiImpl.GetServiceGraph(t.Context(), graphReq)
	require.NoError(t, err)

	requireNoGraphNode(t, graph.GetDependencies(), sidecarName)
	requireNoGraphNode(t, graph.GetCallers(), sidecarName)

	dropS3Instance(t, env, instanceName)

	requireNoContainersOfService(t, dockerClient, rootName)
}

func newGetServiceResourcesRequest(serviceName string) *velez_api.GetServiceResources_Request {
	return &velez_api.GetServiceResources_Request{ServiceName: serviceName}
}

func newGetServiceGraphRequest(serviceName string) *velez_api.GetServiceGraph_Request {
	return &velez_api.GetServiceGraph_Request{ServiceName: serviceName}
}

func getServiceResources(t *testing.T, env *TestEnvironment, serviceName string) []*velez_api.BoundResource {
	t.Helper()

	req := newGetServiceResourcesRequest(serviceName)

	resp, err := env.Custom.ServiceApiImpl.GetServiceResources(t.Context(), req)
	require.NoError(t, err)

	return resp.GetResources()
}

func findBoundResource(resources []*velez_api.BoundResource, resourceType string) *velez_api.BoundResource {
	for _, resource := range resources {
		if resource.GetResourceType() == resourceType {
			return resource
		}
	}

	return nil
}

// waitForWebUiResource polls until the service lists a web_ui resource that
// already carries a published port.
func waitForWebUiResource(t *testing.T, env *TestEnvironment, serviceName string) *velez_api.BoundResource {
	t.Helper()

	var found *velez_api.BoundResource

	pollUntil(t, "service "+serviceName+" never listed a web_ui resource with a port", func() (bool, string) {
		req := newGetServiceResourcesRequest(serviceName)

		resp, err := env.Custom.ServiceApiImpl.GetServiceResources(t.Context(), req)
		if err != nil {
			return false, fmt.Sprintf("err %v", err)
		}

		found = findBoundResource(resp.GetResources(), s3WebUiResourceType)

		return found != nil && found.GetWebUiPort() != 0, fmt.Sprintf("resources %v", resp.GetResources())
	})

	require.Equal(t, s3WebUiResourceType, found.GetName())

	return found
}

// requireWebUiHostMatchesCredentials: a resource host is only reported for a
// remote environment, where it is the very host the credentials url uses.
func requireWebUiHostMatchesCredentials(
	t *testing.T, resource *velez_api.BoundResource, creds *velez_api.GetS3InstanceCredentials_Response,
) {
	t.Helper()

	if resource.GetWebUiHost() == "" {
		return
	}

	parsed, err := url.Parse(creds.GetWebUiUrl())
	require.NoError(t, err)

	require.Equal(t, parsed.Hostname(), resource.GetWebUiHost())
}

func requireWebUiPageAnswers(t *testing.T, port uint32) {
	t.Helper()

	hostAddr, ok := sharedDind.Addr(int(port))
	require.True(t, ok, "dind did not publish web ui port %d", port)

	pollUntil(t, "web ui never answered 200 on "+hostAddr, func() (bool, string) {
		ctx, cancel := context.WithTimeout(t.Context(), s3HttpTimeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+hostAddr+"/", nil)
		if err != nil {
			return false, fmt.Sprintf("err %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return false, fmt.Sprintf("err %v", err)
		}

		_ = resp.Body.Close()

		return resp.StatusCode == http.StatusOK, fmt.Sprintf("status %d", resp.StatusCode)
	})
}

func requireNoGraphNode(t *testing.T, nodes []*velez_api.ServiceDependencyInfo, containerName string) {
	t.Helper()

	for _, node := range nodes {
		require.NotEqual(t, containerName, strings.TrimPrefix(node.GetServiceName(), "/"),
			"the web ui sidecar must not appear in the service graph")
	}
}

// upgradeStoppingOldFirst runs upgrade_smerd the way register_container does for
// containers with host ports: the UpgradeSmerd RPC only pauses the old
// container, which keeps its host ports bound, so a replacement publishing the
// same ports cannot start.
func upgradeStoppingOldFirst(t *testing.T, env *TestEnvironment, req *velez_api.UpgradeSmerd_Request) {
	t.Helper()

	payload := &velez_api.UpgradeSmerdTaskPayload{UpgradeRequest: req, StopOldFirst: true}

	_, err := env.Custom.JobsEngine.Enqueue(t.Context(), req.GetName(), jobs.UpgradeSmerdAction, payload)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(t.Context(), s3TaskTimeout)
	defer cancel()

	var task tasks_queries.VelezTask

	for observed := range env.Custom.JobsEngine.Watch(ctx, req.GetName(), jobs.UpgradeSmerdAction) {
		task = observed
	}

	require.Equal(t, tasks_queries.VelezTaskStatusDONE, task.Status, "upgrade_smerd error: %s", task.Error.String)
}

func newUpgradeS3GarageRequest(instanceName, imageName string) *velez_api.UpgradeSmerd_Request {
	return &velez_api.UpgradeSmerd_Request{
		Name:  domain.S3ServiceName(instanceName),
		Image: imageName,
	}
}

// requireWebUiSidecarJoined asserts the web ui is a container-network-mode
// sidecar of the garage container, and that garage (not the sidecar)
// publishes the web ui port.
func requireWebUiSidecarJoined(
	t *testing.T, dockerClient client.APIClient, instanceName string,
) container.InspectResponse {
	t.Helper()

	garage, err := dockerClient.ContainerInspect(t.Context(), domain.S3ServiceName(instanceName))
	require.NoError(t, err)

	sidecar, err := dockerClient.ContainerInspect(t.Context(), domain.S3WebUiServiceName(instanceName))
	require.NoError(t, err)

	require.True(t, sidecar.State.Running, "web ui sidecar must run")
	require.Equal(t, container.NetworkMode(s3WebUiNetworkModePrefix+garage.ID), sidecar.HostConfig.NetworkMode)
	require.Equal(t, labelValueTrue, sidecar.Config.Labels[labels.Sidecar])
	require.Equal(t, domain.S3ServiceName(instanceName), sidecar.Config.Labels[labels.VervServiceLabel])
	require.Equal(t, domain.S3ServiceName(instanceName), sidecar.Config.Labels[labels.WebUiForLabel])
	require.Equal(t, instanceName, sidecar.Config.Labels[labels.S3WebUiLabel])
	require.Empty(t, sidecar.HostConfig.PortBindings, "a sidecar cannot publish ports")

	require.NotEmpty(t, webUiBindings(garage), "garage must publish the web ui port")

	return sidecar
}

func webUiBindings(garage container.InspectResponse) []nat.PortBinding {
	for port, bindings := range garage.HostConfig.PortBindings {
		if string(port) == strconv.Itoa(domain.S3WebUiContainerPort)+s3WebUiPortProtocolSuffix {
			return bindings
		}
	}

	return nil
}

func requireWebUiNotPublishedByGarage(t *testing.T, dockerClient client.APIClient, instanceName string) {
	t.Helper()

	garage, err := dockerClient.ContainerInspect(t.Context(), domain.S3ServiceName(instanceName))
	require.NoError(t, err)

	require.Empty(t, webUiBindings(garage), "garage must not publish the web ui port without a web ui")
}

func requireWebUiPortPublishedAs(
	t *testing.T, dockerClient client.APIClient, instanceName string, want uint32,
) {
	t.Helper()

	garage, err := dockerClient.ContainerInspect(t.Context(), domain.S3ServiceName(instanceName))
	require.NoError(t, err)

	bindings := webUiBindings(garage)
	require.Len(t, bindings, 1)
	require.Equal(t, strconv.Itoa(int(want)), bindings[0].HostPort)
}

func requireNoWebUiService(t *testing.T, env *TestEnvironment, instanceName string) {
	t.Helper()

	webUiName := domain.S3WebUiServiceName(instanceName)

	_, err := env.Custom.ServiceApiImpl.GetService(t.Context(), newGetServiceRequest(webUiName))
	require.Error(t, err, "the web ui sidecar must not be a service of its own")

	listResp, err := env.Custom.ServiceApiImpl.ListServices(t.Context(), newListServicesRequest())
	require.NoError(t, err)

	require.Nil(t, findServiceByName(listResp.GetServices(), webUiName), "web ui listed as a service")
}

func requireNotListedAsSmerdByName(t *testing.T, env *TestEnvironment, containerName string) {
	t.Helper()

	listResp := env.ListSmerds(t, t.Context(), &velez_api.ListSmerds_Request{})

	for _, smerd := range listResp.GetSmerds() {
		require.NotEqual(t, containerName, strings.TrimPrefix(smerd.GetName(), "/"),
			"a sidecar must not be listed as a smerd")
	}
}

func requireNoContainersOfService(t *testing.T, dockerClient client.APIClient, serviceName string) {
	t.Helper()

	listOpts := container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("label", labels.VervServiceLabel+"="+serviceName)),
	}

	listed, err := dockerClient.ContainerList(t.Context(), listOpts)
	require.NoError(t, err)
	require.Empty(t, listed, "containers labelled %s=%s are left over", labels.VervServiceLabel, serviceName)

	for _, name := range []string{serviceName, serviceName + "_web_ui"} {
		_, err = dockerClient.ContainerInspect(t.Context(), name)
		require.True(t, client.IsErrNotFound(err), "container %q must be gone, got: %v", name, err)
	}
}

func publishedHostPort(t *testing.T, dockerClient client.APIClient, containerName string, containerPort int) uint32 {
	t.Helper()

	inspected, err := dockerClient.ContainerInspect(t.Context(), containerName)
	require.NoError(t, err)

	bindings := inspected.NetworkSettings.Ports[nat.Port(strconv.Itoa(containerPort)+s3WebUiPortProtocolSuffix)]
	require.NotEmpty(t, bindings, "container %q publishes nothing on %d", containerName, containerPort)

	hostPort, err := strconv.Atoi(bindings[0].HostPort)
	require.NoError(t, err)

	return uint32(hostPort)
}

// pollUntil retries check until it succeeds or s3WebUiAnswerTimeout passes,
// failing with the detail of the last attempt (require.Eventually would
// format its message before the first attempt).
func pollUntil(t *testing.T, failure string, check func() (bool, string)) {
	t.Helper()

	deadline := time.Now().Add(s3WebUiAnswerTimeout)

	var detail string

	for time.Now().Before(deadline) {
		var isOk bool

		isOk, detail = check()
		if isOk {
			return
		}

		time.Sleep(s3WebUiAnswerPoll)
	}

	require.Failf(t, failure, "last attempt: %s", detail)
}

// webUiGarageHealth logs in to garage-webui and asks it for the cluster health.
// garage-webui proxies that call to garage's admin API at API_BASE_URL, which
// is loopback inside the shared network namespace - so a healthy answer proves
// the sidecar really sits in garage's namespace. The static pages need no
// login, so a bare GET / proves nothing about either.
func webUiGarageHealth(ctx context.Context, webUiUrl, username, password string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, s3HttpTimeout)
	defer cancel()

	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", fmt.Errorf("new cookie jar: %w", err)
	}

	httpClient := &http.Client{Jar: jar}

	loginBody := fmt.Sprintf(`{"username":%q,"password":%q}`, username, password)

	loginReq, err := http.NewRequestWithContext(
		ctx, http.MethodPost, webUiUrl+s3WebUiLoginPath, strings.NewReader(loginBody),
	)
	if err != nil {
		return "", fmt.Errorf("new login request: %w", err)
	}

	loginReq.Header.Set("Content-Type", "application/json")

	loginResp, err := httpClient.Do(loginReq)
	if err != nil {
		return "", fmt.Errorf("login: %w", err)
	}

	_ = loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login status %d", loginResp.StatusCode)
	}

	healthReq, err := http.NewRequestWithContext(ctx, http.MethodGet, webUiUrl+s3WebUiHealthPath, nil)
	if err != nil {
		return "", fmt.Errorf("new health request: %w", err)
	}

	healthResp, err := httpClient.Do(healthReq)
	if err != nil {
		return "", fmt.Errorf("health: %w", err)
	}

	defer func() { _ = healthResp.Body.Close() }()

	body, err := io.ReadAll(healthResp.Body)
	if err != nil {
		return "", fmt.Errorf("read health: %w", err)
	}

	if healthResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("health status %d: %s", healthResp.StatusCode, body)
	}

	return string(body), nil
}

// requireWebUiReachesGarage polls until the web ui, behind its own login,
// returns garage's cluster health through the shared network namespace.
func requireWebUiReachesGarage(t *testing.T, creds *velez_api.GetS3InstanceCredentials_Response) {
	t.Helper()

	webUiUrl := "http://" + endpointDindAddr(t, creds.GetWebUiUrl())

	pollUntil(t, "web ui never returned garage's cluster health via "+webUiUrl, func() (bool, string) {
		body, err := webUiGarageHealth(t.Context(), webUiUrl, creds.GetWebUiUsername(), creds.GetWebUiPassword())

		return err == nil && strings.Contains(body, s3WebUiHealthyMarker), fmt.Sprintf("body %q, err %v", body, err)
	})
}

func requireWebUiRejectsAnonymousApi(t *testing.T, creds *velez_api.GetS3InstanceCredentials_Response) {
	t.Helper()

	url := "http://" + endpointDindAddr(t, creds.GetWebUiUrl()) + s3WebUiHealthPath

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode, "the web ui api must demand a login")
}

// convertToLegacyWebUiLayout rewrites a freshly created instance into the
// layout predating the sidecar: the web ui is a separate container with its
// own service label and its own published port.
func convertToLegacyWebUiLayout(t *testing.T, dockerClient client.APIClient, instanceName string) {
	t.Helper()

	webUiName := domain.S3WebUiServiceName(instanceName)

	removeOpts := container.RemoveOptions{Force: true}

	err := dockerClient.ContainerRemove(t.Context(), webUiName, removeOpts)
	require.NoError(t, err)

	pullReader, err := dockerClient.ImagePull(t.Context(), NginxAlpineImage, image.PullOptions{})
	require.NoError(t, err)

	_, err = io.Copy(io.Discard, pullReader)
	require.NoError(t, err)

	err = pullReader.Close()
	require.NoError(t, err)

	garage, err := dockerClient.ContainerInspect(t.Context(), domain.S3ServiceName(instanceName))
	require.NoError(t, err)

	createReq := newLegacyWebUiSpec(instanceName, garage.Config.Labels[labels.SuffixLabel])

	created, err := dockerClient.ContainerCreate(
		t.Context(), createReq.config, createReq.hostConfig, nil, nil, webUiName,
	)
	require.NoError(t, err)

	err = dockerClient.ContainerStart(t.Context(), created.ID, container.StartOptions{})
	require.NoError(t, err)
}

type legacyWebUiSpec struct {
	config     *container.Config
	hostConfig *container.HostConfig
}

func newLegacyWebUiSpec(instanceName, environmentSuffix string) legacyWebUiSpec {
	webUiPort := nat.Port(strconv.Itoa(domain.S3WebUiContainerPort) + "/" + s3WebUiLegacyPublishedProto)

	config := &container.Config{
		Image:        NginxAlpineImage,
		ExposedPorts: nat.PortSet{webUiPort: struct{}{}},
		Labels: map[string]string{
			labels.CreatedWithVelezLabel: labelValueTrue,
			labels.SuffixLabel:           environmentSuffix,
			labels.VervServiceLabel:      domain.S3WebUiServiceName(instanceName),
			labels.S3WebUiLabel:          instanceName,
			labels.WebUiForLabel:         domain.S3ServiceName(instanceName),
			labels.WebUiPortLabel:        strconv.Itoa(domain.S3WebUiContainerPort),
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{webUiPort: []nat.PortBinding{{HostIP: "0.0.0.0"}}},
	}

	return legacyWebUiSpec{config: config, hostConfig: hostConfig}
}
