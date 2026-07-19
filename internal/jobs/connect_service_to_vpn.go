package jobs

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/errdefs"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/makosh/pkg/makosh_be"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/headscale"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/patterns"
	"go.vervstack.ru/Velez/internal/pipelines/steps/network_steps"
)

const ConnectServiceToVpnAction = "connect_service_to_vpn"

// Accessor interfaces the connect_service_to_vpn jobs need from their
// TaskContext. *velez_api.ConnectServiceToVpnTaskPayload satisfies all of
// them. containerIdAccessor is declared in create_smerd.go and reused here
// as-is.

type connectServiceToVpnRequestAccessor interface {
	GetServiceName() string
}

type namespaceIdAccessor interface {
	GetNamespaceId() string
	SetNamespaceId(string)
}

type clientKeyAccessor interface {
	GetClientKey() string
	SetClientKey(string)
}

type loginServerUrlAccessor interface {
	GetLoginServerUrl() string
	SetLoginServerUrl(string)
}

type connectServiceToVpnHandler struct {
	nodeClients      node_clients.NodeClients
	vpnClient        cluster_clients.VervClosedNetworkClient
	serviceDiscovery cluster_clients.ServiceDiscovery
}

func NewConnectServiceToVpnHandler(
	nodeClients node_clients.NodeClients,
	vpnClient cluster_clients.VervClosedNetworkClient,
	serviceDiscovery cluster_clients.ServiceDiscovery,
) TaskHandler {
	return &connectServiceToVpnHandler{
		nodeClients:      nodeClients,
		vpnClient:        vpnClient,
		serviceDiscovery: serviceDiscovery,
	}
}

func (h *connectServiceToVpnHandler) Action() string {
	return ConnectServiceToVpnAction
}

func (h *connectServiceToVpnHandler) NewContext() TaskContext {
	return &velez_api.ConnectServiceToVpnTaskPayload{}
}

// BuildJobs folds do_connect_service_to_vpn.go's inline env-append closure
// step into createSidecarContainerJob (question-worthy per
// docs/jobs_migrations/questions.md #2's precedent for copy_to_volume: it's
// pure logic with no side effect of its own, so it doesn't need its own
// checkpoint row) - 9 pipeline steps become 8 named jobs.
func (h *connectServiceToVpnHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload := taskCtx.(*velez_api.ConnectServiceToVpnTaskPayload)

	serviceName := payload.GetServiceName()
	containerName := serviceName + "-" + patterns.TailscaleSidecarSuffix
	hostname := strings.ReplaceAll(containerName, "_", "-")
	launchContainer := patterns.TailScaleContainerSidecar(serviceName)

	return []NamedJob{
		{
			// network_steps.CheckSidecarExist has no result to persist and no
			// pipeline-only dependency, so it's reused as-is - steps.Step and
			// jobs.Job are both just Do(ctx) error (see create_service.go's
			// reuse of service_steps.ValidateServiceName).
			Name: "check_sidecar",
			Job:  network_steps.CheckSidecarExist(h.nodeClients, containerName),
		},
		{
			Name: "prepare_namespace",
			Job: &prepareNamespaceJob{
				vpnClient: h.vpnClient,
				req:       payload,
				ctx:       payload,
			},
		},
		{
			Name: "get_client_key",
			Job: &getClientKeyJob{
				vpnClient: h.vpnClient,
				namespace: payload,
				ctx:       payload,
			},
		},
		{
			Name: "get_login_server_url",
			Job: &getLoginServerUrlJob{
				ctx: payload,
			},
		},
		{
			Name: "prepare_image",
			Job: &prepareSidecarImageJob{
				docker: h.nodeClients.Docker(),
				req:    &launchContainer,
			},
		},
		{
			Name: "create_container",
			Job: &createSidecarContainerJob{
				nodeClients:     h.nodeClients,
				launchContainer: &launchContainer,
				containerName:   containerName,
				hostname:        hostname,
				clientKey:       payload,
				loginUrl:        payload,
				ctx:             payload,
			},
		},
		{
			Name: "start_container",
			Job: &startSidecarContainerJob{
				dockerAPI: h.nodeClients.Docker().Client(),
				ctx:       payload,
			},
		},
		{
			Name: "add_makosh_record",
			Job: &addMakoshRecordJob{
				sd:          h.serviceDiscovery,
				serviceName: serviceName,
				hostname:    hostname,
			},
		},
	}
}

type prepareNamespaceJob struct {
	vpnClient cluster_clients.VervClosedNetworkClient

	req connectServiceToVpnRequestAccessor
	ctx namespaceIdAccessor
}

func (j *prepareNamespaceJob) Do(ctx context.Context) error {
	serviceName := j.req.GetServiceName()

	namespace, err := j.vpnClient.GetNamespace(ctx, serviceName)
	if err != nil {
		return rerrors.Wrap(err, "error getting vcs namespace when preparing")
	}

	if namespace.Id == "" {
		namespace, err = j.vpnClient.CreateNamespace(ctx, serviceName)
		if err != nil {
			return rerrors.Wrap(err, "error creating vcs namespace when preparing")
		}
	}

	j.ctx.SetNamespaceId(namespace.Id)

	return nil
}

type getClientKeyJob struct {
	vpnClient cluster_clients.VervClosedNetworkClient

	namespace namespaceIdAccessor
	ctx       clientKeyAccessor
}

// Do mirrors network_steps.GetClientKey, except do_connect_service_to_vpn.go's
// original "existing key found" branch (`h.keyResponse = &authKey.Key`)
// reassigns the step's own pointer field instead of writing through it, so
// the found key was silently never returned to the caller - the pipeline
// always fell through to issuing a brand-new key. The Set-based job model
// makes that bug structurally unreproducible (SetClientKey always writes the
// value), so this job returns the existing key correctly. Flagged as a
// behavior improvement, not a deliberate migration decision.
func (j *getClientKeyJob) Do(ctx context.Context) error {
	namespaceId := j.namespace.GetNamespaceId()

	getAuthKeyReq := domain.GetVcnAuthKeyReq{
		NamespaceId:  namespaceId,
		ReusableOnly: true,
	}

	authKey, err := j.vpnClient.GetClientAuthKey(ctx, getAuthKeyReq)
	if err != nil {
		if !rerrors.Is(err, headscale.ErrNotFound) {
			return rerrors.Wrap(err)
		}
	}

	if authKey.Key != "" {
		j.ctx.SetClientKey(authKey.Key)
		return nil
	}

	issueClientKeyReq := domain.IssueClientKey{
		NamespaceId: namespaceId,
		Reusable:    true,
	}

	clientKey, err := j.vpnClient.IssueClientKey(ctx, issueClientKeyReq)
	if err != nil {
		return rerrors.Wrap(err)
	}

	j.ctx.SetClientKey(clientKey)

	return nil
}

type getLoginServerUrlJob struct {
	ctx loginServerUrlAccessor
}

func (j *getLoginServerUrlJob) Do(_ context.Context) error {
	// TODO For multiple nodes implement different urls - mirrors
	// network_steps.GetLoginServerUrl's hardcoded constant.
	j.ctx.SetLoginServerUrl("https://vcn.redsock.ru")

	return nil
}

type prepareSidecarImageJob struct {
	docker node_clients.Docker

	req *container.CreateRequest
}

func (j *prepareSidecarImageJob) Do(ctx context.Context) error {
	_, err := j.docker.PullImage(ctx, j.req.Config.Image)
	if err != nil {
		return rerrors.Wrap(err, "error pulling image")
	}

	return nil
}

// createSidecarContainerJob mirrors container_steps.Create restricted to
// what the tailscale sidecar needs. Unlike container_steps.Create, it doesn't
// tolerate an existing container of the same name (docker.ErrNameIsTaken) or
// clean up volume mounts on rollback - it follows create_smerd.go's/
// copy_to_volume.go's simpler create+Remove pattern instead, consistent with
// how this migration already treats container create/rollback elsewhere.
type createSidecarContainerJob struct {
	nodeClients node_clients.NodeClients

	launchContainer *container.CreateRequest
	containerName   string
	hostname        string

	clientKey clientKeyAccessor
	loginUrl  loginServerUrlAccessor
	ctx       containerIdAccessor
}

func (j *createSidecarContainerJob) Do(ctx context.Context) error {
	j.launchContainer.Config.Env = append(j.launchContainer.Config.Env,
		"TS_HOSTNAME="+j.hostname,
		"TS_AUTHKEY="+j.clientKey.GetClientKey(),
		"TS_EXTRA_ARGS=--login-server="+j.loginUrl.GetLoginServerUrl(),
	)

	dockerClient := j.nodeClients.Docker()

	created, err := dockerClient.ContainerCreate(ctx,
		j.launchContainer.Config, j.launchContainer.HostConfig, j.launchContainer.NetworkingConfig,
		&v1.Platform{}, j.containerName)
	if err != nil {
		return rerrors.Wrap(err, "error creating sidecar container")
	}

	j.ctx.SetContainerId(created.ID)

	return nil
}

func (j *createSidecarContainerJob) Rollback(ctx context.Context) error {
	containerId := j.ctx.GetContainerId()
	if containerId == "" {
		return nil
	}

	err := j.nodeClients.Docker().Remove(ctx, containerId)
	if err != nil && !errdefs.IsNotFound(err) {
		return rerrors.Wrapf(err, "error removing sidecar container '%s'", containerId)
	}

	return nil
}

// startSidecarContainerJob depends on the narrow startAPI interface declared
// in copy_to_volume.go rather than client.APIClient directly (see
// docs/jobs_migrations/questions.md #8), for full unit-test coverage.
type startSidecarContainerJob struct {
	dockerAPI startAPI

	ctx containerIdAccessor
}

func (j *startSidecarContainerJob) Do(ctx context.Context) error {
	containerId := j.ctx.GetContainerId()
	if containerId == "" {
		return rerrors.New("no container id provided")
	}

	err := j.dockerAPI.ContainerStart(ctx, containerId, container.StartOptions{})
	if err != nil {
		return rerrors.Wrap(err, "error starting sidecar container")
	}

	return nil
}

func (j *startSidecarContainerJob) Rollback(ctx context.Context) error {
	containerId := j.ctx.GetContainerId()
	if containerId == "" {
		return nil
	}

	err := j.dockerAPI.ContainerStop(ctx, containerId, container.StopOptions{})
	if err != nil {
		return rerrors.Wrapf(err, "error stopping sidecar container '%s'", containerId)
	}

	return nil
}

type addMakoshRecordJob struct {
	sd cluster_clients.ServiceDiscovery

	serviceName string
	hostname    string
}

func (j *addMakoshRecordJob) Do(ctx context.Context) error {
	upsertReq := &makosh_be.UpsertEndpoints_Request{
		Endpoints: []*makosh_be.Endpoint{
			{
				ServiceName: j.serviceName,
				Addrs:       []string{j.hostname},
			},
		},
	}

	_, err := j.sd.UpsertEndpoints(ctx, upsertReq)
	if err != nil {
		return rerrors.Wrap(err, "error during upsertion of makosh endpoints")
	}

	return nil
}
