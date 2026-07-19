package jobs

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/evon"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/parser"
	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/utils/configutils"
)

const CreateSmerdAction = "create_smerd"

// Accessor interfaces the create_smerd jobs need from their TaskContext.
// *velez_api.CreateSmerdTaskPayload satisfies all of them, but any other
// task payload with the same fields could reuse these jobs too.

type smerdRequestAccessor interface {
	GetRequest() *velez_api.CreateSmerd_Request
}

type imageIdAccessor interface {
	GetImageId() string
	SetImageId(string)
}

type containerIdAccessor interface {
	GetContainerId() string
	SetContainerId(string)
}

// createSmerdImageAccessor is what prepare_image needs to persist after
// pulling/inspecting the image - imageMetaAccessor (labels/tags) is declared
// in assemble_config.go and reused here as-is.
type createSmerdImageAccessor interface {
	imageIdAccessor
	imageMetaAccessor
	SetImageExposedPorts([]string)
}

type imageExposedPortsAccessor interface {
	GetImageExposedPorts() []string
}

type pathToFilesAccessor interface {
	GetPathToFiles() map[string][]byte
	SetPathToFiles(map[string][]byte)
}

type createSmerdHandler struct {
	nodeClients   node_clients.NodeClients
	configService service.ConfigurationService
}

// NewCreateSmerdHandler builds the TaskHandler for the "create_smerd" action.
func NewCreateSmerdHandler(nodeClients node_clients.NodeClients, configService service.ConfigurationService) TaskHandler {
	return &createSmerdHandler{
		nodeClients:   nodeClients,
		configService: configService,
	}
}

func (h *createSmerdHandler) Action() string {
	return CreateSmerdAction
}

func (h *createSmerdHandler) NewContext() TaskContext {
	return &velez_api.CreateSmerdTaskPayload{}
}

func (h *createSmerdHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload := taskCtx.(*velez_api.CreateSmerdTaskPayload)
	dockerAPI := h.nodeClients.Docker().Client()

	return []NamedJob{
		{
			Name: "prepare_request",
			Job: &prepareSmerdRequestJob{
				req: payload,
			},
		},
		{
			Name: "prepare_image",
			Job: &prepareImageJob{
				docker: h.nodeClients.Docker(),
				req:    payload,
				ctx:    payload,
			},
		},
		{
			Name: "fetch_config",
			Job: &fetchSmerdConfigJob{
				configService: h.configService,
				req:           payload,
				imageMeta:     payload,
				mounts:        payload,
			},
		},
		{
			Name: "prepare_verv_config",
			Job: &prepareSmerdVervConfigJob{
				dockerAPI:         dockerAPI,
				portManager:       h.nodeClients.PortManager(),
				req:               payload,
				imageMeta:         payload,
				imageExposedPorts: payload,
			},
		},
		{
			Name: "create_container",
			Job: &createContainerJob{
				nodeClients: h.nodeClients,
				req:         payload,
				ctx:         payload,
			},
		},
		{
			Name: "copy_to_container",
			Job: &copyToContainerJob{
				dockerAPI: dockerAPI,
				ctx:       payload,
				mounts:    payload,
			},
		},
		{
			Name: "start_container",
			Job: &startContainerJob{
				dockerAPI: dockerAPI,
				ctx:       payload,
			},
		},
		{
			Name: "healthcheck",
			Job: &healthcheckJob{
				dockerAPI: dockerAPI,
				req:       payload,
				ctx:       payload,
			},
		},
		{
			Name: "subscribe_for_config_changes",
			Job: &subscribeSmerdConfigChangesJob{
				configService: h.configService,
				req:           payload,
			},
		},
	}
}

// prepareSmerdRequestJob mirrors steps.PrepareCreateRequest: fills in the
// defaults every later job assumes are non-nil, and defaults Config to a
// Verv spec named after the request when the caller didn't supply one.
type prepareSmerdRequestJob struct {
	req smerdRequestAccessor
}

func (j *prepareSmerdRequestJob) Do(_ context.Context) error {
	request := j.req.GetRequest()

	if request.Settings == nil {
		request.Settings = &velez_api.Container_Settings{}
	}

	if request.Hardware == nil {
		request.Hardware = &velez_api.Container_Hardware{}
	}

	if request.Env == nil {
		request.Env = make(map[string]string)
	}

	if request.Labels == nil {
		request.Labels = make(map[string]string)
	}

	if request.Config == nil {
		request.Config = &velez_api.CreateSmerd_Request_Verv{
			Verv: &velez_api.MatreshkaConfigSpec{
				ConfigName: &request.Name,
			},
		}
	}

	return nil
}

type prepareImageJob struct {
	docker node_clients.Docker

	req smerdRequestAccessor
	ctx createSmerdImageAccessor
}

func (j *prepareImageJob) Do(ctx context.Context) error {
	imageInfo, err := j.docker.PullImage(ctx, j.req.GetRequest().GetImageName())
	if err != nil {
		return rerrors.Wrap(err, "error pulling image")
	}

	j.ctx.SetImageId(imageInfo.ID)

	var imageLabels map[string]string

	var exposedPorts []string

	if imageInfo.Config != nil {
		imageLabels = imageInfo.Config.Labels

		exposedPorts = make([]string, 0, len(imageInfo.Config.ExposedPorts))
		for port := range imageInfo.Config.ExposedPorts {
			exposedPorts = append(exposedPorts, string(port))
		}
	}

	j.ctx.SetImageLabels(imageLabels)
	j.ctx.SetImageTags(imageInfo.RepoTags)
	j.ctx.SetImageExposedPorts(exposedPorts)

	return nil
}

// fetchSmerdConfigJob mirrors config_steps.fetchConfigStep, restricted to
// the branches actually reachable once prepare_request has run: Config is
// always non-nil (Verv or Plain), so the original step's "no config at all"
// error path is unreachable in practice - kept anyway for parity. Reuses
// classifyImage/confType constants from assemble_config.go rather than
// re-deriving fillMeta's image-classification rules.
type fetchSmerdConfigJob struct {
	configService service.ConfigurationService

	req       smerdRequestAccessor
	imageMeta imageMetaAccessor
	mounts    pathToFilesAccessor
}

func (j *fetchSmerdConfigJob) Do(ctx context.Context) error {
	request := j.req.GetRequest()

	vervCfg := request.GetVerv()
	if vervCfg != nil {
		return j.doVerv(ctx, request, vervCfg)
	}

	plainCfg := request.GetPlain()
	if plainCfg != nil {
		j.doPlain(plainCfg)

		return nil
	}

	return rerrors.New("only verv config supported for now")
}

func (j *fetchSmerdConfigJob) doVerv(
	ctx context.Context,
	request *velez_api.CreateSmerd_Request,
	spec *velez_api.MatreshkaConfigSpec,
) error {
	if request.IgnoreConfig {
		return nil
	}

	confType, format, defaultSystemPath := classifyImage(j.imageMeta.GetImageLabels(), j.imageMeta.GetImageTags())
	if spec.ConfigFormat != nil {
		format = *spec.ConfigFormat
	}

	systemPath := defaultSystemPath
	if spec.SystemPath != nil {
		systemPath = *spec.SystemPath
	}

	configName := request.GetName()
	if spec.ConfigName != nil {
		configName = *spec.ConfigName
	}

	confTypePrefix := toConfTypePrefix(confType)
	configName = configutils.AppendPrefix(confTypePrefix, configName)

	meta := domain.ConfigMeta{
		Name:     configName,
		Version:  spec.ConfigVersion,
		ConfType: confTypePrefix,
		Format:   format,
	}

	if format == velez_api.ConfigFormat_env {
		return j.setEnv(ctx, request, meta)
	}

	return j.getPlain(ctx, meta, systemPath)
}

func (j *fetchSmerdConfigJob) setEnv(
	ctx context.Context,
	request *velez_api.CreateSmerd_Request,
	meta domain.ConfigMeta,
) error {
	envEvon, err := j.configService.GetEnvFromApi(ctx, meta)
	if err != nil {
		if status.Code(err) != codes.NotFound {
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

func (j *fetchSmerdConfigJob) getPlain(ctx context.Context, meta domain.ConfigMeta, systemPath string) error {
	if systemPath == "" {
		return nil
	}

	content, err := j.configService.GetPlainFromApi(ctx, meta)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return rerrors.Wrap(err, "error getting matreshka config from matreshka api")
		}
	}

	j.addMount(systemPath, content)

	return nil
}

func (j *fetchSmerdConfigJob) doPlain(spec *velez_api.PlainConfigSpec) {
	for path, content := range spec.GetConfigs() {
		j.addMount(path, content)
	}
}

func (j *fetchSmerdConfigJob) addMount(path string, content []byte) {
	mounts := j.mounts.GetPathToFiles()
	if mounts == nil {
		mounts = make(map[string][]byte)
	}

	mounts[path] = content
	j.mounts.SetPathToFiles(mounts)
}

// prepareSmerdVervConfigJob mirrors steps.prepareVervConfig.
type prepareSmerdVervConfigJob struct {
	dockerAPI   createNetworkAPI
	portManager node_clients.PortManager

	req               smerdRequestAccessor
	imageMeta         imageMetaAccessor
	imageExposedPorts imageExposedPortsAccessor

	lockedPorts []uint32
}

func (j *prepareSmerdVervConfigJob) Do(ctx context.Context) error {
	request := j.req.GetRequest()

	if request.UseImagePorts {
		err := j.getPortsFromImage(request)
		if err != nil {
			return rerrors.Wrap(err, "error locking ports")
		}
	}

	err := j.lockPorts(request)
	if err != nil {
		return rerrors.Wrap(err, "error locking ports")
	}

	if !request.IgnoreConfig {
		request.Env[matreshka.VervName] = request.GetName()
	}

	j.applyLabels(request)

	err = j.ensureNetworks(ctx, request)
	if err != nil {
		return err
	}

	return nil
}

func (j *prepareSmerdVervConfigJob) applyLabels(request *velez_api.CreateSmerd_Request) {
	for name, val := range j.imageMeta.GetImageLabels() {
		if request.IgnoreConfig && name == labels.MatreshkaConfigLabel && val == vervConfigLabelEnabled {
			val = vervConfigLabelDisabled
		}

		request.Labels[name] = val
	}

	request.Labels[labels.ComposeGroupLabel] = request.GetName()
	if request.AutoUpgrade {
		request.Labels[labels.AutoUpgrade] = vervConfigLabelEnabled
	}
}

func (j *prepareSmerdVervConfigJob) ensureNetworks(ctx context.Context, request *velez_api.CreateSmerd_Request) error {
	for _, n := range request.Settings.Network {
		err := createNetworkIfMissing(ctx, j.dockerAPI, n.NetworkName)
		if err != nil {
			return rerrors.Wrap(err, "error creating network: %s", n.NetworkName)
		}
	}

	return nil
}

const (
	vervConfigLabelEnabled  = "true"
	vervConfigLabelDisabled = "false"
)

func (j *prepareSmerdVervConfigJob) getPortsFromImage(request *velez_api.CreateSmerd_Request) error {
	portsFromReq := map[uint32]*velez_api.Port{}
	for _, port := range request.Settings.Ports {
		portsFromReq[port.ServicePortNumber] = port
	}

	for _, portProtoc := range j.imageExposedPorts.GetImageExposedPorts() {
		portParts := strings.Split(portProtoc, "/")

		portVal, err := strconv.ParseUint(portParts[0], 10, 32)
		if err != nil {
			return rerrors.Wrap(err, "error parsing port")
		}

		_, ok := portsFromReq[uint32(portVal)]
		if ok {
			continue
		}

		portBind := &velez_api.Port{
			ServicePortNumber: uint32(portVal),
			Protocol:          velez_api.Port_Protocol(velez_api.Port_Protocol_value[portParts[1]]),
		}

		request.Settings.Ports = append(request.Settings.Ports, portBind)
		portsFromReq[uint32(portVal)] = portBind
	}

	return nil
}

func (j *prepareSmerdVervConfigJob) lockPorts(request *velez_api.CreateSmerd_Request) (err error) {
	j.lockedPorts = make([]uint32, 0, len(request.Settings.Ports))

	for _, imagePort := range request.Settings.Ports {
		if imagePort.ExposedTo == nil {
			var port uint32

			port, err = j.portManager.GetPort()
			imagePort.ExposedTo = &port
		} else {
			ok := j.portManager.UnHoldPort(*imagePort.ExposedTo)
			if !ok {
				err = j.portManager.LockPort(*imagePort.ExposedTo)
			}
		}

		if err != nil {
			return rerrors.Wrap(err, "error locking host port")
		}

		j.lockedPorts = append(j.lockedPorts, *imagePort.ExposedTo)
	}

	return nil
}

func (j *prepareSmerdVervConfigJob) Rollback(_ context.Context) error {
	for _, port := range j.lockedPorts {
		if !j.portManager.UnHoldPort(port) {
			j.portManager.UnlockPorts(j.lockedPorts)
		}
	}

	return nil
}

// copyToContainerJob mirrors container_steps.CopyToContainer.
type copyToContainerJob struct {
	dockerAPI client.APIClient

	ctx    containerIdAccessor
	mounts pathToFilesAccessor
}

func (j *copyToContainerJob) Do(ctx context.Context) error {
	containerId := j.ctx.GetContainerId()

	for path, content := range j.mounts.GetPathToFiles() {
		if content == nil {
			continue
		}

		if containerId == "" {
			return rerrors.New("no container id provided")
		}

		err := dockerutils.WriteToContainer(ctx, j.dockerAPI, containerId, path, content)
		if err != nil {
			return rerrors.Wrap(err, "error copying to container")
		}
	}

	return nil
}

// subscribeSmerdConfigChangesJob mirrors config_steps.SubscribeForConfigChanges:
// a failed subscription is logged, not fatal to the task, same as the
// original step.
type subscribeSmerdConfigChangesJob struct {
	configService service.ConfigurationService
	req           smerdRequestAccessor
}

func (j *subscribeSmerdConfigChangesJob) Do(_ context.Context) error {
	name := j.req.GetRequest().GetName()

	err := j.configService.SubscribeOnChanges(name)
	if err != nil {
		log.Error().Err(rerrors.Wrap(err, "error handling subscription on service with name: "+name)).Send()
	}

	return nil
}

type createContainerJob struct {
	nodeClients node_clients.NodeClients

	req smerdRequestAccessor
	ctx containerIdAccessor
}

func (j *createContainerJob) Do(ctx context.Context) error {
	req := j.req.GetRequest()

	cfg := &container.Config{
		Image:       req.GetImageName(),
		Hostname:    req.GetName(),
		Cmd:         parser.FromCommand(req.Command),
		Healthcheck: parser.FromHealthcheck(req.Healthcheck),
		Env:         parser.FromDockerEnv(req.Env),
		Labels:      req.Labels,
	}

	hostCfg := &container.HostConfig{
		PortBindings:  parser.FromPorts(req.Settings),
		Mounts:        parser.FromVolume(req.Settings),
		RestartPolicy: parser.FromRestart(req.Restart),
	}
	if req.Settings != nil && len(req.Settings.Volumes) != 0 {
		hostCfg.VolumeDriver = req.Settings.Volumes[0].VolumeName
	}

	netCfg := &network.NetworkingConfig{}
	if req.Settings != nil && len(req.Settings.Ports) != 0 {
		netCfg.EndpointsConfig = make(map[string]*network.EndpointSettings)
		netCfg.EndpointsConfig[env.VervNetwork] = &network.EndpointSettings{
			Aliases: []string{req.GetName()},
		}
		// required in order to expose ports on some platforms (e.g. orbs)
		netCfg.EndpointsConfig["bridge"] = &network.EndpointSettings{}
	}

	dockerClient := j.nodeClients.Docker()

	created, err := dockerClient.ContainerCreate(ctx, cfg, hostCfg, netCfg, &v1.Platform{}, req.GetName())
	if err != nil {
		return rerrors.Wrap(err, "error creating container")
	}

	containerInfo, err := dockerClient.Client().ContainerInspect(ctx, created.ID)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container by id")
	}

	if req.Settings != nil {
		for _, n := range req.Settings.Network {
			connectReq := dockerutils.ConnectToNetworkRequest{
				NetworkName: n.NetworkName,
				ContId:      created.ID,
				Aliases:     n.Aliases,
			}

			err = dockerutils.ConnectToNetwork(ctx, dockerClient.Client(), connectReq)
			if err != nil {
				return rerrors.Wrap(err)
			}
		}
	}

	j.ctx.SetContainerId(containerInfo.ID)

	return nil
}

func (j *createContainerJob) Rollback(ctx context.Context) error {
	containerId := j.ctx.GetContainerId()
	if containerId == "" {
		return nil
	}

	err := j.nodeClients.Docker().Remove(ctx, containerId)
	if err != nil && !errdefs.IsNotFound(err) {
		return rerrors.Wrapf(err, "error removing container '%s'", containerId)
	}

	return nil
}

type startContainerJob struct {
	dockerAPI client.APIClient

	ctx containerIdAccessor
}

func (j *startContainerJob) Do(ctx context.Context) error {
	containerId := j.ctx.GetContainerId()
	if containerId == "" {
		return rerrors.New("no container id provided")
	}

	err := j.dockerAPI.ContainerStart(ctx, containerId, container.StartOptions{})
	if err != nil {
		return rerrors.Wrap(err, "error starting container")
	}

	return nil
}

func (j *startContainerJob) Rollback(ctx context.Context) error {
	containerId := j.ctx.GetContainerId()
	if containerId == "" {
		return nil
	}

	err := j.dockerAPI.ContainerStop(ctx, containerId, container.StopOptions{})
	if err != nil {
		return rerrors.Wrapf(err, "error stopping container '%s'", containerId)
	}

	return nil
}

type healthcheckJob struct {
	dockerAPI client.APIClient

	req smerdRequestAccessor
	ctx containerIdAccessor
}

func (j *healthcheckJob) Do(ctx context.Context) error {
	healthcheck := j.req.GetRequest().GetHealthcheck()
	if healthcheck == nil {
		return nil
	}

	containerId := j.ctx.GetContainerId()
	if containerId == "" {
		return rerrors.New("container was not created")
	}

	for i := uint32(0); i < healthcheck.GetRetries(); i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(healthcheck.GetIntervalSecond()) * time.Second):
		}

		cont, err := j.dockerAPI.ContainerInspect(ctx, containerId)
		if err != nil {
			return rerrors.Wrap(err, "error during healthcheck")
		}

		if cont.State.Status == dockerContainerStatusRunning {
			return nil
		}
	}

	return rerrors.New("healthcheck retries exhausted")
}

const dockerContainerStatusRunning = "running"
