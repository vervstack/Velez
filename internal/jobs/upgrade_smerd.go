package jobs

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"maps"
	"strconv"
	"strings"

	errdefs2 "github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/evon"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/pipelines/steps/upgrade_steps"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/matreshka/pkg/matreshka"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	UpgradeSmerdAction = "upgrade_smerd"

	configFetcherContainerSuffix = "_configuration_fetcher"
	newContainerSuffix           = "_new"
	oldContainerSuffix           = "_old"

	stepCheckSelfUpgrade             = "check_self_upgrade"
	stepCaptureOldContainer          = "capture_old_container"
	stepPauseOldContainer            = "pause_old_container"
	stepCreateConfigFetcherContainer = "create_config_fetcher_container"
	stepGetConfigFromContainer       = "get_config_from_container"
	stepDropConfigFetcherContainer   = "drop_config_fetcher_container"
	stepCreateFinalContainer         = "create_final_container"
	stepStartFinalContainer          = "start_final_container"
	stepRenameOldContainer           = "rename_old_container"
	stepDropOldContainer             = "drop_old_container"
	stepRenameNewContainer           = "rename_new_container"
)

// Accessor interfaces the upgrade_smerd jobs need from their TaskContext.
// *velez_api.UpgradeSmerdTaskPayload satisfies all of them.
// smerdRequestAccessor, containerIDAccessor and imageMetaAccessor are
// declared in create_smerd.go/assemble_config.go and reused here as-is.

type upgradeRequestAccessor interface {
	GetUpgradeRequest() *velez_api.UpgradeSmerd_Request
}

type oldContainerIDAccessor interface {
	GetOldContainerId() string
	SetOldContainerId(string)
}

type captureOldContainerCtx interface {
	SetRequest(*velez_api.CreateSmerd_Request)
	oldContainerIDAccessor
}

// oldContainerAsCurrent adapts oldContainerIDAccessor to containerIDAccessor
// so jobs already written against "the current container id"
// (dropScratchContainerJob, renameContainerJob) can be reused unmodified
// against old_container_id too, instead of a second near-identical copy of
// each job type.
type oldContainerAsCurrent struct {
	ctx oldContainerIDAccessor
}

func (a oldContainerAsCurrent) GetContainerId() string {
	return a.ctx.GetOldContainerId()
}

func (a oldContainerAsCurrent) SetContainerId(v string) {
	a.ctx.SetOldContainerId(v)
}

type upgradeSmerdHandler struct {
	nodeClients      node_clients.NodeClients
	containerService service.ContainerService
	configService    service.ConfigurationService
}

func NewUpgradeSmerdHandler(
	nodeClients node_clients.NodeClients,
	containerService service.ContainerService,
	configService service.ConfigurationService,
) TaskHandler {
	return &upgradeSmerdHandler{
		nodeClients:      nodeClients,
		containerService: containerService,
		configService:    configService,
	}
}

func (h *upgradeSmerdHandler) Action() string {
	return UpgradeSmerdAction
}

func (h *upgradeSmerdHandler) NewContext() TaskContext {
	return &velez_api.UpgradeSmerdTaskPayload{}
}

func (h *upgradeSmerdHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.UpgradeSmerdTaskPayload)
	if !ok {
		panic("upgrade_smerd: BuildJobs called with mismatched TaskContext type")
	}

	dockerAPI := h.nodeClients.Docker().Client()

	return []NamedJob{
		{
			Name: stepCheckSelfUpgrade,
			Job: &checkSelfUpgradeJob{
				containerService: h.containerService,
				upgradeReq:       payload,
			},
		},
		{
			Name: stepCaptureOldContainer,
			Job: &captureOldContainerJob{
				containerService: h.containerService,
				upgradeReq:       payload,
				ctx:              payload,
			},
		},
		{
			Name: stepPrepareCreateImage,
			Job: &prepareUpgradeImageJob{
				docker:     h.nodeClients.Docker(),
				upgradeReq: payload,
				ctx:        payload,
			},
		},
		{
			Name: stepPauseOldContainer,
			Job: &pauseOldContainerJob{
				dockerAPI:   dockerAPI,
				portManager: h.nodeClients.PortManager(),
				ctx:         payload,
			},
		},
		{
			Name: stepCreateConfigFetcherContainer,
			Job: &renamingCreateContainerJob{
				nodeClients: h.nodeClients,
				req:         payload,
				ctx:         payload,
				newName: func(current string) string {
					return current + configFetcherContainerSuffix
				},
			},
		},
		{
			Name: stepGetConfigFromContainer,
			Job: &getConfigFromScratchContainerJob{
				dockerAPI: dockerAPI,
				imageMeta: payload,
				ctx:       payload,
			},
		},
		{
			Name: stepDropConfigFetcherContainer,
			Job: &dropScratchContainerJob{
				docker: h.nodeClients.Docker(),
				ctx:    payload,
			},
		},
		{
			Name: stepFetchConfig,
			Job: &fetchUpgradeConfigJob{
				configService: h.configService,
				upgradeReq:    payload,
				imageMeta:     payload,
				req:           payload,
			},
		},
		{
			Name: stepPrepareVervConfig,
			Job: &prepareUpgradeVervConfigJob{
				dockerAPI:   dockerAPI,
				portManager: h.nodeClients.PortManager(),
				imageMeta:   payload,
				req:         payload,
			},
		},
		{
			Name: stepCreateFinalContainer,
			Job: &renamingCreateContainerJob{
				nodeClients: h.nodeClients,
				req:         payload,
				ctx:         payload,
				newName: func(current string) string {
					return current + newContainerSuffix
				},
			},
		},
		{
			Name: stepStartFinalContainer,
			Job: &startContainerJob{
				dockerAPI: dockerAPI,
				ctx:       payload,
			},
		},
		{
			Name: stepHealthcheck,
			Job: &healthcheckJob{
				dockerAPI: dockerAPI,
				req:       payload,
				ctx:       payload,
			},
		},
		{
			Name: stepRenameOldContainer,
			Job: &renameContainerJob{
				dockerAPI: dockerAPI,
				ctx:       oldContainerAsCurrent{payload},
				newName:   payload.GetUpgradeRequest().GetName() + oldContainerSuffix,
			},
		},
		{
			Name: stepDropOldContainer,
			Job: &dropScratchContainerJob{
				docker: h.nodeClients.Docker(),
				ctx:    oldContainerAsCurrent{payload},
			},
		},
		{
			Name: stepRenameNewContainer,
			Job: &renameContainerJob{
				dockerAPI: dockerAPI,
				ctx:       payload,
				newName:   payload.GetUpgradeRequest().GetName(),
			},
		},
	}
}

type checkSelfUpgradeJob struct {
	containerService service.ContainerService

	upgradeReq upgradeRequestAccessor
}

// Do duplicates upgrade_steps.CheckUpgradeIsAvailable's self-upgrade guard.
// That function takes a full service.Services (only to call .SmerdManager()
// internally), while this handler takes the narrower ContainerService
// directly, so its few lines of logic are reproduced here rather than
// widening the handler's dependency. The sentinel error is reused as-is.
func (j *checkSelfUpgradeJob) Do(ctx context.Context) error {
	id := env.GetContainerId()
	if id == nil {
		return nil
	}

	smerd, err := j.containerService.InspectSmerd(ctx, j.upgradeReq.GetUpgradeRequest().GetName())
	if err != nil {
		return rerrors.Wrap(err)
	}

	if smerd.GetUuid() == *id {
		return upgrade_steps.ErrSelfUpgradeIsForbidden
	}

	return nil
}

type captureOldContainerJob struct {
	containerService service.ContainerService

	upgradeReq upgradeRequestAccessor
	ctx        captureOldContainerCtx
}

// Do mirrors steps.fromContainerToRequest.Do, folding in the immediately
// adjacent SingleFunc step that sets ImageName - see
// docs/jobs_migrations/questions.md for the fold precedent.
func (j *captureOldContainerJob) Do(ctx context.Context) error {
	name := j.upgradeReq.GetUpgradeRequest().GetName()

	cont, err := j.containerService.InspectSmerd(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container")
	}

	req := &velez_api.CreateSmerd_Request{
		Name:      cont.GetName(),
		ImageName: j.upgradeReq.GetUpgradeRequest().GetImage(),
		Settings: &velez_api.Container_Settings{
			Ports:   cont.GetPorts(),
			Network: fromContainerNetwork(cont),
			Volumes: cont.GetVolumes(),
		},
		Env:    cont.GetEnv(),
		Labels: cont.GetLabels(),
	}

	j.ctx.SetRequest(req)
	j.ctx.SetOldContainerId(cont.GetUuid())

	return nil
}

func fromContainerNetwork(cont *velez_api.Smerd) []*velez_api.NetworkBind {
	out := make([]*velez_api.NetworkBind, 0, len(cont.GetNetworks()))

	for _, n := range cont.GetNetworks() {
		net := &velez_api.NetworkBind{NetworkName: n.GetNetworkName()}
		for _, a := range n.GetAliases() {
			if !strings.HasPrefix(cont.GetUuid(), a) {
				net.Aliases = append(net.Aliases, a)
			}
		}

		if len(net.GetAliases()) != 0 {
			out = append(out, net)
		}
	}

	return out
}

type prepareUpgradeImageJob struct {
	docker node_clients.Docker

	upgradeReq upgradeRequestAccessor
	ctx        imageMetaAccessor
}

func (j *prepareUpgradeImageJob) Do(ctx context.Context) error {
	imageInfo, err := j.docker.PullImage(ctx, j.upgradeReq.GetUpgradeRequest().GetImage())
	if err != nil {
		return rerrors.Wrap(err, "error pulling image")
	}

	var imageLabels map[string]string

	if imageInfo.Config != nil {
		imageLabels = imageInfo.Config.Labels
	}

	j.ctx.SetImageLabels(imageLabels)
	j.ctx.SetImageTags(imageInfo.RepoTags)

	return nil
}

// pauseAPI is the narrow slice of client.APIClient pauseOldContainerJob needs
// to pause/stop the old container and disconnect/reconnect its networks on
// rollback. Kept narrow (rather than depending on client.APIClient directly)
// so it's hand-fakeable in unit tests, same rationale as copy_to_volume.go's
// startAPI/copyAPI - client.APIClient's method set is a superset of
// pauseAPI's, so the real client still satisfies it.
type pauseAPI interface {
	ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error)
	ContainerPause(ctx context.Context, containerID string) error
	ContainerUnpause(ctx context.Context, containerID string) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	NetworkDisconnect(ctx context.Context, networkID, containerID string, force bool) error
	NetworkConnect(ctx context.Context, networkID, containerID string, config *network.EndpointSettings) error
}

type pauseOldContainerJob struct {
	dockerAPI   pauseAPI
	portManager node_clients.PortManager

	ctx oldContainerIDAccessor

	stateBeforePause container.ContainerState
	disconnectedNets map[string]*network.EndpointSettings
	portsOnHold      []uint32
}

// Do mirrors smerd_steps.detachContainerFromVervStep.Do, reading the old
// container id from the persisted context at Do()-time instead of a raw
// pointer fixed at construction. dockerutils.DisconnectFromNetworks/
// ConnectToNetwork can't be reused directly since they're parameterized on
// the full client.APIClient rather than an interface a fake could satisfy -
// same wall copy_to_volume.go hit (see docs/jobs_migrations/questions.md
// #8) - so their logic is duplicated below against pauseAPI instead.
func (j *pauseOldContainerJob) Do(ctx context.Context) error {
	containerID := j.ctx.GetOldContainerId()
	if containerID == "" {
		return rerrors.New("container id is required")
	}

	cont, err := j.dockerAPI.ContainerInspect(ctx, containerID)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container")
	}

	j.stateBeforePause = cont.State.Status

	err = j.stopContainer(ctx, containerID, cont)
	if err != nil {
		return rerrors.Wrap(err, "error stopping container")
	}

	j.disconnectedNets, err = disconnectFromNetworks(ctx, j.dockerAPI, containerID)
	if err != nil {
		return rerrors.Wrap(err, "error disconnecting from network")
	}

	for _, hostPorts := range cont.NetworkSettings.Ports {
		for _, hostPort := range hostPorts {
			port, _ := strconv.ParseUint(hostPort.HostPort, 10, 32)
			p := uint32(port)
			j.portManager.HoldPort(p)

			j.portsOnHold = append(j.portsOnHold, p)
		}
	}

	return nil
}

func (j *pauseOldContainerJob) stopContainer(
	ctx context.Context, containerID string, cont container.InspectResponse,
) error {
	switch cont.State.Status {
	case container.StateRunning:
		err := j.dockerAPI.ContainerPause(ctx, containerID)
		if err != nil {
			if !errdefs2.IsConflict(err) {
				return rerrors.Wrap(err, "error pausing container")
			}
		}
	case container.StateRestarting:
		err := j.dockerAPI.ContainerStop(ctx, containerID, container.StopOptions{})
		if err != nil {
			return rerrors.Wrap(err, "error stopping container")
		}
	}

	return nil
}

func (j *pauseOldContainerJob) Rollback(ctx context.Context) error {
	containerID := j.ctx.GetOldContainerId()
	if containerID == "" {
		return nil
	}

	if j.stateBeforePause != container.StateRunning {
		return nil
	}

	for _, p := range j.portsOnHold {
		j.portManager.UnHoldPort(p)
	}

	err := j.dockerAPI.ContainerUnpause(ctx, containerID)
	if err != nil {
		return rerrors.Wrapf(err, "error unpausing container '%s'", containerID)
	}

	var globErr error

	for netName, net := range j.disconnectedNets {
		err = connectToNetwork(ctx, j.dockerAPI, netName, containerID, net.Aliases)
		if err != nil {
			globErr = rerrors.Join(globErr, rerrors.Wrap(err, "error connecting to network on rollback"))
		}
	}

	if globErr != nil {
		return rerrors.Wrap(globErr)
	}

	return nil
}

// disconnectFromNetworks duplicates dockerutils.DisconnectFromNetworks
// against the narrow pauseAPI interface - see pauseOldContainerJob.Do's
// comment for why the dockerutils helper itself can't be called here.
func disconnectFromNetworks(
	ctx context.Context, d pauseAPI, contID string,
) (map[string]*network.EndpointSettings, error) {
	cont, err := d.ContainerInspect(ctx, contID)
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting container info")
	}

	disconnected := make(map[string]*network.EndpointSettings)

	for netName, net := range cont.NetworkSettings.Networks {
		err = d.NetworkDisconnect(ctx, netName, cont.Name, false)
		if err != nil {
			return nil, rerrors.Wrap(err, "error disconnecting from network")
		}

		disconnected[netName] = net
	}

	return disconnected, nil
}

// connectToNetwork duplicates dockerutils.ConnectToNetwork against the
// narrow pauseAPI interface, for the same reason as disconnectFromNetworks.
func connectToNetwork(ctx context.Context, d pauseAPI, networkName, contID string, aliases []string) error {
	cont, err := d.ContainerInspect(ctx, contID)
	if err != nil {
		return rerrors.Wrap(err, "error getting container info")
	}

	isConnected := cont.NetworkSettings != nil && cont.NetworkSettings.Networks[networkName] != nil
	if !isConnected {
		conn := &network.EndpointSettings{Aliases: aliases}

		err = d.NetworkConnect(ctx, networkName, cont.Name, conn)
		if err != nil {
			return rerrors.Wrap(err, "error connecting to network")
		}
	}

	return nil
}

// renamingCreateContainerJob renames the shared Request before delegating to
// createContainerJob, mirroring do_smerd_upgrade.go's SingleFunc-then-Create
// step pairs: both container-create stages reuse the very same Request
// value, only its Name differs between them.
type renamingCreateContainerJob struct {
	nodeClients node_clients.NodeClients

	req smerdRequestAccessor
	ctx containerIDAccessor

	newName func(current string) string
}

func (j *renamingCreateContainerJob) Do(ctx context.Context) error {
	request := j.req.GetRequest()

	request.Name = j.newName(request.GetName())

	inner := &createContainerJob{nodeClients: j.nodeClients, req: j.req, ctx: j.ctx}

	return inner.Do(ctx)
}

func (j *renamingCreateContainerJob) Rollback(ctx context.Context) error {
	inner := &createContainerJob{nodeClients: j.nodeClients, ctx: j.ctx}

	return inner.Rollback(ctx)
}

// copyFromAPI is the narrow slice of client.APIClient
// getConfigFromScratchContainerJob needs to read a file out of a container.
// dockerutils.ReadFromContainer is parameterized on the full client.APIClient
// rather than an interface, so its tar-extraction logic is duplicated below
// against copyFromAPI instead - same wall as pauseAPI (see
// docs/jobs_migrations/questions.md #8).
type copyFromAPI interface {
	CopyFromContainer(ctx context.Context, containerID, srcPath string) (io.ReadCloser, container.PathStat, error)
}

type getConfigFromScratchContainerJob struct {
	dockerAPI copyFromAPI

	imageMeta imageMetaAccessor
	ctx       containerIDAccessor
}

// Do mirrors config_steps.getConfigFromContainerStep, restricted to the
// shape UpgradeSmerd actually exercises (same restriction style as
// assemble_config.go's fetchConfigJob). Its result - the container's own
// default config file, if any - is discarded here exactly as it is in the
// original pipeline: do_smerd_upgrade.go extracts it into cfgMount but never
// reads cfgMount again afterwards. That looks like dead/unfinished wiring
// rather than intentional behavior; see docs/jobs_migrations/questions.md -
// preserved as-is rather than fixed, per the behavior-preserving migration
// rule.
func (j *getConfigFromScratchContainerJob) Do(ctx context.Context) error {
	containerID := j.ctx.GetContainerId()
	if containerID == "" {
		return rerrors.New("empty container id")
	}

	_, _, systemPath := classifyImage(j.imageMeta.GetImageLabels(), j.imageMeta.GetImageTags())
	if systemPath == "" {
		return nil
	}

	_, err := readFileFromContainer(ctx, j.dockerAPI, containerID, systemPath)
	if err != nil {
		return rerrors.Wrap(err, "error getting config to mount")
	}

	return nil
}

// readFileFromContainer duplicates dockerutils.ReadFromContainer against the
// narrow copyFromAPI interface.
func readFileFromContainer(ctx context.Context, d copyFromAPI, contID, path string) ([]byte, error) {
	rc, _, err := d.CopyFromContainer(ctx, contID, path)
	if err != nil {
		return nil, rerrors.Wrap(err, "error copying from container")
	}

	defer func() {
		closeErr := rc.Close()
		if closeErr != nil {
			log.Error().Err(closeErr).Msg("error closing container copy-from reader")
		}
	}()

	reader := tar.NewReader(rc)

	_, err = reader.Next()
	if err != nil {
		return nil, rerrors.Wrap(err, "error reading tar header")
	}

	res, err := io.ReadAll(reader)
	if err != nil {
		return nil, rerrors.Wrap(err, "error reading config from tar")
	}

	return res, nil
}

type fetchUpgradeConfigJob struct {
	configService service.ConfigurationService

	upgradeReq upgradeRequestAccessor
	imageMeta  imageMetaAccessor
	req        smerdRequestAccessor
}

// Do mirrors config_steps.fetchConfigStep.doVerv+setEnv. UpgradeSmerd never
// sets req.Config (capture_old_container leaves it nil), so the original
// step's doPlain/yaml branches are unreachable here - restricted to the env
// branch only, same restriction style as assemble_config.go's fetchConfigJob.
// Also restores Request.Name to the upgrade's target name (the SingleFunc
// step immediately preceding FetchConfig in the original pipeline), since
// the preceding create_config_fetcher_container stage left it renamed.
func (j *fetchUpgradeConfigJob) Do(ctx context.Context) error {
	request := j.req.GetRequest()

	request.Name = j.upgradeReq.GetUpgradeRequest().GetName()

	confType, format, _ := classifyImage(j.imageMeta.GetImageLabels(), j.imageMeta.GetImageTags())

	configName := request.GetName()
	if confType != confTypePlain && !strings.HasPrefix(configName, confType) {
		configName = confType + "_" + configName
	}

	meta := domain.ConfigMeta{
		Name:     configName,
		ConfType: toConfTypePrefix(confType),
		Format:   format,
	}

	envEvon, err := j.configService.GetEnvFromApi(ctx, meta)
	if err != nil {
		code := status.Code(err)
		if code != codes.NotFound {
			return rerrors.Wrap(err, "error getting matreshka config from matreshka api")
		}
	}

	ns := evon.NodeStorage{}
	ns.AddNode(envEvon)

	for _, n := range ns {
		if n.Value == nil {
			continue
		}

		request.Env[n.Name] = fmt.Sprint(n.Value)
	}

	return nil
}

func toConfTypePrefix(s string) matreshka_api.ConfigTypePrefix {
	switch s {
	case confTypeVerv:
		return matreshka_api.ConfigTypePrefix_verv
	case confTypePg:
		return matreshka_api.ConfigTypePrefix_pg
	default:
		return matreshka_api.ConfigTypePrefix_plain
	}
}

// createNetworkAPI is the narrow slice of client.APIClient
// prepareUpgradeVervConfigJob needs to ensure a service's networks exist.
// dockerutils.CreateNetwork is parameterized on the full client.APIClient
// rather than an interface, so its logic is duplicated below against
// createNetworkAPI instead - same wall as pauseAPI (see
// docs/jobs_migrations/questions.md #8).
type createNetworkAPI interface {
	NetworkList(ctx context.Context, options network.ListOptions) ([]network.Summary, error)
	NetworkCreate(ctx context.Context, name string, options network.CreateOptions) (network.CreateResponse, error)
}

type prepareUpgradeVervConfigJob struct {
	dockerAPI   createNetworkAPI
	portManager node_clients.PortManager

	imageMeta imageMetaAccessor
	req       smerdRequestAccessor

	lockedPorts []uint32
}

// Do mirrors steps.prepareVervConfig, restricted to the branches UpgradeSmerd
// actually reaches: capture_old_container always builds a Request with
// UseImagePorts=false, IgnoreConfig=false and AutoUpgrade=false, so
// getPortsFromImage, the IgnoreConfig label-disable branch and the
// AutoUpgrade label are all unreachable here (same restriction style as
// assemble_config.go's fetchConfigJob).
func (j *prepareUpgradeVervConfigJob) Do(ctx context.Context) error {
	request := j.req.GetRequest()

	err := j.lockPorts(request)
	if err != nil {
		return rerrors.Wrap(err, "error locking ports")
	}

	request.Env[matreshka.VervName] = request.GetName()

	maps.Copy(request.GetLabels(), j.imageMeta.GetImageLabels())

	request.Labels[labels.ComposeGroupLabel] = request.GetName()

	for _, n := range request.GetSettings().GetNetwork() {
		err = createNetworkIfMissing(ctx, j.dockerAPI, n.GetNetworkName())
		if err != nil {
			return rerrors.Wrap(err, "error creating network: %s", n.GetNetworkName())
		}
	}

	return nil
}

// createNetworkIfMissing duplicates dockerutils.CreateNetwork against the
// narrow createNetworkAPI interface.
func createNetworkIfMissing(ctx context.Context, dockerAPI createNetworkAPI, networkName string) error {
	f := filters.NewArgs()
	f.Add("name", networkName)

	nets, err := dockerAPI.NetworkList(ctx, network.ListOptions{Filters: f})
	if err != nil {
		return rerrors.Wrap(err, "error inspecting network")
	}

	for _, n := range nets {
		if n.Name == networkName {
			return nil
		}
	}

	createOpts := network.CreateOptions{
		Driver: "bridge",
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{
				{
					// TODO make it auto-configurable among cluster - matches
					// dockerutils.CreateNetwork's own TODO.
					Subnet: "10.0.1.0/24",
				},
			},
		},
	}

	_, err = dockerAPI.NetworkCreate(ctx, networkName, createOpts)
	if err != nil {
		return rerrors.Wrap(err, "error creating network")
	}

	return nil
}

func (j *prepareUpgradeVervConfigJob) lockPorts(request *velez_api.CreateSmerd_Request) (err error) {
	j.lockedPorts = make([]uint32, 0, len(request.GetSettings().GetPorts()))

	for _, p := range request.GetSettings().GetPorts() {
		if p.ExposedTo == nil {
			var port uint32

			port, err = j.portManager.GetPort()
			p.ExposedTo = &port
		} else {
			ok := j.portManager.UnHoldPort(p.GetExposedTo())
			if !ok {
				err = j.portManager.LockPort(p.GetExposedTo())
			}
		}

		if err != nil {
			return rerrors.Wrap(err, "error locking host port")
		}

		j.lockedPorts = append(j.lockedPorts, p.GetExposedTo())
	}

	return nil
}

func (j *prepareUpgradeVervConfigJob) Rollback(_ context.Context) error {
	for _, port := range j.lockedPorts {
		if !j.portManager.UnHoldPort(port) {
			j.portManager.UnlockPorts(j.lockedPorts)
		}
	}

	return nil
}

// renameAPI is the narrow slice of client.APIClient renameContainerJob
// needs. smerd_steps.RenameContainer calls these two methods directly on a
// client.APIClient field rather than through a dockerutils helper, so no
// existing function is being bypassed here - this interface only exists to
// make the job hand-fakeable in unit tests.
type renameAPI interface {
	ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error)
	ContainerRename(ctx context.Context, containerID, newContainerName string) error
}

type renameContainerJob struct {
	dockerAPI renameAPI

	ctx     containerIDAccessor
	newName string

	oldName string
}

func (j *renameContainerJob) Do(ctx context.Context) error {
	containerID := j.ctx.GetContainerId()
	if containerID == "" {
		return rerrors.New("container id is required")
	}

	cont, err := j.dockerAPI.ContainerInspect(ctx, containerID)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container")
	}

	j.oldName = cont.Name

	err = j.dockerAPI.ContainerRename(ctx, containerID, j.newName)
	if err != nil {
		return rerrors.Wrap(err, "error renaming container")
	}

	return nil
}

func (j *renameContainerJob) Rollback(ctx context.Context) error {
	containerID := j.ctx.GetContainerId()
	if containerID == "" {
		return nil
	}

	err := j.dockerAPI.ContainerRename(ctx, containerID, j.oldName)
	if err != nil {
		return rerrors.Wrap(err, "error renaming container on rollback")
	}

	return nil
}
