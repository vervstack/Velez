package jobs

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/parser"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	stepLinkBindMounts = "link_bind_mounts"

	localVolumeDriver = "local"
)

var invalidVolumeNameChars = regexp.MustCompile(`[^a-zA-Z0-9_.-]+`)

type bindMountLink = velez_api.RegisterContainer_Request_BindMountLink

// linkedBindMount is a bind mount of the registered container replaced by a
// named volume backed by the same host directory.
type linkedBindMount struct {
	source     string
	target     string
	volumeName string
}

// deriveLinkedVolumeName builds "<service>_<target path>" with every
// character Docker rejects in a volume name collapsed to "_".
func deriveLinkedVolumeName(serviceName, target string) string {
	raw := serviceName + "_" + strings.Trim(target, "/")

	return invalidVolumeNameChars.ReplaceAllString(raw, "_")
}

// resolveBindMountLinks pairs every bind mount of the container with its
// link. A bind mount without a link, or a link matching no bind mount, is
// refused.
func resolveBindMountLinks(
	serviceName string, mounts []container.MountPoint, links []*bindMountLink,
) ([]linkedBindMount, error) {
	linksBySource := make(map[string]*bindMountLink, len(links))
	for _, link := range links {
		linksBySource[link.GetSource()] = link
	}

	linked := make([]linkedBindMount, 0, len(links))
	unlinked := make([]string, 0)
	matched := make(map[string]struct{}, len(links))

	for _, m := range mounts {
		if m.Type != mount.TypeBind {
			continue
		}

		link, isLinked := linksBySource[m.Source]
		if !isLinked {
			unlinked = append(unlinked, m.Source+" -> "+m.Destination)

			continue
		}

		matched[m.Source] = struct{}{}

		volumeName := link.GetVolumeName()
		if volumeName == "" {
			volumeName = deriveLinkedVolumeName(serviceName, m.Destination)
		}

		linked = append(linked, linkedBindMount{
			source:     m.Source,
			target:     m.Destination,
			volumeName: volumeName,
		})
	}

	if len(unlinked) != 0 {
		return nil, rerrors.Wrap(user_errors.ErrBindMountNotLinked, strings.Join(unlinked, ", "))
	}

	unknown := make([]string, 0)

	for _, link := range links {
		_, isMatched := matched[link.GetSource()]
		if !isMatched {
			unknown = append(unknown, link.GetSource())
		}
	}

	if len(unknown) != 0 {
		return nil, rerrors.Wrap(user_errors.ErrBindMountLinkUnknown, strings.Join(unknown, ", "))
	}

	return linked, nil
}

func linkedVolumes(linked []linkedBindMount) []*velez_api.Volume {
	out := make([]*velez_api.Volume, 0, len(linked))

	for _, l := range linked {
		out = append(out, &velez_api.Volume{
			VolumeName:    l.volumeName,
			ContainerPath: l.target,
		})
	}

	return out
}

func bindVolumeRequest(l linkedBindMount) container_runtime.EnsureVolumeRequest {
	return container_runtime.EnsureVolumeRequest{
		Name:   l.volumeName,
		Driver: localVolumeDriver,
		DriverOpts: map[string]string{
			"type":   "none",
			"o":      "bind",
			"device": l.source,
		},
	}
}

// registerVolumes are the volumes the registered container is recreated with:
// its named volumes plus the linked bind mounts.
func registerVolumes(info container.InspectResponse, linked []linkedBindMount) []*velez_api.Volume {
	volumes := parser.ToVolume(info.Mounts)

	return append(volumes, linkedVolumes(linked)...)
}

type registerUpgradeAccessor interface {
	GetServiceName() string
	GetKeepPortMapping() bool
	GetPorts() []*velez_api.Port
	GetBindMountLinks() []*bindMountLink
}

// applyRegisterOverrides sets the ports, volumes and stop-first behaviour the
// recreate upgrade needs. Ports are always overridden (empty means none) since
// the request states them explicitly; volumes only when bind mounts are linked.
func applyRegisterOverrides(
	payload *velez_api.UpgradeSmerdTaskPayload, info container.InspectResponse, req registerUpgradeAccessor,
) error {
	payload.StopOldFirst = req.GetKeepPortMapping()

	ports := req.GetPorts()
	if req.GetKeepPortMapping() {
		ports = parser.ToPortsFromInspect(info)
	}

	payload.PortsOverride = &velez_api.UpgradeSmerdTaskPayload_PortsOverride{Ports: ports}

	if len(req.GetBindMountLinks()) == 0 {
		return nil
	}

	linked, err := resolveBindMountLinks(req.GetServiceName(), info.Mounts, req.GetBindMountLinks())
	if err != nil {
		return rerrors.Wrap(err)
	}

	payload.VolumesOverride = &velez_api.UpgradeSmerdTaskPayload_VolumesOverride{
		Volumes: registerVolumes(info, linked),
	}

	return nil
}

// linkBindMountsJob creates the named volumes that replace the container's
// bind mounts, before the container is touched.
type linkBindMountsJob struct {
	runtimes container_runtime.RuntimeResolver

	req registerRequestAccessor
}

func (j *linkBindMountsJob) Do(ctx context.Context) error {
	if len(j.req.GetBindMountLinks()) == 0 {
		return nil
	}

	runtime, err := j.runtimes.Runtime(ctx, j.req.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	info, found, err := runtime.InspectAny(ctx, j.req.GetContainerId())
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container")
	}

	if !found {
		return rerrors.Wrap(user_errors.ErrRegisterContainerNotFound)
	}

	linked, err := resolveBindMountLinks(j.req.GetServiceName(), info.Mounts, j.req.GetBindMountLinks())
	if err != nil {
		return rerrors.Wrap(err)
	}

	for _, l := range linked {
		err = runtime.EnsureVolume(ctx, bindVolumeRequest(l))
		if err != nil {
			return rerrors.Wrap(err, fmt.Sprintf("error creating volume %q", l.volumeName))
		}
	}

	return nil
}
