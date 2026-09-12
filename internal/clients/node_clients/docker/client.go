package docker

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox/closer"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

// Docker is environment-agnostic on purpose: the container suffix that scopes
// containers to an environment used to be baked in here at construction time
// (one Velez process = one environment). It's now passed explicitly per call
// by whoever resolved the caller's environment - see ListContainers and
// ContainerCreate.
type Docker struct {
	directApi   client.APIClient
	bakedLabels []string
	// host is cli.DaemonHost() at construction time - the resolved address
	// this connection actually talks to. See Host().
	host string
}

func NewClient(bakedLabels []string) (*Docker, error) {
	return NewClientWithOpts(bakedLabels, client.FromEnv, client.WithAPIVersionNegotiation())
}

// NewClientWithOpts is NewClient's variant for callers that need control over
// the underlying SDK client's construction - e.g. tests pointing at a
// specific daemon/socket instead of the ambient DOCKER_HOST. NewClient itself
// is unchanged and remains the production path.
func NewClientWithOpts(bakedLabels []string, opts ...client.Opt) (*Docker, error) {
	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting docker client")
	}

	closer.Add(cli.Close)

	return &Docker{
		directApi:   cli,
		bakedLabels: bakedLabels,
		host:        cli.DaemonHost(),
	}, nil
}

// Host returns the Docker daemon address this connection resolved to - see
// the Docker interface's Host doc comment.
func (d *Docker) Host() string {
	return d.host
}

func (d *Docker) PullImage(ctx context.Context, imageName string) (image.InspectResponse, error) {
	_, err := dockerutils.PullImage(ctx, d.directApi, imageName, false)
	if err != nil {
		return image.InspectResponse{}, rerrors.Wrap(err, "error pulling image")
	}

	img, err := d.directApi.ImageInspect(ctx, imageName)
	if err != nil {
		return image.InspectResponse{}, rerrors.Wrap(err, "error inspecting image")
	}

	return img, nil
}

func (d *Docker) Remove(ctx context.Context, contUUID string) error {
	roReq := container.RemoveOptions{
		Force: true,
	}

	err := d.directApi.ContainerRemove(ctx, contUUID, roReq)
	if err != nil {
		if strings.Contains(err.Error(), NoSuchContainerError) {
			return nil
		}

		return rerrors.Wrap(err, "error removing container")
	}

	return nil
}

func (d *Docker) Stop(ctx context.Context, nameOrId string) error {
	stopOpts := container.StopOptions{}

	err := d.directApi.ContainerStop(ctx, nameOrId, stopOpts)
	if err != nil {
		if strings.Contains(err.Error(), NoSuchContainerError) {
			return nil
		}

		return rerrors.Wrap(err, "error stopping container")
	}

	return nil
}

func (d *Docker) Restart(ctx context.Context, nameOrId string) error {
	restartOpts := container.StopOptions{}

	err := d.directApi.ContainerRestart(ctx, nameOrId, restartOpts)
	if err != nil {
		if strings.Contains(err.Error(), NoSuchContainerError) {
			return nil
		}

		return rerrors.Wrap(err, "error restarting container")
	}

	return nil
}

func (d *Docker) IsContainerRunning(ctx context.Context, nameOrId string) (bool, bool, error) {
	resp, err := d.directApi.ContainerInspect(ctx, nameOrId)
	if err != nil {
		if strings.Contains(err.Error(), NoSuchContainerError) {
			return false, false, nil
		}

		return false, false, rerrors.Wrap(err, "error inspecting container")
	}

	return resp.State.Running, true, nil
}

func (d *Docker) ListOccupiedPorts(ctx context.Context) ([]uint32, error) {
	containerList, err := d.directApi.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	usedPorts := make([]uint32, 0)

	for _, c := range containerList {
		for _, p := range c.Ports {
			if p.PublicPort != 0 {
				usedPorts = append(usedPorts, uint32(p.PublicPort))
			}
		}
	}

	return usedPorts, nil
}

// ListContainers lists containers, optionally scoped to a single environment.
//
// suffix is the environment's resolved labels.SuffixLabel value. An empty
// suffix means "don't scope by environment" - which is what internal,
// node-wide callers (local_storage's docker-backed views) want.
func (d *Docker) ListContainers(
	ctx context.Context,
	req *velez_api.ListSmerds_Request,
	suffix string,
) ([]container.Summary, error) {
	if req.Label == nil {
		req.Label = map[string]string{}
	}

	if suffix != "" {
		req.Label[labels.SuffixLabel] = suffix
	}

	list, err := dockerutils.ListContainers(ctx, d.directApi, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	return list, nil
}

func (d *Docker) Exec(ctx context.Context, containerId string, execCfg container.ExecOptions) ([]byte, error) {
	execResp, err := d.directApi.ContainerExecCreate(ctx, containerId, execCfg)
	if err != nil {
		return nil, rerrors.Wrap(err, "error calling exec create on container via docker api")
	}

	if !execCfg.AttachStdout && !execCfg.AttachStderr {
		return nil, nil
	}

	// Attach to execution
	attachResp, err := d.directApi.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return nil, rerrors.Wrap(err, "error calling exec attach on container via docker api")
	}
	defer attachResp.Close()

	dataOut := bytes.NewBuffer(nil)

	_, err = io.Copy(dataOut, attachResp.Reader)
	if err != nil {
		return nil, rerrors.Wrap(err, "error during copying bytes from attached to container exec")
	}

	return asciiSymbolsOnly(dataOut.Bytes()), nil
}

func (d *Docker) Client() client.APIClient {
	return d.directApi
}

func asciiSymbolsOnly(in []byte) []byte {
	cleanBuff := bytes.NewBuffer(nil)

	for _, b := range in {
		if b >= 32 && b <= 127 || b == '\n' {
			cleanBuff.WriteByte(b)
		}
	}

	return cleanBuff.Bytes()
}

func (d *Docker) ContainerCreate(
	ctx context.Context,
	config *container.Config,
	hostConfig *container.HostConfig,
	networkingConfig *network.NetworkingConfig,
	platform *v1.Platform,
	containerName string,
	suffix string,
) (container.CreateResponse, error) {
	if config.Labels == nil {
		config.Labels = map[string]string{}
	}

	config.Labels[labels.CreatedWithVelezLabel] = "true"
	// suffix is the resolved environment suffix, passed in by whoever knows
	// which environment this container belongs to. Empty means "the node's
	// unscoped/default environment", preserving pre-multi-environment behavior.
	config.Labels[labels.SuffixLabel] = suffix

	for _, label := range d.bakedLabels {
		before, after, ok := strings.Cut(label, "=")
		name := label

		var val string

		if ok {
			val = after
			name = before
		}

		config.Labels[name] = val
	}

	createResponse, err := d.directApi.ContainerCreate(ctx, config, hostConfig, networkingConfig, platform, containerName)
	if err != nil {
		if errdefs.IsConflict(err) {
			return container.CreateResponse{}, HandleConflictMessage(err)
		}

		return container.CreateResponse{}, rerrors.Wrap(err, "error during container creation via docker api")
	}

	return createResponse, nil
}

// Stats delegates to dockerutils.Stats, shared with
// container_runtime.labelBasedRuntime.Stats so the CPU-percent/memory
// computation is defined in exactly one place.
func (d *Docker) Stats(ctx context.Context, nameOrId string) (domain.ContainerStats, error) {
	stats, err := dockerutils.Stats(ctx, d.directApi, nameOrId)
	if err != nil {
		return domain.ContainerStats{}, rerrors.Wrap(err, "error getting container stats")
	}

	return stats, nil
}
