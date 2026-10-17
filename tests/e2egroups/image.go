package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.redsock.ru/rerrors"
)

const (
	velezImageRef    = "velez:e2e"
	envVelezImage    = "VELEZ_E2E_VELEZ_IMAGE"
	imageDockerfile  = "tests/e2e/image/Dockerfile"
	imageVersion     = "e2e"
	imageContextPerm = 0o755

	archArm64 = "arm64"
	archAmd64 = "amd64"
)

//nolint:forbidigo // package-private tool sentinel, not shared/user-facing
var errUnknownArch = rerrors.New("unsupported docker daemon architecture")

// daemonGoArch maps the bootstrap daemon's architecture to GOARCH; the image
// runs on that daemon (and, loaded into a DinD, on the same architecture).
func daemonGoArch(ctx context.Context) (string, error) {
	out, err := docker(ctx, "info", "--format", "{{.Architecture}}")
	if err != nil {
		return "", rerrors.Wrap(err, "error reading docker daemon architecture: "+out)
	}

	arch := strings.TrimSpace(out)

	switch arch {
	case "aarch64", "arm64":
		return archArm64, nil
	case "x86_64", "amd64":
		return archAmd64, nil
	default:
		return "", rerrors.Wrap(errUnknownArch, arch)
	}
}

func runBuildStep(cmd *exec.Cmd, description string) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return rerrors.Wrap(err, description)
	}

	return nil
}

func copyIntoContext(ctx context.Context, root, contextDir, name string) error {
	cmd := exec.CommandContext(ctx, "cp", "-R", filepath.Join(root, name), filepath.Join(contextDir, name))

	return runBuildStep(cmd, "error copying "+name+" into the image context")
}

// buildVelezImage cross-compiles Velez natively and packs it onto the cached
// alpine, so the build never touches the network: --pull=false makes the
// builder take the base image from the daemon's local store.
func buildVelezImage(ctx context.Context) error {
	started := time.Now()

	root, err := findRepoRoot()
	if err != nil {
		return rerrors.Wrap(err, "error finding repo root")
	}

	arch, err := daemonGoArch(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error detecting image architecture")
	}

	contextDir, err := os.MkdirTemp("", "velez-e2e-image-")
	if err != nil {
		return rerrors.Wrap(err, "error creating image context")
	}

	defer func() { _ = os.RemoveAll(contextDir) }()

	ldflags := "-X go.vervstack.ru/Velez/internal/version.version=" + imageVersion

	goBuild := exec.CommandContext(ctx, "go", "build", "-ldflags", ldflags,
		"-o", filepath.Join(contextDir, "service"), "./cmd/service/main.go")

	goBuild.Dir = root

	goBuild.Env = append(os.Environ(), "GOOS=linux", "GOARCH="+arch, "CGO_ENABLED=0")

	err = runBuildStep(goBuild, "error cross-compiling velez")
	if err != nil {
		return rerrors.Wrap(err, "error building velez binary")
	}

	for _, name := range []string{"config", "migrations"} {
		err = copyIntoContext(ctx, root, contextDir, name)
		if err != nil {
			return rerrors.Wrap(err, "error assembling image context")
		}
	}

	err = os.Chmod(contextDir, imageContextPerm)
	if err != nil {
		return rerrors.Wrap(err, "error chmod image context")
	}

	dockerBuild := exec.CommandContext(ctx, "docker", dockerArgs(
		"build", "--pull=false", "-q",
		"-f", filepath.Join(root, imageDockerfile),
		"-t", velezImageRef,
		contextDir,
	)...)

	err = runBuildStep(dockerBuild, "error building velez image")
	if err != nil {
		return rerrors.Wrap(err, "error building image "+velezImageRef)
	}

	fmt.Printf("built %s (%s) in %s\n", velezImageRef, arch, time.Since(started).Round(time.Millisecond))

	return nil
}
