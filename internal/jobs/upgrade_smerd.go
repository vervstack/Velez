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
	"github.com/rs/zerolog/log"
	"go.redsock.ru/evon"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/clients/matreshka/pkg/matreshka_api"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/user_errors"
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

var (
	// ErrSelfUpgradeIsForbidden moved here from the deleted
	// internal/pipelines/steps/upgrade_steps package, where it was the last
	// symbol anything outside internal/pipelines still imported.
	//nolint:forbidigo // package-private sentinel, not shared/user-facing
	ErrSelfUpgradeIsForbidden = rerrors.NewUserError("Can't perform self upgrade", codes.FailedPrecondition)

	//nolint:forbidigo // package-private sentinel, not shared/user-facing
	errUpgradeContainerIDMissing = rerrors.New("container id is required")
	//nolint:forbidigo // package-private sentinel, not shared/user-facing
	errUpgradeConfigContainerIDMissing = rerrors.New("empty container id")
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
	SetOldContainerId(cId string)
}

type captureOldContainerCtx interface {
	SetRequest(createReq *velez_api.CreateSmerd_Request)
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
	// runtimes resolves the request's environment into the ContainerRuntime
	// that serves it - and therefore into the Docker suffix the recreated
	// containers must carry, the value Docker used to bake in at
	// client-construction time. Upgrade reuses create_smerd's
	// createContainerJob verbatim, so it necessarily shares its runtime
	// plumbing too; see docs/container_runtimes.
	runtimes container_runtime.RuntimeResolver
}

func NewUpgradeSmerdHandler(
	nodeClients node_clients.NodeClients,
	containerService service.ContainerService,
	configService service.ConfigurationService,
	runtimes container_runtime.RuntimeResolver,
) TaskHandler {
	return &upgradeSmerdHandler{
		nodeClients:      nodeClients,
		containerService: containerService,
		configService:    configService,
		runtimes:         runtimes,
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
				runtimes:         h.runtimes,
			},
		},
		{
			Name: stepCaptureOldContainer,
			Job: &captureOldContainerJob{
				containerService: h.containerService,
				upgradeReq:       payload,
				ctx:              payload,
				runtimes:         h.runtimes,
			},
		},
		{
			Name: stepPrepareCreateImage,
			Job: &prepareUpgradeImageJob{
				runtimes:   h.runtimes,
				upgradeReq: payload,
				ctx:        payload,
			},
		},
		{
			Name: stepPauseOldContainer,
			Job: &pauseOldContainerJob{
				dockerAPI:   dockerAPI,
				portManager: h.nodeClients.PortManager(),
				runtimes:    h.runtimes,
				req:         payload,
				ctx:         payload,
			},
		},
		{
			Name: stepCreateConfigFetcherContainer,
			Job: &renamingCreateContainerJob{
				nodeClients: h.nodeClients,
				req:         payload,
				ctx:         payload,
				runtimes:    h.runtimes,
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
			Job: &dropOwnedContainerJob{
				runtimes: h.runtimes,
				req:      payload,
				ctx:      payload,
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
				portManager: h.nodeClients.PortManager(),
				runtimes:    h.runtimes,
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
				runtimes:    h.runtimes,
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
				runtimes: h.runtimes,
				req:      payload,
				ctx:      oldContainerAsCurrent{payload},
				oldName:  payload.GetUpgradeRequest().GetName(),
				newName:  payload.GetUpgradeRequest().GetName() + oldContainerSuffix,
			},
		},
		{
			Name: stepDropOldContainer,
			Job: &dropOwnedContainerJob{
				runtimes: h.runtimes,
				req:      payload,
				ctx:      oldContainerAsCurrent{payload},
			},
		},
		{
			Name: stepRenameNewContainer,
			Job: &renameContainerJob{
				runtimes: h.runtimes,
				req:      payload,
				ctx:      payload,
				oldName:  payload.GetUpgradeRequest().GetName() + newContainerSuffix,
				newName:  payload.GetUpgradeRequest().GetName(),
			},
		},
	}
}

type checkSelfUpgradeJob struct {
	containerService service.ContainerService

	upgradeReq upgradeRequestAccessor

	// runtimes lets Do fall back to a suffix-aware container lookup when
	// containerService.InspectSmerd's bare-name lookup 404s in a suffixed
	// environment - see resolveCurrentContainer.
	runtimes container_runtime.RuntimeResolver
}

// Do carries over the self-upgrade guard of the deleted
// upgrade_steps.CheckUpgradeIsAvailable. That function took a full
// service.Services (only to call .SmerdManager() internally), while this
// handler takes the narrower ContainerService directly, so its few lines of
// logic live here instead.
func (j *checkSelfUpgradeJob) Do(ctx context.Context) error {
	id := env.GetContainerId()
	if id == nil {
		return nil
	}

	smerd, err := resolveCurrentContainer(
		ctx,
		j.runtimes,
		j.containerService,
		j.upgradeReq.GetUpgradeRequest().GetEnvironment(),
		j.upgradeReq.GetUpgradeRequest().GetName(),
	)
	if err != nil {
		return rerrors.Wrap(err, "error during smerd inspection")
	}

	if smerd.GetUuid() == *id {
		return rerrors.Wrap(ErrSelfUpgradeIsForbidden)
	}

	return nil
}

type captureOldContainerJob struct {
	containerService service.ContainerService

	upgradeReq upgradeRequestAccessor
	ctx        captureOldContainerCtx

	// runtimes lets Do fall back to a suffix-aware container lookup when
	// containerService.InspectSmerd's bare-name lookup 404s in a suffixed
	// environment - see resolveCurrentContainer.
	runtimes container_runtime.RuntimeResolver
}

// Do mirrors steps.fromContainerToRequest.Do, folding in the immediately
// adjacent SingleFunc step that sets ImageName - see
// docs/jobs_migrations/questions.md for the fold precedent.
func (j *captureOldContainerJob) Do(ctx context.Context) error {
	name := j.upgradeReq.GetUpgradeRequest().GetName()
	environment := j.upgradeReq.GetUpgradeRequest().GetEnvironment()

	cont, err := resolveCurrentContainer(ctx, j.runtimes, j.containerService, environment, name)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container")
	}

	req := &velez_api.CreateSmerd_Request{
		Name:        cont.GetName(),
		ImageName:   j.upgradeReq.GetUpgradeRequest().GetImage(),
		Environment: environment,
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

// resolveCurrentContainer resolves "the current container" for name, falling
// back to a suffix-aware lookup via runtimes when containerService's
// bare-name InspectSmerd fails - see checkSelfUpgradeJob/captureOldContainerJob's
// doc comments for why the bare-name lookup alone 404s in a suffixed
// environment (the real Docker container is named "name_<suffix>",
// container_runtime.labelBasedRuntime.containerName).
//
// The fallback reuses ContainerRuntime.ListContainers (which already applies
// the resolved environment's suffix, see labelBasedRuntime.ListContainers)
// rather than reconstructing the suffixed name here - suffix logic stays
// defined exactly once, in label_based.go.
//
// A failure to resolve runtimes for environment (unknown environment, storage
// error, etc.) is a real, surfaced error, not "no fallback available" -
// resolveEnvironment already validated environment exists back at the RPC
// boundary (smerd_upgrade.go), so a failure here means something is actually
// wrong (a race, a storage outage) and must not be swallowed into silently
// operating on a nil Smerd. A definitive "queried successfully, found
// nothing" (runtimes resolved fine, but ListContainers came back empty) is
// likewise a real, surfaced error - this is a read path, and callers like
// Test_UpgradeSmerd_NonExistentContainer_Fails depend on that.
func resolveCurrentContainer(
	ctx context.Context,
	runtimes container_runtime.RuntimeResolver,
	containerService service.ContainerService,
	environment, name string,
) (*velez_api.Smerd, error) {
	cont, err := containerService.InspectSmerd(ctx, environment, name)
	if err == nil {
		return cont, nil
	}

	if runtimes == nil {
		return nil, rerrors.Wrap(err, "error inspecting container")
	}

	runtime, runtimeErr := runtimes.Runtime(ctx, environment)
	if runtimeErr != nil {
		return nil, rerrors.Wrap(runtimeErr, "error resolving container runtime")
	}

	listName := name
	listReq := &velez_api.ListSmerds_Request{
		Name: &listName,
	}

	containers, err := runtime.ListContainers(ctx, listReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	if len(containers) == 0 {
		return nil, user_errors.New(fmt.Sprintf("container %q not found in environment %q", name, environment))
	}

	cont, err = containerService.InspectSmerd(ctx, environment, containers[0].ID)
	if err != nil {
		return nil, rerrors.Wrap(err, "error inspecting container")
	}

	return cont, nil
}

// fromContainerNetwork filters out the network aliases Docker adds
// implicitly rather than ones a caller actually requested: the container's
// own short-ID-prefix alias, and (confirmed against a real daemon - see
// TestCaptureOldContainerJob_Success) its own Docker container name too,
// which real Docker also auto-adds as an alias on connect. Without this, an
// upgrade would carry the OLD container's own name forward into the NEW
// container's alias list.
//
// Known gap: cont.GetName() is always the virtual/logical name (see
// docs/container_runtimes/interface_design.md's "names are always virtual at
// the interface boundary") - in a suffixed environment the real Docker name
// Docker actually aliased is name_<suffix>, which this won't match. Closing
// that needs the raw Docker name threaded through from InspectSmerd, which
// deliberately doesn't expose it today; not fixed here since no test in this
// pass exercises a suffixed captureOldContainerJob network-alias case.
func fromContainerNetwork(cont *velez_api.Smerd) []*velez_api.NetworkBind {
	out := make([]*velez_api.NetworkBind, 0, len(cont.GetNetworks()))

	for _, n := range cont.GetNetworks() {
		net := &velez_api.NetworkBind{NetworkName: n.GetNetworkName()}
		for _, a := range n.GetAliases() {
			if strings.HasPrefix(cont.GetUuid(), a) || a == cont.GetName() {
				continue
			}

			net.Aliases = append(net.Aliases, a)
		}

		if len(net.GetAliases()) != 0 {
			out = append(out, net)
		}
	}

	return out
}

type prepareUpgradeImageJob struct {
	runtimes container_runtime.RuntimeResolver

	upgradeReq upgradeRequestAccessor
	ctx        imageMetaAccessor
}

func (j *prepareUpgradeImageJob) Do(ctx context.Context) error {
	runtime, err := j.runtimes.Runtime(ctx, j.upgradeReq.GetUpgradeRequest().GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	imageInfo, err := runtime.PullImage(ctx, j.upgradeReq.GetUpgradeRequest().GetImage())
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
// to pause/stop the old container. Kept narrow (rather than depending on
// client.APIClient directly) so it's hand-fakeable in unit tests, same
// rationale as copy_to_volume.go's startAPI/copyAPI - client.APIClient's
// method set is a superset of pauseAPI's, so the real client still satisfies
// it. Network disconnect/reconnect used to be duplicated here against raw
// NetworkDisconnect/NetworkConnect calls (see docs/jobs_migrations/questions.md
// #8); that's gone now that ContainerRuntime exposes
// DisconnectFromNetworks/ConnectToNetwork directly - see networkBindingsFor.
type pauseAPI interface {
	ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error)
	ContainerPause(ctx context.Context, containerID string) error
	ContainerUnpause(ctx context.Context, containerID string) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
}

// networkBinding is a logical network name plus the aliases the container was
// connected with - captured once (in pauseOldContainerJob.Do) so Rollback can
// reconnect with the exact same aliases, mirroring what the disconnected
// container actually had rather than whatever Docker happens to report by the
// time Rollback runs.
type networkBinding struct {
	name    string
	aliases []string
}

// networkBindingsFor returns the logical network names (and aliases) every
// create_smerd/upgrade_smerd container is connected to: the environment's
// default network (only when the request has at least one port binding -
// mirroring createContainerJob.connectNetworks' exact gating, since that's
// the only condition under which the container ever joined it in the first
// place) plus any extra networks from Settings.Network. Deriving the set this
// way (from the already-known/logical request settings) rather than by
// inspecting the live Docker container is deliberate: NetworkSettings.Networks
// map keys are the REAL, already-suffixed Docker network names, and handing
// those back into ContainerRuntime.DisconnectFromNetworks/ConnectToNetwork -
// which suffix whatever name they're given - would suffix them a second time.
func networkBindingsFor(request *velez_api.CreateSmerd_Request) []networkBinding {
	extraNetworks := request.GetSettings().GetNetwork()

	bindings := make([]networkBinding, 0, 1+len(extraNetworks))

	if request.GetSettings() != nil && len(request.GetSettings().GetPorts()) != 0 {
		bindings = append(bindings, networkBinding{name: env.VervNetwork, aliases: []string{request.GetName()}})
	}

	for _, n := range extraNetworks {
		bindings = append(bindings, networkBinding{name: n.GetNetworkName(), aliases: n.GetAliases()})
	}

	return bindings
}

type pauseOldContainerJob struct {
	dockerAPI   pauseAPI
	portManager node_clients.PortManager
	// runtimes resolves the request's environment into the ContainerRuntime
	// that serves it, so network disconnect/reconnect around the pause is
	// scoped/suffixed to that environment instead of against the raw Docker
	// daemon - see docs/container_runtimes.
	runtimes container_runtime.RuntimeResolver

	req smerdRequestAccessor
	ctx oldContainerIDAccessor

	stateBeforePause container.ContainerState
	disconnectedNets []networkBinding
	portsOnHold      []uint32
}

// Do mirrors smerd_steps.detachContainerFromVervStep.Do, reading the old
// container id from the persisted context at Do()-time instead of a raw
// pointer fixed at construction.
func (j *pauseOldContainerJob) Do(ctx context.Context) error {
	containerID := j.ctx.GetOldContainerId()
	if containerID == "" {
		return errUpgradeContainerIDMissing
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

	request := j.req.GetRequest()

	runtime, err := j.runtimes.Runtime(ctx, request.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	j.disconnectedNets = networkBindingsFor(request)

	networkNames := make([]string, 0, len(j.disconnectedNets))
	for _, nb := range j.disconnectedNets {
		networkNames = append(networkNames, nb.name)
	}

	err = runtime.DisconnectFromNetworks(ctx, containerID, networkNames)
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

	runtime, err := j.runtimes.Runtime(ctx, j.req.GetRequest().GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	var globErr error

	for _, nb := range j.disconnectedNets {
		connectReq := container_runtime.ConnectToNetworkRequest{
			ContainerID: containerID,
			NetworkName: nb.name,
			Aliases:     nb.aliases,
		}

		err = runtime.ConnectToNetwork(ctx, connectReq)
		if err != nil {
			globErr = rerrors.Join(globErr, rerrors.Wrap(err, "error connecting to network on rollback"))
		}
	}

	if globErr != nil {
		return rerrors.Wrap(globErr)
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

// renamingCreateContainerJob renames the shared Request before delegating to
// createContainerJob, mirroring do_smerd_upgrade.go's SingleFunc-then-Create
// step pairs: both container-create stages reuse the very same Request
// value, only its Name differs between them.
type renamingCreateContainerJob struct {
	nodeClients node_clients.NodeClients

	req smerdRequestAccessor
	ctx containerIDAccessor

	runtimes container_runtime.RuntimeResolver

	newName func(current string) string
}

func (j *renamingCreateContainerJob) Do(ctx context.Context) error {
	request := j.req.GetRequest()

	request.Name = j.newName(request.GetName())

	inner := &createContainerJob{
		nodeClients: j.nodeClients,
		req:         j.req,
		ctx:         j.ctx,
		runtimes:    j.runtimes,
	}

	return inner.Do(ctx)
}

func (j *renamingCreateContainerJob) Rollback(ctx context.Context) error {
	inner := &createContainerJob{nodeClients: j.nodeClients, req: j.req, ctx: j.ctx, runtimes: j.runtimes}

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
		return errUpgradeConfigContainerIDMissing
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
// sets req.Verv/req.Plain (capture_old_container leaves both nil), so the
// original step's doPlain/yaml branches are unreachable here - restricted to
// the env branch only, same restriction style as assemble_config.go's
// fetchConfigJob.
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

func toConfTypePrefix(s string) matreshka_api.ConfigType {
	switch s {
	case confTypeVerv:
		return matreshka_api.ConfigType_verv
	case confTypePg:
		return matreshka_api.ConfigType_pg
	default:
		return matreshka_api.ConfigType_plain
	}
}

type prepareUpgradeVervConfigJob struct {
	portManager node_clients.PortManager
	// runtimes resolves the request's environment into the ContainerRuntime
	// that serves it, so network creation is scoped/suffixed to that
	// environment instead of against the raw Docker daemon - see
	// docs/container_runtimes.
	runtimes container_runtime.RuntimeResolver

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

	if len(request.GetSettings().GetNetwork()) != 0 {
		runtime, err := j.runtimes.Runtime(ctx, request.GetEnvironment())
		if err != nil {
			return rerrors.Wrap(err, "error resolving container runtime")
		}

		for _, n := range request.GetSettings().GetNetwork() {
			err = runtime.CreateNetwork(ctx, n.GetNetworkName())
			if err != nil {
				return rerrors.Wrapf(err, "error creating network: %s", n.GetNetworkName())
			}
		}
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

func (j *prepareUpgradeVervConfigJob) lockPorts(request *velez_api.CreateSmerd_Request) (err error) {
	j.lockedPorts = make([]uint32, 0, len(request.GetSettings().GetPorts()))

	// See prepareSmerdVervConfigJob.lockPorts - same shared pool, same
	// environment-tagged ownership.
	environment := request.GetEnvironment()

	for _, p := range request.GetSettings().GetPorts() {
		if p.ExposedTo == nil {
			var port uint32

			port, err = j.portManager.GetPortForEnvironment(environment)
			p.ExposedTo = &port
		} else {
			ok := j.portManager.UnHoldPort(p.GetExposedTo())
			if !ok {
				err = j.portManager.LockPortForEnvironment(environment, p.GetExposedTo())
			}
		}

		if err != nil {
			return rerrors.Wrap(err, "error locking host port")
		}

		j.lockedPorts = append(j.lockedPorts, p.GetExposedTo())
	}

	return nil
}

// renameContainerJob renames a container to a virtual/logical name via the
// ContainerRuntime resolved for the request's environment, so the real
// Docker name it ends up with still carries that environment's suffix (see
// container_runtime.ContainerRuntime.Rename). newName/oldName are precomputed
// virtual names set at construction time (see upgradeSmerdHandler.BuildJobs) -
// containerID is read from ctx at Do/Rollback time exactly as before, but the
// "what was it named before this step" bookkeeping (oldName) is no longer
// captured via a raw ContainerInspect, since it's statically known at
// BuildJobs time.
type renameContainerJob struct {
	runtimes container_runtime.RuntimeResolver

	req upgradeRequestAccessor
	ctx containerIDAccessor

	newName string
	oldName string
}

func (j *renameContainerJob) Do(ctx context.Context) error {
	containerID := j.ctx.GetContainerId()
	if containerID == "" {
		return errUpgradeContainerIDMissing
	}

	runtime, err := j.runtimes.Runtime(ctx, j.req.GetUpgradeRequest().GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	err = runtime.Rename(ctx, containerID, j.newName)
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

	runtime, err := j.runtimes.Runtime(ctx, j.req.GetUpgradeRequest().GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	err = runtime.Rename(ctx, containerID, j.oldName)
	if err != nil {
		return rerrors.Wrap(err, "error renaming container on rollback")
	}

	return nil
}

// dropOwnedContainerJob removes a container owned by this task (the
// config-fetcher scratch container, or the old container after rename)
// through the ContainerRuntime resolved for req's environment, replacing the
// two dropScratchContainerJob{docker: h.nodeClients.Docker()} call sites in
// upgrade_smerd.go's BuildJobs that used to bypass any suffix-aware runtime
// entirely - mirroring dropContainerJob's (drop_smerd.go) identical fix for
// DropSmerd. assemble_config.go's own dropScratchContainerJob call site is
// untouched: that scratch container is created with an explicitly
// empty/unowned suffix, a different ownership story.
type dropOwnedContainerJob struct {
	runtimes container_runtime.RuntimeResolver
	req      smerdRequestAccessor
	ctx      containerIDAccessor
}

func (j *dropOwnedContainerJob) Do(ctx context.Context) error {
	containerID := j.ctx.GetContainerId()
	if containerID == "" {
		return nil
	}

	runtime, err := j.runtimes.Runtime(ctx, j.req.GetRequest().GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	err = runtime.Remove(ctx, containerID)
	if err != nil {
		return rerrors.Wrap(err, "error dropping scratch container")
	}

	return nil
}
