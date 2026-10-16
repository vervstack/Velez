package dind

import (
	"context"
	"strings"
	"time"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"
)

const (
	cacheMaxAge = 72 * time.Hour
)

// collectStaleCaches removes DinD cache volumes (and their stopped containers)
// unused for longer than cacheMaxAge. Best effort: failures are logged only.
func collectStaleCaches(ctx context.Context, cli *client.Client, currentCacheVolume string) {
	listOpts := volume.ListOptions{
		Filters: filters.NewArgs(filters.Arg("label", labelHarness+"=true")),
	}

	resp, err := cli.VolumeList(ctx, listOpts)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("error listing dind cache volumes for gc")

		return
	}

	names := make([]string, 0, len(resp.Volumes))
	createdAt := make(map[string]string, len(resp.Volumes))

	for _, v := range resp.Volumes {
		names = append(names, v.Name)
		createdAt[v.Name] = v.CreatedAt
	}

	now := time.Now()

	for _, name := range selectCacheCandidates(names, currentCacheVolume) {
		err = collectCacheIfStale(ctx, cli, name, createdAt[name], now)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("volume", name).Msg("error collecting stale dind cache volume")
		}
	}
}

func collectCacheIfStale(
	ctx context.Context,
	cli *client.Client,
	volumeName string,
	volumeCreatedAt string,
	now time.Time,
) error {
	containerName := containerNameOfCache(volumeName)
	volumeCreated := parseDockerTime(volumeCreatedAt)

	info, err := cli.ContainerInspect(ctx, containerName)
	hasContainer := err == nil

	if err != nil && !errdefs.IsNotFound(err) {
		return rerrors.Wrap(err, "error inspecting container "+containerName)
	}

	lastUsed := volumeCreated
	isRunning := false

	if hasContainer {
		isRunning = info.State != nil && info.State.Running
		lastUsed = containerLastUsed(info, volumeCreated)
	}

	if lastUsed.IsZero() || !isCacheStale(now, lastUsed, isRunning) {
		return nil
	}

	if hasContainer {
		removeOpts := container.RemoveOptions{
			Force:         true,
			RemoveVolumes: true,
		}

		err = removeContainerRetrying(ctx, cli, containerName, removeOpts)
		if err != nil {
			return rerrors.Wrap(err, "error removing container "+containerName)
		}
	}

	err = cli.VolumeRemove(ctx, volumeName, true)
	if err != nil && !errdefs.IsNotFound(err) {
		return rerrors.Wrap(err, "error removing volume "+volumeName)
	}

	return nil
}

func containerLastUsed(info container.InspectResponse, fallback time.Time) time.Time {
	if info.ContainerJSONBase == nil || info.State == nil {
		return fallback
	}

	finished := parseDockerTime(info.State.FinishedAt)
	if !finished.IsZero() {
		return finished
	}

	started := parseDockerTime(info.State.StartedAt)
	if !started.IsZero() {
		return started
	}

	return fallback
}

// parseDockerTime returns the zero time for an empty, malformed or
// "0001-01-01..." (never happened) timestamp.
func parseDockerTime(raw string) time.Time {
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil || parsed.Year() <= 1 {
		return time.Time{}
	}

	return parsed
}

func isCacheStale(now, lastUsed time.Time, isRunning bool) bool {
	if isRunning {
		return false
	}

	return now.Sub(lastUsed) > cacheMaxAge
}

func selectCacheCandidates(volumeNames []string, currentCacheVolume string) []string {
	candidates := make([]string, 0, len(volumeNames))

	for _, name := range volumeNames {
		if name == currentCacheVolume {
			continue
		}

		if !strings.HasPrefix(name, namePrefix) || !strings.HasSuffix(name, cacheVolumeSuffix) {
			continue
		}

		candidates = append(candidates, name)
	}

	return candidates
}

func containerNameOfCache(volumeName string) string {
	return strings.TrimSuffix(volumeName, cacheVolumeSuffix)
}
