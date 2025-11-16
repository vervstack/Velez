package docker

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox/closer"

	"go.vervstack.ru/Velez/internal/clients/docker/dockerutils"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

type Docker struct {
	client client.APIClient
}

func NewClient() (*Docker, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting docker client")
	}

	closer.Add(cli.Close)

	return &Docker{
		client: cli,
	}, nil
}

func (d *Docker) PullImage(ctx context.Context, imageName string) (image.InspectResponse, error) {
	_, err := dockerutils.PullImage(ctx, d.client, imageName, false)
	if err != nil {
		return image.InspectResponse{}, rerrors.Wrap(err, "error pulling image")
	}

	img, err := d.client.ImageInspect(ctx, imageName)
	if err != nil {
		return image.InspectResponse{}, rerrors.Wrap(err, "error inspecting image")
	}

	return img, nil
}

func (d *Docker) Remove(ctx context.Context, contUUID string) error {
	roReq := container.RemoveOptions{
		Force: true,
	}

	err := d.client.ContainerRemove(ctx, contUUID, roReq)

	if err != nil {
		if !strings.Contains(err.Error(), NoSuchContainerError) {
			return nil
		}
		return rerrors.Wrap(err, "error removing container")
	}

	return nil
}

func (d *Docker) ListContainers(ctx context.Context, req *velez_api.ListSmerds_Request) ([]container.Summary, error) {
	list, err := dockerutils.ListContainers(ctx, d.client, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	return list, nil
}

func (d *Docker) Exec(ctx context.Context, containerId string, execCfg container.ExecOptions) ([]byte, error) {
	execCfg.AttachStdout = true
	execCfg.AttachStderr = true

	execResp, err := d.client.ContainerExecCreate(ctx, containerId, execCfg)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	// Attach to execution
	attachResp, err := d.client.ContainerExecAttach(ctx, execResp.ID, container.ExecAttachOptions{})
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
	return d.client
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
