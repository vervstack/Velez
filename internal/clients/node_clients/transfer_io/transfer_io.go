// Package transfer_io is the docker side of a container transfer: saving and
// loading images, reading and writing volume content as tar streams.
package transfer_io

import (
	"context"
	"io"

	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"
)

// DockerClient is the narrow slice of node_clients.Docker this package needs;
// it avoids an import cycle with node_clients, which exposes TransferIO.
type DockerClient interface {
	Client() client.APIClient
}

type TransferIO interface {
	ImageSave(ctx context.Context, imageRef string) (io.ReadCloser, error)
	ImageLoad(ctx context.Context, tarStream io.Reader) error
	// VolumeSize is best effort: 0 when unknown.
	VolumeSize(ctx context.Context, volumeName string) (int64, error)
	// VolumeRead returns a tar of the volume content.
	VolumeRead(ctx context.Context, volumeName string) (io.ReadCloser, error)
	// VolumeWrite creates the volume and extracts the tar into it.
	VolumeWrite(ctx context.Context, volumeName string, tarStream io.Reader) error
}

//nolint:unused // dependencies are consumed once the stubs are filled in
type transferIO struct {
	docker DockerClient
}

func New(docker DockerClient) TransferIO {
	return &transferIO{docker: docker}
}

func (t *transferIO) ImageSave(_ context.Context, _ string) (io.ReadCloser, error) {
	return nil, rerrors.Wrap(errNotImplemented)
}

func (t *transferIO) ImageLoad(_ context.Context, _ io.Reader) error {
	return rerrors.Wrap(errNotImplemented)
}

func (t *transferIO) VolumeSize(_ context.Context, _ string) (int64, error) {
	return 0, rerrors.Wrap(errNotImplemented)
}

func (t *transferIO) VolumeRead(_ context.Context, _ string) (io.ReadCloser, error) {
	return nil, rerrors.Wrap(errNotImplemented)
}

func (t *transferIO) VolumeWrite(_ context.Context, _ string, _ io.Reader) error {
	return rerrors.Wrap(errNotImplemented)
}
