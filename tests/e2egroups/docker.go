package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"

	"go.redsock.ru/rerrors"
)

const (
	envBootstrapHost = "VELEZ_E2E_DOCKER_HOST"
)

func dockerArgs(args ...string) []string {
	host := os.Getenv(envBootstrapHost)
	if host == "" {
		return args
	}

	return append([]string{"-H", host}, args...)
}

// docker runs the docker CLI against the bootstrap daemon and returns its
// combined output.
func docker(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "docker", dockerArgs(args...)...)

	var out bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return out.String(), rerrors.Wrap(err, "error running docker "+strings.Join(args, " "))
	}

	return out.String(), nil
}

func isNotFound(output string) bool {
	return strings.Contains(strings.ToLower(output), "no such")
}

func requireCacheVolume(ctx context.Context, g group) error {
	out, err := docker(ctx, "volume", "inspect", g.cacheVolumeName())
	if err != nil {
		if isNotFound(out) {
			return rerrors.Wrap(errNotPrepared, g.cacheVolumeName())
		}

		return rerrors.Wrap(err, "error inspecting volume: "+out)
	}

	return nil
}

func removeContainer(ctx context.Context, name string) error {
	out, err := docker(ctx, "rm", "-f", name)
	if err != nil && !isNotFound(out) {
		return rerrors.Wrap(err, "error removing container: "+out)
	}

	return nil
}

func removeVolume(ctx context.Context, name string) error {
	out, err := docker(ctx, "volume", "rm", "-f", name)
	if err != nil && !isNotFound(out) {
		return rerrors.Wrap(err, "error removing volume: "+out)
	}

	return nil
}
