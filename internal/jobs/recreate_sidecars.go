package jobs

import (
	"context"
	"maps"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	stepRecreateSidecars = "recreate_sidecars"

	sidecarBackupSuffix = "_old"
	networkModeOfPrefix = "container:"
)

// sidecarGroupAccessor describes the containers recreate_sidecars works on.
// Values are read at Do/Rollback time, so a task can fill them in an earlier
// job.
type sidecarGroupAccessor interface {
	GetEnvironment() string
	GetServiceName() string
	// GetContainerName is the network root, already recreated and running.
	GetContainerName() string
	GetSidecarContainerNames() []string
}

// recreateSidecarsJob recreates containers that shared the network namespace
// of a replaced root, so they join the new root's namespace. All or nothing:
// each old sidecar is kept as "<name>_old" until every sidecar is running,
// and Rollback puts the old ones back.
type recreateSidecarsJob struct {
	runtimes container_runtime.RuntimeResolver
	group    sidecarGroupAccessor
}

func newRecreateSidecarsJob(
	runtimes container_runtime.RuntimeResolver, group sidecarGroupAccessor,
) *recreateSidecarsJob {
	return &recreateSidecarsJob{runtimes: runtimes, group: group}
}

// sidecarLabels are the labels a recreated sidecar carries: the service link
// plus the sidecar marker, which keeps it off the service pages.
func sidecarLabels(serviceName string) map[string]string {
	merged := make(map[string]string)

	if serviceName != "" {
		merged = registeredLabels(serviceName, velez_api.ServicePattern_SERVICE_PATTERN_UNSPECIFIED)
	}

	merged[labels.Sidecar] = labelTrueValue

	return merged
}

// isRegisteredSidecar reports whether the labels are those sidecarLabels
// stamps for serviceName.
func isRegisteredSidecar(containerLabels map[string]string, serviceName string) bool {
	_, isSidecar := containerLabels[labels.Sidecar]

	return isSidecar && containerLabels[labels.VervServiceLabel] == serviceName
}

// sidecarCreateRequest clones an inspected sidecar into a create request that
// joins networkMode. Settings Docker refuses next to a container network mode
// (hostname, exposed and published ports, endpoints) are dropped.
func sidecarCreateRequest(
	name, serviceName string,
	networkMode container.NetworkMode,
	source container.InspectResponse,
) container_runtime.ContainerCreateRequest {
	config := *source.Config

	config.Hostname = ""
	config.Domainname = ""
	config.ExposedPorts = nil

	config.Labels = make(map[string]string, len(source.Config.Labels))

	maps.Copy(config.Labels, source.Config.Labels)
	maps.Copy(config.Labels, sidecarLabels(serviceName))

	hostConfig := *source.HostConfig

	hostConfig.NetworkMode = networkMode
	hostConfig.PortBindings = nil
	hostConfig.PublishAllPorts = false

	return container_runtime.ContainerCreateRequest{
		Config:        &container_runtime.ContainerConfig{Config: &config},
		HostConfig:    &container_runtime.HostConfig{HostConfig: &hostConfig},
		ContainerName: name,
	}
}

func (j *recreateSidecarsJob) Do(ctx context.Context) error {
	names := j.group.GetSidecarContainerNames()
	if len(names) == 0 {
		return nil
	}

	runtime, err := j.runtimes.Runtime(ctx, j.group.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	root, found, err := runtime.Inspect(ctx, j.group.GetContainerName())
	if err != nil {
		return rerrors.Wrap(err, "error inspecting network root")
	}

	if !found {
		return rerrors.Wrap(user_errors.ErrRegisterContainerNotFound, j.group.GetContainerName())
	}

	networkMode := container.NetworkMode(networkModeOfPrefix + root.ID)

	for _, name := range names {
		err = j.recreate(ctx, runtime, name, networkMode)
		if err != nil {
			return rerrors.Wrap(err, "error recreating sidecar "+name)
		}
	}

	for _, name := range names {
		err = runtime.Remove(ctx, name+sidecarBackupSuffix)
		if err != nil {
			return rerrors.Wrap(err, "error dropping backup of sidecar "+name)
		}
	}

	return nil
}

// Rollback drops every sidecar created by Do and restores the old ones from
// their backups. Sidecars without a backup were never touched.
func (j *recreateSidecarsJob) Rollback(ctx context.Context) error {
	names := j.group.GetSidecarContainerNames()
	if len(names) == 0 {
		return nil
	}

	runtime, err := j.runtimes.Runtime(ctx, j.group.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	failures := make([]error, 0)

	for _, name := range names {
		err = j.restore(ctx, runtime, name)
		if err != nil {
			failures = append(failures, rerrors.Wrap(err, "error restoring sidecar "+name))
		}
	}

	err = rerrors.Join(failures...)
	if err != nil {
		return rerrors.Wrap(err, "error rolling back sidecars")
	}

	return nil
}

func (j *recreateSidecarsJob) recreate(
	ctx context.Context,
	runtime container_runtime.ContainerRuntime,
	name string,
	networkMode container.NetworkMode,
) error {
	backupName := name + sidecarBackupSuffix

	current, hasCurrent, err := runtime.Inspect(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting sidecar")
	}

	isJoined := hasCurrent && current.HostConfig.NetworkMode == networkMode
	if isJoined && isRegisteredSidecar(current.Config.Labels, j.group.GetServiceName()) {
		return nil
	}

	backup, hasBackup, err := runtime.Inspect(ctx, backupName)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting sidecar backup")
	}

	source := backup

	switch {
	case hasCurrent && hasBackup:
		return rerrors.Wrap(user_errors.ErrSidecarLeftoverExists, backupName)
	case hasCurrent:
		source = current

		err = j.moveAside(ctx, runtime, name, backupName)
		if err != nil {
			return rerrors.Wrap(err)
		}
	case !hasBackup:
		return rerrors.Wrap(user_errors.ErrRegisterContainerNotFound, name)
	}

	createReq := sidecarCreateRequest(name, j.group.GetServiceName(), networkMode, source)

	created, err := runtime.ContainerCreate(ctx, createReq)
	if err != nil {
		return rerrors.Wrap(err, "error creating sidecar")
	}

	err = runtime.Restart(ctx, created.ID)
	if err != nil {
		return rerrors.Wrap(err, "error starting sidecar")
	}

	return nil
}

// moveAside stops the old sidecar and renames it to its backup name; a failed
// rename brings the old sidecar back up.
func (j *recreateSidecarsJob) moveAside(
	ctx context.Context, runtime container_runtime.ContainerRuntime, name, backupName string,
) error {
	err := runtime.Stop(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error stopping sidecar")
	}

	err = runtime.Rename(ctx, name, backupName)
	if err != nil {
		restartErr := runtime.Restart(ctx, name)

		return rerrors.Wrap(rerrors.Join(err, restartErr), "error renaming sidecar to backup")
	}

	return nil
}

func (j *recreateSidecarsJob) restore(
	ctx context.Context, runtime container_runtime.ContainerRuntime, name string,
) error {
	backupName := name + sidecarBackupSuffix

	_, hasBackup, err := runtime.Inspect(ctx, backupName)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting sidecar backup")
	}

	if !hasBackup {
		return nil
	}

	current, hasCurrent, err := runtime.Inspect(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting sidecar")
	}

	isOurs := hasCurrent && isRegisteredSidecar(current.Config.Labels, j.group.GetServiceName())
	if hasCurrent && !isOurs {
		return nil
	}

	err = runtime.Remove(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error dropping new sidecar")
	}

	err = runtime.Rename(ctx, backupName, name)
	if err != nil {
		return rerrors.Wrap(err, "error renaming backup back")
	}

	err = runtime.Restart(ctx, name)
	if err != nil {
		return rerrors.Wrap(err, "error restarting old sidecar")
	}

	return nil
}
