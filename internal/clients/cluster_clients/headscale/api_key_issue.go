package headscale

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"go.redsock.ru/rerrors"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

const (
	managedApiKeyLifetime = "87600h"
)

type keyIssuer struct {
	docker        node_clients.Docker
	containerName string
}

func IssueContainerApiKey(ctx context.Context, docker node_clients.Docker, containerName string) (string, error) {
	issuer := keyIssuer{docker: docker, containerName: containerName}

	return issuer.issueNewKey(ctx, "--expiration", managedApiKeyLifetime)
}

func issueNewAPIKey(ctx context.Context, docker node_clients.Docker, containerName string) (string, error) {
	issuer := keyIssuer{docker: docker, containerName: containerName}

	err := issuer.collectGarbage(ctx)
	if err != nil {
		return "", rerrors.Wrap(err, "error removing expired API keys")
	}

	return issuer.issueNewKey(ctx)
}

func (s *keyIssuer) collectGarbage(ctx context.Context) error {
	data, err := s.run(ctx, "list", "--output", "json-line")
	if err != nil {
		return rerrors.Wrap(err, "error listing API keys")
	}

	// The CLI encodes protobuf timestamps as JSON objects with seconds and nanos.
	var keys []struct {
		Prefix     string                 `json:"prefix"`
		Expiration *timestamppb.Timestamp `json:"expiration"`
	}

	err = json.Unmarshal(data, &keys)
	if err != nil {
		return rerrors.Wrap(err, "error decoding API key list")
	}

	now := time.Now()

	for _, key := range keys {
		if key.Prefix == "" || key.Expiration == nil || key.Expiration.AsTime().After(now) {
			continue
		}

		_, err = s.run(ctx, "delete", "--prefix", key.Prefix)
		if err != nil {
			return rerrors.Wrap(err, "error deleting expired API key")
		}
	}

	return nil
}

func (s *keyIssuer) issueNewKey(ctx context.Context, flags ...string) (string, error) {
	args := append([]string{"create", "--output", "json-line"}, flags...)

	data, err := s.run(ctx, args...)
	if err != nil {
		return "", rerrors.Wrap(err, "error issuing new API key")
	}

	var key string

	err = json.Unmarshal(data, &key)
	if err != nil || key == "" {
		return "", rerrors.Wrap(user_errors.ErrHeadscaleCantParseOutput, "Headscale CLI returned an invalid key")
	}

	return key, nil
}

func (s *keyIssuer) run(ctx context.Context, args ...string) ([]byte, error) {
	docker := s.docker.Client()
	execOpts := container.ExecOptions{
		Cmd:          append([]string{"headscale", "apikey"}, args...),
		Env:          []string{"HEADSCALE_LOG_FORMAT=text", "NO_COLOR=1"},
		AttachStdout: true,
		AttachStderr: true,
	}

	exec, err := docker.ContainerExecCreate(ctx, s.containerName, execOpts)
	if err != nil {
		return nil, rerrors.Wrap(err, "error preparing Headscale command")
	}

	attachOpts := container.ExecAttachOptions{}

	attached, err := docker.ContainerExecAttach(ctx, exec.ID, attachOpts)
	if err != nil {
		return nil, rerrors.Wrap(err, "error running Headscale command")
	}

	defer attached.Close()

	stopClose := context.AfterFunc(ctx, attached.Close)
	defer stopClose()

	var stdout bytes.Buffer

	_, err = stdcopy.StdCopy(&stdout, io.Discard, attached.Reader)
	if err != nil {
		return nil, rerrors.Wrap(err, "error reading Headscale command output")
	}

	result, err := docker.ContainerExecInspect(ctx, exec.ID)
	if err != nil {
		return nil, rerrors.Wrap(err, "error inspecting Headscale command result")
	}

	if result.ExitCode != 0 {
		return nil, rerrors.Wrap(user_errors.ErrHeadscaleUnexpectedStatus,
			fmt.Sprintf("Headscale command exited with status %d", result.ExitCode))
	}

	return stdout.Bytes(), nil
}
