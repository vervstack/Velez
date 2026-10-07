package jobs

import (
	"context"
	"maps"
	"net"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/parser"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	s3LoopbackHost = "127.0.0.1"
)

var errGarageWebUiNotRunning = rerrors.New("garage web ui container is not running")

// garageWebUiSidecarSpec is everything the web ui sidecar container is built
// from. The sidecar shares the garage container's network namespace, so it
// publishes nothing itself - the garage container publishes the web ui port.
type garageWebUiSidecarSpec struct {
	containerName   string
	serviceName     string
	instanceName    string
	imageName       string
	rootContainerId string
	restart         *velez_api.RestartPolicy
	env             map[string]string
}

func newGarageWebUiSidecarRequest(spec garageWebUiSidecarSpec) container_runtime.ContainerCreateRequest {
	containerLabels := sidecarLabels(spec.serviceName)

	containerLabels[labels.S3WebUiLabel] = spec.instanceName
	containerLabels[labels.WebUiForLabel] = spec.serviceName
	containerLabels[labels.WebUiPortLabel] = strconv.Itoa(domain.S3WebUiContainerPort)
	containerLabels[labels.ComposeGroupLabel] = spec.serviceName

	config := &container.Config{
		Image:  spec.imageName,
		Env:    parser.FromDockerEnv(spec.env),
		Labels: containerLabels,
	}

	hostConfig := &container.HostConfig{
		NetworkMode:   container.NetworkMode(networkModeOfPrefix + spec.rootContainerId),
		RestartPolicy: parser.FromRestart(spec.restart),
	}

	configWrapper := &container_runtime.ContainerConfig{Config: config}
	hostConfigWrapper := &container_runtime.HostConfig{HostConfig: hostConfig}

	return container_runtime.ContainerCreateRequest{
		Config:        configWrapper,
		HostConfig:    hostConfigWrapper,
		ContainerName: spec.containerName,
	}
}

// garageWebUiEnv points the web ui at garage over loopback: the sidecar shares
// garage's network namespace, so there is no DNS name to resolve.
func garageWebUiEnv(region string, secretEnv map[string]string) map[string]string {
	env := map[string]string{
		envWebUiApiBaseUrl:    s3LoopbackUrl(domain.S3AdminContainerPort),
		envWebUiS3EndpointUrl: s3LoopbackUrl(domain.S3ApiContainerPort),
		envWebUiS3Region:      region,
	}

	maps.Copy(env, secretEnv)

	return env
}

func s3LoopbackUrl(port int) string {
	return "http://" + net.JoinHostPort(s3LoopbackHost, strconv.Itoa(port))
}

type deployGarageWebUiJob struct {
	boxes    vervonomicon.BoxLookup
	runtimes container_runtime.RuntimeResolver
	secrets  secrets.Store
	payload  *velez_api.CreateS3InstanceTaskPayload
}

func (j *deployGarageWebUiJob) Do(ctx context.Context) error {
	request := j.payload.GetRequest()
	name := request.GetName()
	serviceName := domain.S3ServiceName(name)
	webUiName := domain.S3WebUiServiceName(name)

	_, smerdRequest, err := resolveS3Descriptor(
		ctx, j.boxes, s3WebUiDescriptorName, request.GetEnvironment(), "", nil, nil, nil,
	)
	if err != nil {
		return rerrors.Wrap(err, "error building garage web ui deploy request")
	}

	runtime, err := j.runtimes.Runtime(ctx, request.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	_, err = runtime.PullImage(ctx, smerdRequest.GetImageName())
	if err != nil {
		return rerrors.Wrap(err, "error pulling garage web ui image")
	}

	root, found, err := runtime.Inspect(ctx, serviceName)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting garage container")
	}

	if !found {
		return rerrors.Wrap(user_errors.ErrRegisterContainerNotFound, serviceName)
	}

	secretEnv, err := j.secretEnv(ctx, name)
	if err != nil {
		return err
	}

	err = j.dropLeftover(ctx, runtime, webUiName, serviceName)
	if err != nil {
		return err
	}

	spec := garageWebUiSidecarSpec{
		containerName:   webUiName,
		serviceName:     serviceName,
		instanceName:    name,
		imageName:       smerdRequest.GetImageName(),
		rootContainerId: root.ID,
		restart:         smerdRequest.GetRestart(),
		env:             garageWebUiEnv(s3Region(request), secretEnv),
	}

	createReq := newGarageWebUiSidecarRequest(spec)

	created, err := runtime.ContainerCreate(ctx, createReq)
	if err != nil {
		return rerrors.Wrap(err, "error creating garage web ui sidecar")
	}

	err = runtime.Restart(ctx, created.ID)
	if err != nil {
		return rerrors.Wrap(err, "error starting garage web ui sidecar")
	}

	return nil
}

func (j *deployGarageWebUiJob) Rollback(ctx context.Context) error {
	request := j.payload.GetRequest()

	runtime, err := j.runtimes.Runtime(ctx, request.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	err = runtime.Remove(ctx, domain.S3WebUiServiceName(request.GetName()))
	if err != nil {
		return rerrors.Wrap(err, "error removing garage web ui sidecar")
	}

	return nil
}

// dropLeftover clears a sidecar an earlier, interrupted run of this job
// created, so a resumed task does not collide with its own container.
func (j *deployGarageWebUiJob) dropLeftover(
	ctx context.Context,
	runtime container_runtime.ContainerRuntime,
	webUiName, serviceName string,
) error {
	existing, found, err := runtime.Inspect(ctx, webUiName)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting garage web ui sidecar")
	}

	if !found || !isRegisteredSidecar(existing.Config.Labels, serviceName) {
		return nil
	}

	err = runtime.Remove(ctx, webUiName)
	if err != nil {
		return rerrors.Wrap(err, "error removing leftover garage web ui sidecar")
	}

	return nil
}

func (j *deployGarageWebUiJob) secretEnv(ctx context.Context, name string) (map[string]string, error) {
	adminToken, err := j.secrets.Get(ctx, domain.S3AdminTokenSecretRef(name))
	if err != nil {
		return nil, rerrors.Wrap(err, "error reading admin token")
	}

	password, err := j.secrets.Get(ctx, domain.S3WebUiPasswordSecretRef(name))
	if err != nil {
		return nil, rerrors.Wrap(err, "error reading web ui password")
	}

	line, err := htpasswdLine(domain.S3WebUiUsername, password)
	if err != nil {
		return nil, rerrors.Wrap(err, "error hashing web ui password")
	}

	secretEnv := map[string]string{
		envWebUiApiAdminKey:  adminToken,
		envWebUiAuthUserPass: strings.TrimSuffix(line, "\n"),
	}

	return secretEnv, nil
}

type waitGarageWebUiJob struct {
	runtimes container_runtime.RuntimeResolver
	payload  *velez_api.CreateS3InstanceTaskPayload
}

func (j *waitGarageWebUiJob) Do(ctx context.Context) error {
	request := j.payload.GetRequest()
	webUiName := domain.S3WebUiServiceName(request.GetName())

	runtime, err := j.runtimes.Runtime(ctx, request.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	probe := garageWebUiProbe{runtime: runtime, name: webUiName}

	return pollGarage(ctx, probe.check, "timed out waiting for garage web ui to run")
}

type garageWebUiProbe struct {
	runtime container_runtime.ContainerRuntime
	name    string
}

func (p garageWebUiProbe) check(ctx context.Context) error {
	isRunning, found, err := p.runtime.IsContainerRunning(ctx, p.name)
	if err != nil {
		return rerrors.Wrap(err, "error checking container state")
	}

	if !found || !isRunning {
		return rerrors.Wrap(errGarageWebUiNotRunning)
	}

	return nil
}
