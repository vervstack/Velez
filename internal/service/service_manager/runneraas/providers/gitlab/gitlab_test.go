package gitlab

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testRunnerRepoTarget = "owner/repo"
)

// fakeContainerRuntime implements container_runtime.ContainerRuntime,
// recording the exec call Register makes and returning the exit
// code/error a test case configures. Every method beyond Exec is unused by
// Register and panics if ever called.
type fakeContainerRuntime struct {
	execCalledWith container.ExecOptions
	execCalledID   string

	execOutput   []byte
	execExitCode int
	execErr      error
}

func (f *fakeContainerRuntime) Exec(
	_ context.Context, containerID string, cfg container.ExecOptions,
) ([]byte, int, error) {
	f.execCalledID = containerID
	f.execCalledWith = cfg

	return f.execOutput, f.execExitCode, f.execErr
}

func (f *fakeContainerRuntime) ContainerCreate(
	context.Context, container_runtime.ContainerCreateRequest,
) (container.CreateResponse, error) {
	panic("not implemented")
}

func (f *fakeContainerRuntime) ListContainers(
	context.Context, *velez_api.ListSmerds_Request,
) ([]container.Summary, error) {
	panic("not implemented")
}

func (f *fakeContainerRuntime) Remove(context.Context, string) error {
	panic("not implemented")
}

func (f *fakeContainerRuntime) Rename(context.Context, string, string) error {
	panic("not implemented")
}

func (f *fakeContainerRuntime) IsContainerRunning(context.Context, string) (bool, bool, error) {
	panic("not implemented")
}

func (f *fakeContainerRuntime) Inspect(context.Context, string) (container.InspectResponse, bool, error) {
	panic("not implemented")
}

func (f *fakeContainerRuntime) Stop(context.Context, string) error {
	panic("not implemented")
}

func (f *fakeContainerRuntime) Restart(context.Context, string) error {
	panic("not implemented")
}

func (f *fakeContainerRuntime) Stats(context.Context, string) (domain.ContainerStats, error) {
	panic("not implemented")
}

func (f *fakeContainerRuntime) CreateNetwork(context.Context, string) error {
	panic("not implemented")
}

func (f *fakeContainerRuntime) ConnectToNetwork(context.Context, container_runtime.ConnectToNetworkRequest) error {
	panic("not implemented")
}

func (f *fakeContainerRuntime) DisconnectFromNetworks(context.Context, string, []string) error {
	panic("not implemented")
}

func (f *fakeContainerRuntime) PullImage(context.Context, string) (image.InspectResponse, error) {
	panic("not implemented")
}

func (f *fakeContainerRuntime) ListOccupiedPorts(context.Context) ([]uint32, error) {
	panic("not implemented")
}

func Test_MintRegistrationToken_Scenarios(t *testing.T) {
	cases := []struct {
		name        string
		accessToken string
		want        string
		wantErr     error
	}{
		{"non-empty token is returned unchanged", "glpat-AABBCC", "glpat-AABBCC", nil},
		{"empty token is rejected", "", "", user_errors.ErrGitlabAccessTokenEmpty},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			provider := New()

			got, err := provider.MintRegistrationToken(
				context.Background(), velez_api.RunnerScope_REPO, testRunnerRepoTarget, "", tc.accessToken,
			)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func Test_RegistrationEnv_ReturnsEmpty(t *testing.T) {
	provider := New()

	env := provider.RegistrationEnv(velez_api.RunnerScope_REPO, testRunnerRepoTarget, "", "runner-1", "token-1", nil)

	require.Empty(t, env)
}

func Test_Register_Scenarios(t *testing.T) {
	cases := []struct {
		name        string
		baseUrl     string
		dockerImage string
		wantUrl     string
		wantImage   string
	}{
		{"defaults base url and docker image when both empty", "", "", defaultBaseUrl, defaultDockerImage},
		{
			"uses the given base url and docker image",
			"https://gitlab.example.com", "golang:1",
			"https://gitlab.example.com", "golang:1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			provider := New()
			runtime := &fakeContainerRuntime{execExitCode: 0}

			err := execRegister(provider, runtime, tc.baseUrl, tc.dockerImage)
			require.NoError(t, err)

			require.Equal(t, "runner-container", runtime.execCalledID)
			require.Equal(t, []string{
				"gitlab-runner", "register",
				"--non-interactive",
				"--url", tc.wantUrl,
				"--registration-token", "token-1",
				"--executor", "docker",
				"--docker-image", tc.wantImage,
				"--docker-privileged",
				"--description", "runner-1",
			}, runtime.execCalledWith.Cmd)
			require.True(t, runtime.execCalledWith.AttachStdout)
			require.True(t, runtime.execCalledWith.AttachStderr)
		})
	}
}

func Test_Register_NonZeroExitCode_ReturnsError(t *testing.T) {
	provider := New()
	runtime := &fakeContainerRuntime{execExitCode: 1}

	err := execRegister(provider, runtime, "", "")

	require.ErrorIs(t, err, user_errors.ErrGitlabRunnerRegisterFailed)
}

func Test_Register_ExecError_IsWrapped(t *testing.T) {
	provider := New()
	execErr := user_errors.ErrGitlabAccessTokenEmpty // any sentinel, just to assert wrapping
	runtime := &fakeContainerRuntime{execErr: execErr}

	err := execRegister(provider, runtime, "", "")

	require.ErrorIs(t, err, execErr)
}

// execRegister calls Register with fixed token/runner-name/container-id
// arguments, varying only baseUrl/dockerImage - shared by every Test_Register
// case above.
func execRegister(provider *Provider, runtime *fakeContainerRuntime, baseUrl, dockerImage string) error {
	return provider.Register(
		context.Background(), runtime, "runner-container", baseUrl, "token-1", dockerImage, "runner-1")
}
