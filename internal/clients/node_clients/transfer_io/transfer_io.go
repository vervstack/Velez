// Package transfer_io is the docker side of a container transfer: saving and
// loading images, reading and writing volume content as tar streams.
package transfer_io

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"
)

const (
	helperImage         = "alpine"
	helperMountPath     = "/data"
	helperNamePrefix    = "velez_transfer_"
	helperRemoveTimeout = 30 * time.Second
	bytesInKilobyte     = 1024
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

type transferIO struct {
	docker DockerClient
}

func New(docker DockerClient) TransferIO {
	return &transferIO{docker: docker}
}

func (t *transferIO) ImageSave(ctx context.Context, imageRef string) (io.ReadCloser, error) {
	cli := t.docker.Client()

	stream, err := cli.ImageSave(ctx, []string{imageRef})
	if err != nil {
		return nil, rerrors.Wrap(err, "error saving image")
	}

	return stream, nil
}

func (t *transferIO) ImageLoad(ctx context.Context, tarStream io.Reader) error {
	cli := t.docker.Client()

	resp, err := cli.ImageLoad(ctx, tarStream, client.ImageLoadWithQuiet(true))
	if err != nil {
		return rerrors.Wrap(err, "error loading image")
	}
	defer closeQuietly(resp.Body)

	type loadMessage struct {
		Error       string `json:"error"`
		ErrorDetail struct {
			Message string `json:"message"`
		} `json:"errorDetail"`
	}

	decoder := json.NewDecoder(resp.Body)

	for {
		var msg loadMessage

		err = decoder.Decode(&msg)
		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			return rerrors.Wrap(err, "error reading image load response")
		}

		if msg.Error != "" || msg.ErrorDetail.Message != "" {
			log.Ctx(ctx).Warn().
				Str("error", msg.Error).
				Str("error_detail", msg.ErrorDetail.Message).
				Msg("image load reported an error")

			return rerrors.Wrap(errImageLoadFailed)
		}
	}
}

func (t *transferIO) VolumeSize(ctx context.Context, volumeName string) (int64, error) {
	helperId, err := t.startHelper(ctx, volumeName, true)
	if err != nil {
		return 0, nil
	}
	defer t.removeHelper(ctx, helperId)

	size := t.duSize(ctx, helperId)

	return size, nil
}

func (t *transferIO) VolumeRead(ctx context.Context, volumeName string) (io.ReadCloser, error) {
	cli := t.docker.Client()

	_, err := cli.VolumeInspect(ctx, volumeName)
	if err != nil {
		return nil, rerrors.Wrap(err, "error inspecting volume")
	}

	helperId, err := t.createHelper(ctx, volumeName, true)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	stream, _, err := cli.CopyFromContainer(ctx, helperId, helperMountPath+"/.")
	if err != nil {
		t.removeHelper(ctx, helperId)

		return nil, rerrors.Wrap(err, "error reading volume content")
	}

	reader := &helperReadCloser{
		ReadCloser: stream,
		cleanup: func() {
			t.removeHelper(ctx, helperId)
		},
	}

	return reader, nil
}

func (t *transferIO) VolumeWrite(ctx context.Context, volumeName string, tarStream io.Reader) error {
	cli := t.docker.Client()

	_, err := cli.VolumeInspect(ctx, volumeName)
	if err == nil {
		return rerrors.Wrap(errVolumeExists)
	}

	if !cerrdefs.IsNotFound(err) {
		return rerrors.Wrap(err, "error inspecting volume")
	}

	createOpts := volume.CreateOptions{Name: volumeName}

	_, err = cli.VolumeCreate(ctx, createOpts)
	if err != nil {
		return rerrors.Wrap(err, "error creating volume")
	}

	helperId, err := t.createHelper(ctx, volumeName, false)
	if err != nil {
		return rerrors.Wrap(err)
	}
	defer t.removeHelper(ctx, helperId)

	copyOpts := container.CopyToContainerOptions{}

	err = cli.CopyToContainer(ctx, helperId, helperMountPath, tarStream, copyOpts)
	if err != nil {
		return rerrors.Wrap(err, "error writing volume content")
	}

	return nil
}

func (t *transferIO) startHelper(ctx context.Context, volumeName string, isReadOnly bool) (string, error) {
	cli := t.docker.Client()

	_, err := cli.VolumeInspect(ctx, volumeName)
	if err != nil {
		return "", rerrors.Wrap(err, "error inspecting volume")
	}

	helperId, err := t.createHelper(ctx, volumeName, isReadOnly)
	if err != nil {
		return "", rerrors.Wrap(err)
	}

	startOpts := container.StartOptions{}

	err = cli.ContainerStart(ctx, helperId, startOpts)
	if err != nil {
		t.removeHelper(ctx, helperId)

		return "", rerrors.Wrap(err, "error starting transfer helper container")
	}

	return helperId, nil
}

func (t *transferIO) createHelper(ctx context.Context, volumeName string, isReadOnly bool) (string, error) {
	cli := t.docker.Client()

	err := t.ensureHelperImage(ctx)
	if err != nil {
		return "", rerrors.Wrap(err)
	}

	name := helperNamePrefix + strings.ReplaceAll(uuid.NewString(), "-", "")

	cfg := container.Config{
		Image: helperImage,
		Cmd:   []string{"sleep", "infinity"},
	}

	volumeMount := mount.Mount{
		Type:     mount.TypeVolume,
		Source:   volumeName,
		Target:   helperMountPath,
		ReadOnly: isReadOnly,
	}

	hostCfg := container.HostConfig{
		NetworkMode: container.NetworkMode("none"),
		Mounts:      []mount.Mount{volumeMount},
	}

	created, err := cli.ContainerCreate(ctx, &cfg, &hostCfg, nil, nil, name)
	if err != nil {
		return "", rerrors.Wrap(err, "error creating transfer helper container")
	}

	if created.ID == "" {
		return "", rerrors.Wrap(errHelperNotStarted)
	}

	return created.ID, nil
}

func (t *transferIO) ensureHelperImage(ctx context.Context) error {
	cli := t.docker.Client()

	_, err := cli.ImageInspect(ctx, helperImage)
	if err == nil {
		return nil
	}

	if !cerrdefs.IsNotFound(err) {
		return rerrors.Wrap(err, "error inspecting helper image")
	}

	pullOpts := image.PullOptions{}

	pull, err := cli.ImagePull(ctx, helperImage, pullOpts)
	if err != nil {
		return rerrors.Wrap(err, "error pulling helper image")
	}
	defer closeQuietly(pull)

	_, err = io.Copy(io.Discard, pull)
	if err != nil {
		return rerrors.Wrap(err, "error pulling helper image")
	}

	return nil
}

func (t *transferIO) removeHelper(ctx context.Context, helperId string) {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), helperRemoveTimeout)
	defer cancel()

	removeOpts := container.RemoveOptions{Force: true}

	err := t.docker.Client().ContainerRemove(cleanupCtx, helperId, removeOpts)
	if err != nil && !cerrdefs.IsNotFound(err) {
		log.Ctx(ctx).Warn().
			Str("container_id", helperId).
			Err(err).
			Msg("error removing transfer helper container")
	}
}

func (t *transferIO) duSize(ctx context.Context, helperId string) int64 {
	cli := t.docker.Client()

	execOpts := container.ExecOptions{
		Cmd:          []string{"du", "-sk", helperMountPath},
		AttachStdout: true,
		AttachStderr: true,
	}

	created, err := cli.ContainerExecCreate(ctx, helperId, execOpts)
	if err != nil {
		return 0
	}

	attachOpts := container.ExecAttachOptions{}

	attached, err := cli.ContainerExecAttach(ctx, created.ID, attachOpts)
	if err != nil {
		return 0
	}
	defer attached.Close()

	var stdout bytes.Buffer

	_, err = stdcopy.StdCopy(&stdout, io.Discard, attached.Reader)
	if err != nil {
		return 0
	}

	fields := strings.Fields(stdout.String())
	if len(fields) == 0 {
		return 0
	}

	kilobytes, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return 0
	}

	return kilobytes * bytesInKilobyte
}

type helperReadCloser struct {
	io.ReadCloser

	cleanup func()
}

func (h *helperReadCloser) Close() error {
	err := h.ReadCloser.Close()

	h.cleanup()

	if err != nil {
		return rerrors.Wrap(err, "error closing volume stream")
	}

	return nil
}

func closeQuietly(closer io.Closer) {
	_ = closer.Close()
}
