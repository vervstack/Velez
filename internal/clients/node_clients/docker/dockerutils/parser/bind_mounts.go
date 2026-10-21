package parser

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

const (
	hostMode = "host"
)

var (
	propagationFromApi = map[velez_api.BindPropagation]mount.Propagation{
		velez_api.BindPropagation_BIND_PROPAGATION_PRIVATE:  mount.PropagationPrivate,
		velez_api.BindPropagation_BIND_PROPAGATION_RPRIVATE: mount.PropagationRPrivate,
		velez_api.BindPropagation_BIND_PROPAGATION_SHARED:   mount.PropagationShared,
		velez_api.BindPropagation_BIND_PROPAGATION_RSHARED:  mount.PropagationRShared,
		velez_api.BindPropagation_BIND_PROPAGATION_SLAVE:    mount.PropagationSlave,
		velez_api.BindPropagation_BIND_PROPAGATION_RSLAVE:   mount.PropagationRSlave,
	}

	propagationToApi = map[mount.Propagation]velez_api.BindPropagation{
		mount.PropagationPrivate:  velez_api.BindPropagation_BIND_PROPAGATION_PRIVATE,
		mount.PropagationRPrivate: velez_api.BindPropagation_BIND_PROPAGATION_RPRIVATE,
		mount.PropagationShared:   velez_api.BindPropagation_BIND_PROPAGATION_SHARED,
		mount.PropagationRShared:  velez_api.BindPropagation_BIND_PROPAGATION_RSHARED,
		mount.PropagationSlave:    velez_api.BindPropagation_BIND_PROPAGATION_SLAVE,
		mount.PropagationRSlave:   velez_api.BindPropagation_BIND_PROPAGATION_RSLAVE,
	}
)

func FromBindMount(settings *velez_api.Container_Settings) []mount.Mount {
	binds := settings.GetBindMounts()
	if len(binds) == 0 {
		return nil
	}

	out := make([]mount.Mount, 0, len(binds))

	for _, item := range binds {
		m := mount.Mount{
			Type:     mount.TypeBind,
			Source:   item.GetSource(),
			Target:   item.GetTarget(),
			ReadOnly: item.GetIsReadOnly(),
		}

		propagation, ok := propagationFromApi[item.GetPropagation()]
		if ok {
			m.BindOptions = &mount.BindOptions{Propagation: propagation}
		}

		out = append(out, m)
	}

	return out
}

func ToBindMount(mounts []container.MountPoint) []*velez_api.BindMount {
	out := make([]*velez_api.BindMount, 0, len(mounts))

	for _, item := range mounts {
		if item.Type != mount.TypeBind {
			continue
		}

		out = append(out, &velez_api.BindMount{
			Source:      item.Source,
			Target:      item.Destination,
			IsReadOnly:  !item.RW,
			Propagation: propagationToApi[item.Propagation],
		})
	}

	if len(out) == 0 {
		return nil
	}

	return out
}

func ToElevatedAccess(hostConfig *container.HostConfig) *velez_api.ElevatedAccess {
	if hostConfig == nil {
		return nil
	}

	access := &velez_api.ElevatedAccess{
		IsPrivileged: hostConfig.Privileged,
		CapAdd:       []string(hostConfig.CapAdd),
		IsHostPid:    hostConfig.PidMode.IsHost(),
		IsHostIpc:    hostConfig.IpcMode.IsHost(),
		IsHostUserns: hostConfig.UsernsMode.IsHost(),
	}

	for _, device := range hostConfig.Devices {
		access.Devices = append(access.Devices, &velez_api.DeviceMapping{
			HostPath:      device.PathOnHost,
			ContainerPath: device.PathInContainer,
			Permissions:   device.CgroupPermissions,
		})
	}

	isElevated := access.GetIsPrivileged() ||
		len(access.GetCapAdd()) != 0 ||
		len(access.GetDevices()) != 0 ||
		access.GetIsHostPid() ||
		access.GetIsHostIpc() ||
		access.GetIsHostUserns()
	if !isElevated {
		return nil
	}

	return access
}

func ApplyElevatedAccess(hostConfig *container.HostConfig, access *velez_api.ElevatedAccess) {
	if hostConfig == nil || access == nil {
		return
	}

	hostConfig.Privileged = hostConfig.Privileged || access.GetIsPrivileged()

	if len(access.GetCapAdd()) != 0 {
		hostConfig.CapAdd = access.GetCapAdd()
	}

	for _, device := range access.GetDevices() {
		hostConfig.Devices = append(hostConfig.Devices, container.DeviceMapping{
			PathOnHost:        device.GetHostPath(),
			PathInContainer:   device.GetContainerPath(),
			CgroupPermissions: device.GetPermissions(),
		})
	}

	if access.GetIsHostPid() {
		hostConfig.PidMode = hostMode
	}

	if access.GetIsHostIpc() {
		hostConfig.IpcMode = hostMode
	}

	if access.GetIsHostUserns() {
		hostConfig.UsernsMode = hostMode
	}
}
