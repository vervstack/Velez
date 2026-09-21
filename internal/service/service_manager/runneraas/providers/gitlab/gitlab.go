// Package gitlab implements runneraas.Provider for GitLab. The official
// gitlab/gitlab-runner image doesn't self-register from env vars alone (it
// needs `gitlab-runner register` run once to write config.toml before
// `gitlab-runner run` serves jobs), so this provider deploys a bare
// container and trusts the caller's supplied access token as an
// already-valid registration token, then execs the register command inside
// the container itself once it's deployed (see Register). GitLab's own API
// for minting a registration token (the legacy shared-token reset endpoint)
// is deprecated since 16.0 and disabled by default on 17.0+/gitlab.com, so
// this provider does not call it.
package gitlab

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	defaultBaseUrl = "https://gitlab.com"

	// defaultDockerImage is the docker executor's job image when the
	// request's GitlabConfig.docker_image is empty.
	defaultDockerImage = "alpine:latest"

	// registerExecutor is always "docker" - not caller-configurable, see
	// GitlabConfig.docker_image's doc comment.
	registerExecutor = "docker"

	descriptorName = "gitlab_runner"

	// dataPath must match builtin/gitlab_runner/deployment.yaml's volume
	// mount - gitlab-runner's config/registration directory, so a
	// registration written by Register survives a container restart.
	dataPath = "/etc/gitlab-runner"
)

// Provider implements runneraas.Provider for GitLab.
type Provider struct{}

// New builds a Provider.
func New() *Provider {
	return &Provider{}
}

func (p *Provider) DescriptorName() string {
	return descriptorName
}

func (p *Provider) DataPath() string {
	return dataPath
}

// MintRegistrationToken is a pass-through in v1: it trusts the caller's
// accessToken as an already-valid GitLab runner registration token and
// returns it unchanged. A future PR wires GitLab's real token-minting API
// instead of trusting the caller's raw input (see the package doc comment
// for why v1 doesn't call GitLab's API today).
func (p *Provider) MintRegistrationToken(
	_ context.Context, _ velez_api.RunnerScope, _, _, accessToken string,
) (string, error) {
	if accessToken == "" {
		return "", user_errors.ErrGitlabAccessTokenEmpty
	}

	return accessToken, nil
}

// RegistrationEnv returns no env vars - v1 has no automated registration
// step for GitLab (see the package doc comment), so the container is
// deployed bare and every argument here is unused.
func (p *Provider) RegistrationEnv(
	_ velez_api.RunnerScope, _, _, _, _ string, _ []string,
) map[string]string {
	return map[string]string{}
}

// Register execs `gitlab-runner register --non-interactive` inside
// containerID via runtime, finishing what the package doc comment describes
// - config.toml doesn't exist until this runs once. Fails on a non-zero
// exit code; the register command's own output is never folded into the
// returned error, since it can carry the caller's GitLab server's response
// text (untrusted external content).
func (p *Provider) Register(
	ctx context.Context, runtime container_runtime.ContainerRuntime,
	containerID, baseUrl, registrationToken, dockerImage, runnerName string,
) error {
	base := baseUrl
	if base == "" {
		base = defaultBaseUrl
	}

	image := dockerImage
	if image == "" {
		image = defaultDockerImage
	}

	execOpts := container.ExecOptions{
		Cmd: []string{
			"gitlab-runner", "register",
			"--non-interactive",
			"--url", base,
			"--registration-token", registrationToken,
			"--executor", registerExecutor,
			"--docker-image", image,
			"--description", runnerName,
		},
		AttachStdout: true,
		AttachStderr: true,
	}

	_, exitCode, err := runtime.Exec(ctx, containerID, execOpts)
	if err != nil {
		return rerrors.Wrap(err, "error executing gitlab-runner register")
	}

	if exitCode != 0 {
		return rerrors.Wrap(user_errors.ErrGitlabRunnerRegisterFailed)
	}

	return nil
}
