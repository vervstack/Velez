package jobs

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/parser"
	"go.vervstack.ru/Velez/internal/cluster/env"
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

type createSmerdHandler struct {
	nodeClients node_clients.NodeClients
}

func NewCreateSmerdHandler(nodeClients node_clients.NodeClients) TaskHandler {
	return &createSmerdHandler{
		nodeClients: nodeClients,
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

	return []NamedJob{
		{
			Name: "prepare_image",
			Job: &prepareImageJob{
				docker: h.nodeClients.Docker(),
				req:    payload,
				ctx:    payload,
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
			Name: "start_container",
			Job: &startContainerJob{
				dockerAPI: h.nodeClients.Docker().Client(),
				ctx:       payload,
			},
		},
		{
			Name: "healthcheck",
			Job: &healthcheckJob{
				dockerAPI: h.nodeClients.Docker().Client(),
				req:       payload,
				ctx:       payload,
			},
		},
	}
}

type prepareImageJob struct {
	docker node_clients.Docker

	req smerdRequestAccessor
	ctx imageIdAccessor
}

func (j *prepareImageJob) Do(ctx context.Context) error {
	imageInfo, err := j.docker.PullImage(ctx, j.req.GetRequest().GetImageName())
	if err != nil {
		return rerrors.Wrap(err, "error pulling image")
	}

	j.ctx.SetImageId(imageInfo.ID)

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
