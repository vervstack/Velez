package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/parser"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/patterns"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	BootstrapVcnSidecarAction = "bootstrap_vcn_sidecar"
	sidecarHealthTimeout      = 45 * time.Second
	sidecarHealthInterval     = time.Second
)

type bootstrapVcnSidecarHandler struct {
	nodeClients node_clients.NodeClients
	runtimes    container_runtime.RuntimeResolver
}

func NewBootstrapVcnSidecarHandler(
	nodeClients node_clients.NodeClients, runtimes container_runtime.RuntimeResolver,
) TaskHandler {
	return &bootstrapVcnSidecarHandler{nodeClients: nodeClients, runtimes: runtimes}
}

func (h *bootstrapVcnSidecarHandler) Action() string { return BootstrapVcnSidecarAction }

func (h *bootstrapVcnSidecarHandler) NewContext() TaskContext {
	return &velez_api.BootstrapVcnSidecarTaskPayload{}
}

func (h *bootstrapVcnSidecarHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.BootstrapVcnSidecarTaskPayload)
	if !ok {
		panic("bootstrap_vcn_sidecar: BuildJobs called with mismatched TaskContext type")
	}

	req := payload.GetRequest()
	serviceName := domain.HaproxyServiceName(req.GetName())
	name := haproxySidecarName(req.GetName())
	hostname := req.GetHostname()

	if hostname == "" {
		hostname = strings.ReplaceAll(name, "_", "-")
	}

	launch := patterns.TailScaleContainerSidecar(serviceName)

	launch.Labels[labels.HaproxySidecarLabel] = req.GetName()
	launch.HostConfig.NetworkMode = "bridge"
	launch.HostConfig.RestartPolicy.Name = container.RestartPolicyUnlessStopped
	launch.Env = append(launch.Env, "TS_HOSTNAME="+hostname, "TS_AUTHKEY="+req.GetAuthKey(),
		"TS_ACCEPT_DNS=true", "TS_EXTRA_ARGS=--login-server="+req.GetLoginServerUrl())

	settings := &velez_api.Container_Settings{Ports: req.GetPorts()}

	launch.HostConfig.PortBindings = parser.FromPorts(settings)
	launch.ExposedPorts = parser.ToExposedPorts(launch.HostConfig.PortBindings)

	return []NamedJob{
		{Name: stepPrepareSidecarImage, Job: &prepareSidecarImageJob{runtimes: h.runtimes, req: &launch}},
		{Name: "create_container", Job: &createBootstrapSidecarJob{
			runtimes: h.runtimes, payload: payload, request: &launch, name: name,
			portManager: h.nodeClients.PortManagerContainer(), ports: haproxyPublishedPorts(launch.HostConfig.PortBindings),
		}},
		{Name: "connect_tailnet", Job: &startSidecarContainerJob{
			dockerAPI: h.nodeClients.Docker().Client(), ctx: payload,
		}},
		{Name: "wait_tailnet_healthy", Job: &waitBootstrapSidecarJob{runtimes: h.runtimes, payload: payload}},
	}
}

type createBootstrapSidecarJob struct {
	runtimes       container_runtime.RuntimeResolver
	payload        containerIDAccessor
	request        *container.CreateRequest
	name           string
	portManager    node_clients.PortManager
	ports          []uint32
	hasLockedPorts bool
}

func (j *createBootstrapSidecarJob) Do(ctx context.Context) error {
	runtime, err := j.runtimes.Runtime(ctx, "")
	if err != nil {
		return rerrors.Wrap(err)
	}

	existing, isFound, err := runtime.Inspect(ctx, j.name)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if isFound {
		hasSamePorts := maps.EqualFunc(
			existing.HostConfig.PortBindings, j.request.HostConfig.PortBindings, slices.Equal[[]nat.PortBinding],
		)
		if existing.Config.Labels[labels.HaproxySidecarLabel] != j.request.Labels[labels.HaproxySidecarLabel] ||
			!containsSidecarEnv(existing.Config.Env, j.request.Env) || !hasSamePorts {
			return user_errors.New("A VCN sidecar already exists with different settings")
		}

		j.payload.SetContainerId(existing.ID)

		return nil
	}

	err = j.portManager.LockPort(j.ports...)
	if err != nil {
		return rerrors.Wrap(err, "error reserving VCN sidecar ports")
	}

	j.hasLockedPorts = true

	config := &container_runtime.ContainerConfig{Config: j.request.Config}
	host := &container_runtime.HostConfig{HostConfig: j.request.HostConfig}
	req := container_runtime.ContainerCreateRequest{Config: config, HostConfig: host, ContainerName: j.name}

	created, err := runtime.ContainerCreate(ctx, req)
	if err != nil {
		return rerrors.Wrap(err, "error creating VCN sidecar")
	}

	j.payload.SetContainerId(created.ID)

	return nil
}

func (j *createBootstrapSidecarJob) Rollback(ctx context.Context) error {
	runtime, err := j.runtimes.Runtime(ctx, "")
	if err != nil {
		return rerrors.Wrap(err)
	}

	id := j.payload.GetContainerId()
	if id != "" {
		err = runtime.Remove(ctx, id)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	if id != "" || j.hasLockedPorts {
		j.portManager.UnlockPorts(j.ports)
	}

	return nil
}

type waitBootstrapSidecarJob struct {
	runtimes container_runtime.RuntimeResolver
	payload  containerIDAccessor
}

func (j *waitBootstrapSidecarJob) Do(ctx context.Context) error {
	runtime, err := j.runtimes.Runtime(ctx, "")
	if err != nil {
		return rerrors.Wrap(err)
	}

	ctx, cancel := context.WithTimeout(ctx, sidecarHealthTimeout)
	defer cancel()

	ticker := time.NewTicker(sidecarHealthInterval)
	defer ticker.Stop()

	for {
		if isTailnetHealthy(ctx, runtime, j.payload.GetContainerId()) {
			return nil
		}

		select {
		case <-ctx.Done():
			return user_errors.New("The VCN sidecar did not join the tailnet. Check the login server and auth key.")
		case <-ticker.C:
		}
	}
}

func isTailnetHealthy(ctx context.Context, runtime container_runtime.ContainerRuntime, id string) bool {
	isRunning, _, err := runtime.IsContainerRunning(ctx, id)
	if err != nil || !isRunning {
		return false
	}

	opts := container.ExecOptions{Cmd: []string{"tailscale", "status", "--json"}, AttachStdout: true}
	output, code, err := runtime.Exec(ctx, id, opts)

	if err != nil || code != 0 {
		return false
	}

	var status struct {
		BackendState string   `json:"BackendState"`
		TailscaleIps []string `json:"TailscaleIPs"`
	}

	// Runtime.Exec strips control bytes after Docker's stdout framing, leaving a prefix before the first brace.
	start := bytes.IndexByte(output, '{')
	if start < 0 {
		return false
	}

	err = json.Unmarshal(output[start:], &status)

	return err == nil && status.BackendState == "Running" && len(status.TailscaleIps) > 0
}

func containsSidecarEnv(actual, expected []string) bool {
	for _, value := range expected {
		if !slices.Contains(actual, value) {
			return false
		}
	}

	return true
}

func BootstrapSidecarResult(task tasks_queries.VelezTask) (string, error) {
	var payload velez_api.BootstrapVcnSidecarTaskPayload

	err := json.Unmarshal(task.Context.RawMessage, &payload)
	if err != nil {
		return "", rerrors.Wrap(err, "error decoding VCN bootstrap task")
	}

	if payload.GetContainerId() == "" {
		return "", rerrors.Wrap(user_errors.ErrContainerIdMissing)
	}

	return payload.GetContainerId(), nil
}

func haproxySidecarName(name string) string {
	return domain.HaproxyServiceName(name) + "-" + patterns.TailscaleSidecarSuffix
}

func haproxyPublishedPorts(bindings map[nat.Port][]nat.PortBinding) []uint32 {
	var ports []uint32

	for _, port := range parser.ToPortsMapping(bindings, nil) {
		if value := port.GetExposedTo(); value != 0 && !slices.Contains(ports, value) {
			ports = append(ports, value)
		}
	}

	return ports
}
