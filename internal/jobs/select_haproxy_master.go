package jobs

import (
	"context"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage"
)

const (
	SelectHaproxyMasterAction = "select_haproxy_master"

	stepStageHaproxyConfig = "write_config"
	stepCheckHaproxyConfig = "check_config"
	stepReloadHaproxy      = "reload_haproxy"

	haproxyPendingConfigPath = domain.HaproxyConfigPath + ".pending"
	haproxyInstallConfigPath = domain.HaproxyConfigPath + ".next"
	haproxyReloadSignal      = "SIGUSR2"
	modeHaproxyConfig        = 0o644
)

type selectHaproxyMasterHandler struct {
	dataStorage    storage.Storage
	configResolver service.ServiceConfigResolver
	runtimes       container_runtime.RuntimeResolver
}

func NewSelectHaproxyMasterHandler(
	dataStorage storage.Storage,
	configResolver service.ServiceConfigResolver,
	runtimes container_runtime.RuntimeResolver,
) TaskHandler {
	return &selectHaproxyMasterHandler{
		dataStorage:    dataStorage,
		configResolver: configResolver,
		runtimes:       runtimes,
	}
}

func (h *selectHaproxyMasterHandler) Action() string {
	return SelectHaproxyMasterAction
}

func (h *selectHaproxyMasterHandler) NewContext() TaskContext {
	return &velez_api.SelectHaproxyMasterTaskPayload{}
}

func (h *selectHaproxyMasterHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.SelectHaproxyMasterTaskPayload)
	if !ok {
		panic("select_haproxy_master: BuildJobs called with mismatched TaskContext type")
	}

	return []NamedJob{
		{
			Name: stepValidateHaproxyMaster,
			Job: &validateHaproxyMasterJob{
				dataStorage: h.dataStorage,
				runtimes:    h.runtimes,
				environment: payload.GetEnvironment(),
				master:      payload.GetRequest().GetMasterPgInstance(),
			},
		},
		{
			Name: stepStageHaproxyConfig,
			Job: &stageHaproxyConfigJob{
				payload:        payload,
				dataStorage:    h.dataStorage,
				configResolver: h.configResolver,
				runtimes:       h.runtimes,
			},
		},
		{
			Name: stepCheckHaproxyConfig,
			Job:  &checkHaproxyConfigJob{payload: payload, runtimes: h.runtimes},
		},
		{
			Name: stepReloadHaproxy,
			Job:  &reloadHaproxyJob{payload: payload, runtimes: h.runtimes},
		},
		{
			Name: stepAwaitHaproxyStable,
			Job: &awaitHaproxyStableJob{
				runtimes:    h.runtimes,
				environment: payload.GetEnvironment(),
				containerId: payload.GetContainerId(),
			},
		},
	}
}

type stageHaproxyConfigJob struct {
	payload        *velez_api.SelectHaproxyMasterTaskPayload
	dataStorage    storage.Storage
	configResolver service.ServiceConfigResolver
	runtimes       container_runtime.RuntimeResolver
}

func (j *stageHaproxyConfigJob) Do(ctx context.Context) error {
	user, err := haproxyMasterUsername(ctx, j.dataStorage, j.payload.GetRequest().GetMasterPgInstance())
	if err != nil {
		return rerrors.Wrap(err, "error resolving HAProxy master username")
	}

	config, err := renderHaproxyConfig(j.payload.GetPort(), j.payload.GetRequest().GetMasterPgInstance(), user)
	if err != nil {
		return rerrors.Wrap(err)
	}

	runtime, err := j.runtimes.Runtime(ctx, j.payload.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving HAProxy container runtime")
	}

	// WriteFile attaches files to deploy requests. For a live update, apply
	// the resulting attachment to the instance's existing mounted volume.
	request := &velez_api.CreateSmerd_Request{}

	err = j.configResolver.WriteFile(ctx, request, haproxyPendingConfigPath, config)
	if err != nil {
		return rerrors.Wrap(err, "error preparing HAProxy config")
	}

	for _, file := range request.GetPlain() {
		err = runtime.CopyToContainer(ctx, j.payload.GetContainerId(), file.GetPath(), file.GetContent(), modeHaproxyConfig)
		if err != nil {
			return rerrors.Wrap(err, "error staging HAProxy config")
		}
	}

	return nil
}

type checkHaproxyConfigJob struct {
	payload  *velez_api.SelectHaproxyMasterTaskPayload
	runtimes container_runtime.RuntimeResolver
}

func (j *checkHaproxyConfigJob) Do(ctx context.Context) error {
	runtime, err := j.runtimes.Runtime(ctx, j.payload.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving HAProxy container runtime")
	}

	check := []string{"haproxy", "-c", "-f", haproxyPendingConfigPath}

	err = execHaproxyConfigCommand(ctx, runtime, j.payload.GetContainerId(), "", check)
	if err != nil {
		return rerrors.Wrap(err, "HAProxy config check failed")
	}

	return nil
}

type reloadHaproxyJob struct {
	payload  *velez_api.SelectHaproxyMasterTaskPayload
	runtimes container_runtime.RuntimeResolver
}

func (j *reloadHaproxyJob) Do(ctx context.Context) error {
	runtime, err := j.runtimes.Runtime(ctx, j.payload.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving HAProxy container runtime")
	}

	containerId := j.payload.GetContainerId()

	// Keep the validated candidate so a retry after installing the config can
	// repeat the atomic rename and signal the same container.
	install := []string{
		"sh", "-ec", `cp "$1" "$2"; mv -f "$2" "$3"`, "haproxy-config",
		haproxyPendingConfigPath, haproxyInstallConfigPath, domain.HaproxyConfigPath,
	}

	err = execHaproxyConfigCommand(ctx, runtime, containerId, "0", install)
	if err != nil {
		return rerrors.Wrap(err, "error installing HAProxy config")
	}

	err = runtime.Signal(ctx, containerId, haproxyReloadSignal)
	if err != nil {
		return rerrors.Wrap(err, "error reloading HAProxy")
	}

	return nil
}

func execHaproxyConfigCommand(
	ctx context.Context,
	runtime container_runtime.ContainerRuntime,
	containerId, user string,
	command []string,
) error {
	opts := container.ExecOptions{
		Cmd: command, User: user, AttachStdout: true, AttachStderr: true,
	}

	output, exitCode, err := runtime.Exec(ctx, containerId, opts)
	if err != nil {
		return rerrors.Wrap(err, "error executing HAProxy config command")
	}

	if exitCode != 0 {
		return user_errors.New(fmt.Sprintf("command exited with status %d: %s", exitCode, strings.TrimSpace(string(output))))
	}

	return nil
}
