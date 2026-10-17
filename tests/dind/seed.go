package dind

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"
)

const (
	hostPullLockName = "velez-e2e-pull.lock"
	hostPullLockMode = 0o600
)

var (
	//nolint:forbidigo // package-private test-infra sentinel, not shared/user-facing
	errImageNotLoaded = rerrors.New("image is absent from dind after loading it from the host")

	//nolint:forbidigo // package-private test-infra sentinel, not shared/user-facing
	errLocalImageMissing = rerrors.New("locally built image is absent from the host, run `make e2e-image`")

	// hostPullMu and the lock file serialize host pulls: several groups'
	// DinDs (separate processes) seed concurrently and may miss the same
	// ref on the host at once.
	hostPullMu sync.Mutex
)

// SeedImage is an image the DinD must hold without ever pulling it from the
// web. Aliases are extra names tagged onto it inside the DinD, for floating
// refs (":latest", an untagged name) the product or tests request verbatim
// while the suite pins one concrete version. A local-only image is built on
// the host (never pulled) and is re-loaded whenever its host image Id differs
// from the DinD's.
type SeedImage struct {
	Ref         string
	Aliases     []string
	IsLocalOnly bool
}

// Seed makes every image available in the DinD daemon. An image the DinD
// lacks is taken from the bootstrap host's image cache (pulled there once,
// under a cross-process lock, if the host lacks it too) and streamed in with
// save/load - the DinD daemon itself never pulls.
func (e *Env) Seed(ctx context.Context, images ...SeedImage) error {
	cli, err := newDockerClient(e.DockerHost)
	if err != nil {
		return rerrors.Wrap(err, "error creating dind docker client")
	}

	defer func() {
		_ = cli.Close()
	}()

	for _, img := range images {
		err = e.seedImage(ctx, cli, img)
		if err != nil {
			return rerrors.Wrap(err, "error seeding image "+img.Ref)
		}
	}

	return nil
}

func (e *Env) seedImage(ctx context.Context, dindCli *client.Client, img SeedImage) error {
	if img.IsLocalOnly {
		return e.seedLocalImage(ctx, dindCli, img)
	}

	isPresent, err := hasImage(ctx, dindCli, img.Ref)
	if err != nil {
		return rerrors.Wrap(err, "error checking dind for image")
	}

	if !isPresent {
		err = ensureHostImage(ctx, e.bootstrap, img.Ref)
		if err != nil {
			return rerrors.Wrap(err, "error caching image on the host")
		}

		err = copyImage(ctx, e.bootstrap, dindCli, img.Ref)
		if err != nil {
			return rerrors.Wrap(err, "error loading image from the host into dind")
		}

		isPresent, err = hasImage(ctx, dindCli, img.Ref)
		if err != nil {
			return rerrors.Wrap(err, "error checking dind for loaded image")
		}

		if !isPresent {
			return rerrors.Wrap(errImageNotLoaded)
		}
	}

	for _, alias := range img.Aliases {
		err = dindCli.ImageTag(ctx, img.Ref, alias)
		if err != nil {
			return rerrors.Wrap(err, "error tagging alias "+alias)
		}
	}

	return nil
}

func (e *Env) seedLocalImage(ctx context.Context, dindCli *client.Client, img SeedImage) error {
	hostId, err := imageId(ctx, e.bootstrap, img.Ref)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting host image")
	}

	if hostId == "" {
		return rerrors.Wrap(errLocalImageMissing, img.Ref)
	}

	dindId, err := imageId(ctx, dindCli, img.Ref)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting dind image")
	}

	if dindId == hostId {
		return nil
	}

	err = copyImage(ctx, e.bootstrap, dindCli, img.Ref)
	if err != nil {
		return rerrors.Wrap(err, "error loading image from the host into dind")
	}

	return nil
}

// imageId returns the image Id behind ref, or "" when the daemon lacks it.
func imageId(ctx context.Context, cli *client.Client, ref string) (string, error) {
	inspected, err := cli.ImageInspect(ctx, ref)
	if err != nil {
		if cerrdefs.IsNotFound(err) {
			return "", nil
		}

		return "", rerrors.Wrap(err, "error inspecting image")
	}

	return inspected.ID, nil
}

func hasImage(ctx context.Context, cli *client.Client, ref string) (bool, error) {
	filterArgs := filters.NewArgs(filters.Arg("reference", ref))

	listOpts := image.ListOptions{
		Filters: filterArgs,
	}

	existing, err := cli.ImageList(ctx, listOpts)
	if err != nil {
		return false, rerrors.Wrap(err, "error listing images")
	}

	return len(existing) > 0, nil
}

func ensureHostImage(ctx context.Context, hostCli *client.Client, ref string) error {
	isPresent, err := hasImage(ctx, hostCli, ref)
	if err != nil {
		return rerrors.Wrap(err, "error checking host for image")
	}

	if isPresent {
		return nil
	}

	hostPullMu.Lock()
	defer hostPullMu.Unlock()

	lockPath := filepath.Join(os.TempDir(), hostPullLockName)

	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, hostPullLockMode)
	if err != nil {
		return rerrors.Wrap(err, "error opening host pull lock")
	}

	defer func() {
		_ = lockFile.Close()
	}()

	err = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX)
	if err != nil {
		return rerrors.Wrap(err, "error locking host pull")
	}

	return ensureImage(ctx, hostCli, ref)
}

func copyImage(ctx context.Context, from, to *client.Client, ref string) error {
	archive, err := from.ImageSave(ctx, []string{ref})
	if err != nil {
		return rerrors.Wrap(err, "error saving image")
	}

	defer func() {
		_ = archive.Close()
	}()

	loadResp, err := to.ImageLoad(ctx, archive, client.ImageLoadWithQuiet(true))
	if err != nil {
		return rerrors.Wrap(err, "error loading image")
	}

	defer func() {
		_ = loadResp.Body.Close()
	}()

	_, err = io.Copy(io.Discard, loadResp.Body)
	if err != nil {
		return rerrors.Wrap(err, "error draining image load log")
	}

	return nil
}
