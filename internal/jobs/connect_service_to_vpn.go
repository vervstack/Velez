package jobs

import (
	"context"
	"strings"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/headscale"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/patterns"
	"go.vervstack.ru/Velez/internal/pipelines/steps/network_steps"
	"go.vervstack.ru/makosh/pkg/makosh_be"
)

const (
	ConnectServiceToVpnAction = "connect_service_to_vpn"

	stepCheckSidecar        = "check_sidecar"
	stepPrepareNamespace    = "prepare_namespace"
	stepGetClientKey        = "get_client_key"
	stepGetLoginServerURL   = "get_login_server_url"
	stepPrepareSidecarImage = "prepare_image"
	stepStartSidecar        = "start_container"
	stepAddMakoshRecord     = "add_makosh_record"
)

// Accessor interfaces the connect_service_to_vpn jobs need from their
// TaskContext. *velez_api.ConnectServiceToVpnTaskPayload satisfies all of
// them. containerIDAccessor is declared in create_smerd.go and reused here
// as-is.

type connectServiceToVpnRequestAccessor interface {
	GetServiceName() string
}

type namespaceIDAccessor interface {
	GetNamespaceId() string
	SetNamespaceId(string)
}

type clientKeyAccessor interface {
	GetClientKey() string
	SetClientKey(string)
}

type loginServerURLAccessor interface {
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
	payload, ok := taskCtx.(*velez_api.ConnectServiceToVpnTaskPayload)
	if !ok {
		panic("connect_service_to_vpn: BuildJobs called with mismatched TaskContext type")
	}

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
			Name: stepCheckSidecar,
			Job:  network_steps.CheckSidecarExist(h.nodeClients, containerName),
		},
		{
			Name: stepPrepareNamespace,
			Job: &prepareNamespaceJob{
				vpnClient: h.vpnClient,
				req:       payload,
				ctx:       payload,
			},
		},
		{
			Name: stepGetClientKey,
			Job: &getClientKeyJob{
				vpnClient: h.vpnClient,
				namespace: payload,
				ctx:       payload,
			},
		},
		{
			Name: stepGetLoginServerURL,
			Job: &getLoginServerURLJob{
				ctx: payload,
			},
		},
		{
			Name: stepPrepareSidecarImage,
			Job: &prepareSidecarImageJob{
				docker: h.nodeClients.Docker(),
				req:    &launchContainer,
			},
		},
		{
			Name: stepCreatePgContainer,
			Job: &createSidecarContainerJob{
				nodeClients:     h.nodeClients,
				launchContainer: &launchContainer,
				containerName:   containerName,
				hostname:        hostname,
				clientKey:       payload,
				loginURL:        payload,
				ctx:             payload,
			},
		},
		{
			Name: stepStartSidecar,
			Job: &startSidecarContainerJob{
				dockerAPI: h.nodeClients.Docker().Client(),
				ctx:       payload,
			},
		},
		{
			Name: stepAddMakoshRecord,
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
	ctx namespaceIDAccessor
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

	namespace namespaceIDAccessor
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
	namespaceID := j.namespace.GetNamespaceId()

	getAuthKeyReq := domain.GetVcnAuthKeyReq{
		NamespaceId:  namespaceID,
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
		NamespaceId: namespaceID,
		Reusable:    true,
	}

	clientKey, err := j.vpnClient.IssueClientKey(ctx, issueClientKeyReq)
	if err != nil {
		return rerrors.Wrap(err)
	}

	j.ctx.SetClientKey(clientKey)

	return nil
}

type getLoginServerURLJob struct {
	ctx loginServerURLAccessor
}

func (j *getLoginServerURLJob) Do(_ context.Context) error {
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
	_, err := j.docker.PullImage(ctx, j.req.Image)
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
	loginURL  loginServerURLAccessor
	ctx       containerIDAccessor
}

func (j *createSidecarContainerJob) Do(ctx context.Context) error {
	j.launchContainer.Env = append(j.launchContainer.Env,
		"TS_HOSTNAME="+j.hostname,
		"TS_AUTHKEY="+j.clientKey.GetClientKey(),
		"TS_EXTRA_ARGS=--login-server="+j.loginURL.GetLoginServerUrl(),
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
	containerID := j.ctx.GetContainerId()
	if containerID == "" {
		return nil
	}

	err := j.nodeClients.Docker().Remove(ctx, containerID)
	if err != nil && !errdefs.IsNotFound(err) {
		return rerrors.Wrapf(err, "error removing sidecar container '%s'", containerID)
	}

	return nil
}

// startSidecarContainerJob depends on the narrow startAPI interface declared
// in copy_to_volume.go rather than client.APIClient directly (see
// docs/jobs_migrations/questions.md #8), for full unit-test coverage.
type startSidecarContainerJob struct {
	dockerAPI startAPI

	ctx containerIDAccessor
}

func (j *startSidecarContainerJob) Do(ctx context.Context) error {
	containerID := j.ctx.GetContainerId()
	if containerID == "" {
		return rerrors.New("no container id provided")
	}

	err := j.dockerAPI.ContainerStart(ctx, containerID, container.StartOptions{})
	if err != nil {
		return rerrors.Wrap(err, "error starting sidecar container")
	}

	return nil
}

func (j *startSidecarContainerJob) Rollback(ctx context.Context) error {
	containerID := j.ctx.GetContainerId()
	if containerID == "" {
		return nil
	}

	err := j.dockerAPI.ContainerStop(ctx, containerID, container.StopOptions{})
	if err != nil {
		return rerrors.Wrapf(err, "error stopping sidecar container '%s'", containerID)
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
