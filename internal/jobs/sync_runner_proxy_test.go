package jobs

import (
	"context"
	"io"
	"io/fs"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

var errDockerDown = rerrors.New("docker is down")

const (
	testProxyRunnerName   = "gitlab_runner_Artel"
	testProxyRunnerUrl    = "socks5://192.168.1.44:1080"
	testProxyRunnerConfig = "[[runners]]\n  name = \"registered\"\n  [runners.docker]\n    image = \"alpine\"\n"
)

// proxyRunnerRuntime serves one inspected container and one config.toml on top of fakeContainerRuntime.
type proxyRunnerRuntime struct {
	*fakeContainerRuntime

	info container.InspectResponse

	inspectErr error
	config     []byte
	isWritten  bool
}

func (r *proxyRunnerRuntime) Inspect(context.Context, string) (container.InspectResponse, bool, error) {
	if r.inspectErr != nil {
		return container.InspectResponse{}, false, r.inspectErr
	}

	return r.info, true, nil
}

func (r *proxyRunnerRuntime) CopyFromContainer(context.Context, string, string) ([]byte, error) {
	return r.config, nil
}

func (r *proxyRunnerRuntime) CopyToContainer(_ context.Context, _, _ string, content []byte, _ fs.FileMode) error {
	r.config = content
	r.isWritten = true

	return nil
}

type proxyRunnerResolver struct {
	runtime container_runtime.ContainerRuntime
}

func (r proxyRunnerResolver) Runtime(context.Context, string) (container_runtime.ContainerRuntime, error) {
	return r.runtime, nil
}

func (r proxyRunnerResolver) NestedRuntime(
	context.Context, string, string,
) (container_runtime.ContainerRuntime, io.Closer, error) {
	return nil, nil, rerrors.Wrap(errNestedRuntimeNotFaked)
}

func newSyncRunnerProxyJob(runtime *proxyRunnerRuntime) *syncRunnerProxyJob {
	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: &velez_api.UpgradeSmerd_Request{Name: testProxyRunnerName},
	}

	return &syncRunnerProxyJob{
		runtimes:      proxyRunnerResolver{runtime: runtime},
		req:           payload,
		containerName: testProxyRunnerName + newContainerSuffix,
	}
}

func newProxyRunnerRuntime(containerLabels map[string]string, containerEnv []string) *proxyRunnerRuntime {
	return &proxyRunnerRuntime{
		fakeContainerRuntime: &fakeContainerRuntime{},
		info: container.InspectResponse{
			Config: &container.Config{Labels: containerLabels, Env: containerEnv},
		},
		config: []byte(testProxyRunnerConfig),
	}
}

func gitlabRunnerLabels() map[string]string {
	return map[string]string{
		labels.RunnerInstanceLabel: labelTrueValue,
		labels.RunnerProviderLabel: velez_api.RunnerProvider_GITLAB.String(),
	}
}

func Test_SyncRunnerProxyJob_GitlabRunner_WritesJobEnvironment(t *testing.T) {
	runtime := newProxyRunnerRuntime(gitlabRunnerLabels(), []string{"HTTPS_PROXY=" + testProxyRunnerUrl})

	err := newSyncRunnerProxyJob(runtime).Do(context.Background())

	require.NoError(t, err)
	require.True(t, runtime.isWritten)
	require.Contains(t, string(runtime.config), "HTTP_PROXY="+testProxyRunnerUrl)
}

func Test_SyncRunnerProxyJob_NotARunner_WritesNothing(t *testing.T) {
	runtime := newProxyRunnerRuntime(map[string]string{}, []string{"HTTPS_PROXY=" + testProxyRunnerUrl})

	err := newSyncRunnerProxyJob(runtime).Do(context.Background())

	require.NoError(t, err)
	require.False(t, runtime.isWritten)
}

func Test_SyncRunnerProxyJob_InspectFails_ReturnsError(t *testing.T) {
	runtime := newProxyRunnerRuntime(gitlabRunnerLabels(), nil)

	runtime.inspectErr = errDockerDown

	err := newSyncRunnerProxyJob(runtime).Do(context.Background())

	require.ErrorIs(t, err, errDockerDown)
	require.False(t, runtime.isWritten)
}
