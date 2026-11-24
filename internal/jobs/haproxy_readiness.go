package jobs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

const (
	haproxyStableTimeout  = 30 * time.Second
	haproxyStableInterval = 500 * time.Millisecond
	haproxyStableWindow   = 5 * time.Second
	haproxyLogTailLines   = 20
)

// haproxyStabilityCheck inspects one container state snapshot. A non-empty
// reason means the container is broken (crash-looping or exited) and waiting
// longer won't help; isStable means it is running cleanly so far, and the
// caller decides how long that must hold before declaring readiness.
func haproxyStabilityCheck(info container.InspectResponse) (isStable bool, reason string) {
	if info.ContainerJSONBase == nil || info.State == nil {
		return false, "container state is unknown"
	}

	state := info.State

	if state.Restarting {
		return false, fmt.Sprintf("container is restarting (restart count %d)", info.RestartCount)
	}

	if info.RestartCount > 0 {
		return false, fmt.Sprintf("container restarted %d time(s)", info.RestartCount)
	}

	if !state.Running {
		reason := fmt.Sprintf("container exited with code %d", state.ExitCode)
		if state.Error != "" {
			reason += ": " + state.Error
		}

		return false, reason
	}

	return true, ""
}

// waitHaproxyStable polls the container until it has been running cleanly for
// haproxyStableWindow. A crash-looping container (unless_stopped restart
// policy) shows Restarting or a non-zero RestartCount within that window and
// fails with a user error carrying the log tail.
func waitHaproxyStable(ctx context.Context, runtime container_runtime.ContainerRuntime, containerId string) error {
	pollCtx, cancel := context.WithTimeout(ctx, haproxyStableTimeout)
	defer cancel()

	ticker := time.NewTicker(haproxyStableInterval)
	defer ticker.Stop()

	var stableSince time.Time

	for {
		info, found, err := runtime.Inspect(pollCtx, containerId)
		if err != nil {
			return rerrors.Wrap(err, "error inspecting HAProxy container")
		}

		if !found {
			msg := fmt.Sprintf("HAProxy container %s disappeared while waiting for it to stabilize", containerId)

			return user_errors.New(msg)
		}

		isStable, reason := haproxyStabilityCheck(info)
		if reason != "" {
			return haproxyUnstableError(pollCtx, runtime, containerId, reason)
		}

		if isStable {
			if stableSince.IsZero() {
				stableSince = time.Now()
			} else if time.Since(stableSince) >= haproxyStableWindow {
				return nil
			}
		}

		select {
		case <-pollCtx.Done():
			return haproxyUnstableError(ctx, runtime, containerId, "did not stay running for "+haproxyStableWindow.String())
		case <-ticker.C:
		}
	}
}

// awaitHaproxyStableJob gates the flow on the container staying up after the
// deploy/reload wait. The create flow knows only the service name (the
// container id comes from the child CreateSmerd task), the select-master flow
// already has the container id.
type awaitHaproxyStableJob struct {
	runtimes    container_runtime.RuntimeResolver
	environment string
	serviceName string
	containerId string
}

func (j *awaitHaproxyStableJob) Do(ctx context.Context) error {
	runtime, err := j.runtimes.Runtime(ctx, j.environment)
	if err != nil {
		return rerrors.Wrap(err, "error resolving HAProxy container runtime")
	}

	containerId := j.containerId

	if containerId == "" {
		listReq := &velez_api.ListSmerds_Request{
			Label: map[string]string{labels.VervServiceLabel: j.serviceName},
		}

		containers, err := runtime.ListContainers(ctx, listReq)
		if err != nil {
			return rerrors.Wrap(err, "error listing HAProxy containers")
		}

		if len(containers) == 0 {
			return user_errors.New("no container found for HAProxy service " + j.serviceName)
		}

		containerId = containers[0].ID
	}

	return waitHaproxyStable(ctx, runtime, containerId)
}

func haproxyUnstableError(
	ctx context.Context,
	runtime container_runtime.ContainerRuntime,
	containerId, reason string,
) error {
	detail := reason

	tail := haproxyLogTail(ctx, runtime, containerId)
	if tail != "" {
		detail += "\nlast container log lines:\n" + tail
	}

	return rerrors.Wrap(user_errors.ErrHaproxyNotStable, detail)
}

// haproxyLogTail returns the last log lines of the container, or "" when the
// stream can't be read - diagnostics must never mask the real failure.
func haproxyLogTail(ctx context.Context, runtime container_runtime.ContainerRuntime, containerId string) string {
	logsReq := container_runtime.LogsRequest{Tail: haproxyLogTailLines}

	logs, err := runtime.Logs(ctx, containerId, logsReq)
	if err != nil {
		return ""
	}

	defer func() { _ = logs.Close() }()

	raw, err := io.ReadAll(logs)
	if err != nil {
		return ""
	}

	var demuxed bytes.Buffer

	_, err = stdcopy.StdCopy(&demuxed, &demuxed, bytes.NewReader(raw))
	if err != nil {
		return strings.TrimSpace(string(raw))
	}

	return strings.TrimSpace(demuxed.String())
}
