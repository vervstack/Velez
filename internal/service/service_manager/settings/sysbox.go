package settings

import (
	"context"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/utils/common"
)

const (
	smokeTestImage          = "alpine:3"
	smokeTestNamePrefix     = "velez-sysbox-smoke-"
	smokeTestTimeout        = 90 * time.Second
	smokeTestCleanupTimeout = 30 * time.Second
	rootlessSecurityOption  = "name=rootless"
	snapRootMarker          = "/snap/"
)

var errSmokeTestNonZeroExit = rerrors.New("smoke test container exited with non-zero code")

func (s *Service) GetSysboxStatus(ctx context.Context) (domain.SysboxStatus, error) {
	cli := s.docker.Client()

	info, err := cli.Info(ctx)
	if err != nil {
		return domain.SysboxStatus{}, rerrors.Wrap(err, "error getting docker info")
	}

	_, isRuntimeRegistered := info.Runtimes[sysboxRuntimeName]

	list, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return domain.SysboxStatus{}, rerrors.Wrap(err, "error listing containers")
	}

	var onSysbox int32

	for _, summary := range list {
		inspected, inspectErr := cli.ContainerInspect(ctx, summary.ID)
		if inspectErr != nil {
			continue
		}

		if inspected.HostConfig != nil && inspected.HostConfig.Runtime == sysboxRuntimeName {
			onSysbox++
		}
	}

	status := domain.SysboxStatus{
		OsType:              info.OSType,
		KernelVersion:       info.KernelVersion,
		DockerVersion:       info.ServerVersion,
		IsRootless:          isRootless(info.SecurityOptions),
		IsSnap:              isSnapRoot(info.DockerRootDir),
		IsRuntimeRegistered: isRuntimeRegistered,
		ContainersTotal:     int32(len(list)),
		ContainersOnSysbox:  onSysbox,
	}

	return status, nil
}

func (s *Service) RunSysboxSmokeTest(ctx context.Context) (domain.SysboxSmokeTestResult, error) {
	ctx, cancel := context.WithTimeout(ctx, smokeTestTimeout)
	defer cancel()

	err := s.runSmokeTest(ctx)
	if err != nil {
		result := domain.SysboxSmokeTestResult{
			Failure: err.Error(),
		}

		return result, nil
	}

	result := domain.SysboxSmokeTestResult{
		IsPassed: true,
	}

	return result, nil
}

func (s *Service) runSmokeTest(ctx context.Context) error {
	cli := s.docker.Client()

	pullReader, err := cli.ImagePull(ctx, smokeTestImage, image.PullOptions{})
	if err != nil {
		return rerrors.Wrap(err, "error pulling smoke test image")
	}

	defer common.CloseWithLog(pullReader.Close, "smoke test image pull")

	_, err = io.Copy(io.Discard, pullReader)
	if err != nil {
		return rerrors.Wrap(err, "error reading smoke test image pull")
	}

	config := &container.Config{
		Image: smokeTestImage,
		Cmd:   []string{"echo", "ok"},
	}
	hostConfig := &container.HostConfig{
		Runtime: sysboxRuntimeName,
	}
	name := smokeTestNamePrefix + strconv.FormatInt(time.Now().UnixNano(), 10)

	created, err := s.docker.ContainerCreate(ctx, config, hostConfig, nil, nil, name, "")
	if err != nil {
		return rerrors.Wrap(err, "error creating smoke test container")
	}

	defer s.removeSmokeTestContainer(created.ID)

	err = cli.ContainerStart(ctx, created.ID, container.StartOptions{})
	if err != nil {
		return rerrors.Wrap(err, "error starting smoke test container")
	}

	statusCh, errCh := cli.ContainerWait(ctx, created.ID, container.WaitConditionNotRunning)

	select {
	case err = <-errCh:
		return rerrors.Wrap(err, "error waiting for smoke test container")
	case status := <-statusCh:
		if status.StatusCode != 0 {
			return rerrors.Wrap(errSmokeTestNonZeroExit, "exit code "+strconv.FormatInt(status.StatusCode, 10))
		}
	}

	return nil
}

func (s *Service) removeSmokeTestContainer(id string) {
	ctx, cancel := context.WithTimeout(context.Background(), smokeTestCleanupTimeout)
	defer cancel()

	options := container.RemoveOptions{
		Force: true,
	}

	err := s.docker.Client().ContainerRemove(ctx, id, options)
	if err != nil {
		log.Error().
			Str("container_id", id).
			Err(err).
			Msg("error removing smoke test container")
	}
}

func isRootless(securityOptions []string) bool {
	for _, option := range securityOptions {
		if strings.Contains(option, rootlessSecurityOption) {
			return true
		}
	}

	return false
}

func isSnapRoot(dockerRootDir string) bool {
	return strings.Contains(dockerRootDir, snapRootMarker)
}
