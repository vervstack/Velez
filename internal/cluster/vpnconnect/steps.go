package vpnconnect

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"
	"go.vervstack.ru/makosh/pkg/makosh_be"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

// singleFunc adapts a plain closure into a step - the deleted
// internal/pipelines/steps.SingleFunc.
type singleFunc struct {
	f func(ctx context.Context) error
}

func newSingleFunc(f func(ctx context.Context) error) step {
	return &singleFunc{
		f: f,
	}
}

func (s *singleFunc) Do(ctx context.Context) error {
	return s.f(ctx)
}

// checkSidecarExist short-circuits the run with user_errors.ErrVpnResultAlreadyExists
// when the sidecar is already running, and prunes a dead one otherwise.
type checkSidecarExist struct {
	docker      node_clients.Docker
	sideCarName string
}

func newCheckSidecarExist(nc node_clients.NodeClients, sideCarName string) step {
	return &checkSidecarExist{
		docker:      nc.Docker(),
		sideCarName: sideCarName,
	}
}

func (s *checkSidecarExist) Do(ctx context.Context) error {
	// Check if there is a headscale to connect to
	r := &velez_api.ListSmerds_Request{
		Name: &s.sideCarName,
	}

	// Sidecars are node-level, not environment-scoped, so this lookup spans
	// every environment (empty suffix).
	conts, err := s.docker.ListContainers(ctx, r, "")
	if err != nil {
		return rerrors.Wrap(err, "error listing container")
	}

	if len(conts) == 0 {
		return nil
	}

	if conts[0].State == "running" {
		return rerrors.Wrap(user_errors.ErrVpnResultAlreadyExists, "container already running")
	}

	err = s.docker.Remove(ctx, conts[0].ID)
	if err != nil {
		return rerrors.Wrap(err, "error removing dead container")
	}

	return nil
}

type prepareNamespace struct {
	vcnClient     cluster_clients.VervClosedNetworkClient
	namespaceName *string

	namespaceIdResp *string
}

func newPrepareNamespace(
	vcnClient cluster_clients.VervClosedNetworkClient,
	namespaceName *string,
	namespaceIdResp *string,
) step {
	return &prepareNamespace{
		vcnClient:       vcnClient,
		namespaceName:   namespaceName,
		namespaceIdResp: namespaceIdResp,
	}
}

func (s *prepareNamespace) Do(ctx context.Context) error {
	namespace, err := s.vcnClient.GetNamespace(ctx, *s.namespaceName)
	if err != nil {
		return rerrors.Wrap(err, "error getting vcs namespace when preparing")
	}

	if namespace.Id == "" {
		namespace, err = s.vcnClient.CreateNamespace(ctx, *s.namespaceName)
		if err != nil {
			return rerrors.Wrap(err, "error creating vcs namespace when preparing")
		}
	}

	*s.namespaceIdResp = namespace.Id

	return nil
}

// getClientKeyStep issues a new client key if none exists in the network,
// or returns the existing one.
type getClientKeyStep struct {
	networkService cluster_clients.VervClosedNetworkClient

	namespaceId *string

	keyResponse *string
}

func newGetClientKey(
	vpnClient cluster_clients.VervClosedNetworkClient,
	namespaceId *string,
	keyResponse *string,
) step {
	return &getClientKeyStep{
		networkService: vpnClient,
		namespaceId:    namespaceId,

		keyResponse: keyResponse,
	}
}

// Do is carried over unchanged from the deleted
// internal/pipelines/steps/network_steps.GetClientKey, including its
// "existing key found" branch: that branch reassigns the step's own
// keyResponse pointer instead of writing through it, so a reusable key that
// already exists is never handed back to the caller and a brand-new key is
// always issued instead. internal/jobs/connect_service_to_vpn.go's
// getClientKeyJob documents and fixes the same defect on the jobs-engine
// path; it is preserved here because deleting internal/pipelines was
// explicitly a behavior-preserving move.
func (h *getClientKeyStep) Do(ctx context.Context) error {
	getAuthKeyReq := domain.GetVcnAuthKeyReq{
		NamespaceId:  *h.namespaceId,
		ReusableOnly: true,
	}

	authKey, err := h.networkService.GetClientAuthKey(ctx, getAuthKeyReq)
	if err != nil {
		if !rerrors.Is(err, user_errors.ErrNotFound) {
			return rerrors.Wrap(err, "error getting client auth key from network service")
		}
	}

	if authKey.Key != "" {
		h.keyResponse = &authKey.Key

		return nil
	}

	issueClientKeyReq := domain.IssueClientKey{
		NamespaceId: *h.namespaceId,
		Reusable:    true,
	}

	clientKey, err := h.networkService.IssueClientKey(ctx, issueClientKeyReq)
	if err != nil {
		return rerrors.Wrap(err, "error issuing client key bia network service")
	}

	*h.keyResponse = clientKey

	return nil
}

type getLoginServerUrlStep struct {
	responsePtr *string
}

func newGetLoginServerUrl(responsePtr *string) step {
	return &getLoginServerUrlStep{
		responsePtr: responsePtr,
	}
}

func (g *getLoginServerUrlStep) Do(_ context.Context) error {
	// TODO For multiple nodes implement different urls
	*g.responsePtr = "https://vcn.redsock.ru"
	// "http://headscale.verv:8080"
	return nil
}

type prepareImageStep struct {
	docker node_clients.Docker

	imageName string
}

func newPrepareImage(nc node_clients.NodeClients, imageName string) step {
	return &prepareImageStep{
		docker:    nc.Docker(),
		imageName: imageName,
	}
}

func (s *prepareImageStep) Do(ctx context.Context) error {
	_, err := s.docker.PullImage(ctx, s.imageName)
	if err != nil {
		return rerrors.Wrap(err, "error pulling image")
	}

	return nil
}

// createContainerStep creates the sidecar container.
//
// The deleted container_steps.Create declared a Rollback that removed the
// container and its volumes, but guarded it behind an isCreated flag nothing
// ever set to true - so it was unconditionally a no-op. It is therefore not
// carried over: omitting it is exactly equivalent to keeping it.
type createContainerStep struct {
	dockerClient node_clients.Docker

	req  *container.CreateRequest
	name *string
	// suffix - resolved environment suffix stamped onto the created
	// container. The VPN sidecar is node-level, so it is always empty here.
	suffix string

	containerIdResp *string
}

func newCreateContainer(
	nc node_clients.NodeClients,
	req *container.CreateRequest,
	name *string,
	suffix string,

	containerIdResp *string,
) step {
	return &createContainerStep{
		dockerClient:    nc.Docker(),
		req:             req,
		name:            name,
		suffix:          suffix,
		containerIdResp: containerIdResp,
	}
}

func (s *createContainerStep) Do(ctx context.Context) error {
	pCfg := &v1.Platform{}

	createResp, createErr := s.dockerClient.ContainerCreate(ctx,
		s.req.Config,
		s.req.HostConfig,
		s.req.NetworkingConfig,
		pCfg,
		toolbox.FromPtr(s.name),
		s.suffix,
	)
	if createErr != nil {
		if !rerrors.Is(createErr, user_errors.ErrNameIsTaken) {
			return rerrors.Wrap(createErr, "error creating container")
		}

		return rerrors.Wrap(createErr)
	}

	*s.containerIdResp = createResp.ID

	return nil
}

type startContainerStep struct {
	dockerAPI client.APIClient

	containerID *string
}

func newStartContainer(nc node_clients.NodeClients, containerID *string) step {
	return &startContainerStep{
		dockerAPI:   nc.Docker().Client(),
		containerID: containerID,
	}
}

func (s *startContainerStep) Do(ctx context.Context) error {
	if s.containerID == nil {
		return user_errors.ErrContainerIdMissing
	}

	err := s.dockerAPI.ContainerStart(ctx, *s.containerID, container.StartOptions{})
	if err != nil {
		return rerrors.Wrap(err, "error starting container")
	}

	return nil
}

func (s *startContainerStep) Rollback(ctx context.Context) error {
	if s.containerID == nil {
		return nil
	}

	err := s.dockerAPI.ContainerStop(ctx, *s.containerID, container.StopOptions{})
	if err != nil {
		return rerrors.Wrapf(err, "error removing container '%s'", *s.containerID)
	}

	return nil
}

type addMakoshRecord struct {
	sd          cluster_clients.ServiceDiscovery
	serviceName string
	vcnAddr     []string // Verv Closed Network address (both ip and hostname)
}

func newAddMakoshRecord(
	sd cluster_clients.ServiceDiscovery,
	serviceName string,
	vcnAddrs ...string,
) step {
	return &addMakoshRecord{
		sd:          sd,
		serviceName: serviceName,
		vcnAddr:     vcnAddrs,
	}
}

func (a *addMakoshRecord) Do(ctx context.Context) error {
	upsertReq := &makosh_be.UpsertEndpoints_Request{
		Endpoints: []*makosh_be.Endpoint{
			{
				ServiceName: a.serviceName,
				Addrs:       a.vcnAddr,
			},
		},
	}

	_, err := a.sd.UpsertEndpoints(ctx, upsertReq)
	if err != nil {
		return rerrors.Wrap(err, "error during upsertion of makosh endpoints")
	}

	return nil
}
