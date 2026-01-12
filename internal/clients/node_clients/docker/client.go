package docker

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox/closer"

	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

type Docker struct {
	directApi client.APIClient
}

func NewClient() (*Docker, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting docker client")
	}

	closer.Add(cli.Close)

	return &Docker{
		directApi: cli,
	}, nil
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
		if !strings.Contains(err.Error(), NoSuchContainerError) {
			return nil
		}
		return rerrors.Wrap(err, "error removing container")
	}

	return nil
}

func (d *Docker) ListContainers(ctx context.Context, req *velez_api.ListSmerds_Request) ([]container.Summary, error) {
	list, err := dockerutils.ListContainers(ctx, d.directApi, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	return list, nil
}

func (d *Docker) Exec(ctx context.Context, containerId string, execCfg container.ExecOptions) ([]byte, error) {
	execResp, err := d.directApi.ContainerExecCreate(ctx, containerId, execCfg)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	if !execCfg.AttachStdout && !execCfg.AttachStderr {
		return nil, nil
	}

	// Attach to execution
	attachResp, err := d.directApi.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return nil, rerrors.Wrap(err)
	}
	defer attachResp.Close()

	dataOut := bytes.NewBuffer(nil)

	_, err = io.Copy(dataOut, attachResp.Reader)
	if err != nil {
		return nil, rerrors.Wrap(err)
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

func (d *Docker) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
	if config.Labels == nil {
		config.Labels = map[string]string{}
	}

	config.Labels[labels.CreatedWithVelezLabel] = "true"

	return d.directApi.ContainerCreate(ctx, config, hostConfig, networkingConfig, platform, containerName)
}
