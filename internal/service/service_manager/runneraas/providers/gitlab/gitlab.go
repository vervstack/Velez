// Package gitlab implements runneraas.Provider for GitLab. v1 deliberately
// stops short of GitHub Actions parity: the official gitlab/gitlab-runner
// image doesn't self-register from env vars alone (it needs
// `gitlab-runner register` run once to write config.toml before
// `gitlab-runner run` serves jobs), so this provider deploys a bare
// container and trusts the caller's supplied access token as an
// already-valid registration token - registration itself is a manual step
// the caller runs, using the command surfaced by
// runneraas.RunneraasService.GetRunnerCredentials. GitLab's own API for
// minting a registration token (the legacy shared-token reset endpoint) is
// deprecated since 16.0 and disabled by default on 17.0+/gitlab.com, so v1
// does not call it.
package gitlab

import (
	"context"
	"fmt"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	defaultBaseUrl = "https://gitlab.com"

	descriptorName = "gitlab_runner"

	// dataPath must match builtin/gitlab_runner/deployment.yaml's volume
	// mount - gitlab-runner's config/registration directory, so a manually
	// run `gitlab-runner register` survives a container restart.
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

// RegisterCommand returns the `gitlab-runner register` invocation a caller
// runs by hand against the deployed container - the step this v1 provider
// never automates (see the package doc comment). --executor and everything
// after it are left as placeholders for the operator to fill in for their
// own setup.
func (p *Provider) RegisterCommand(baseUrl, registrationToken string) string {
	base := baseUrl
	if base == "" {
		base = defaultBaseUrl
	}

	return fmt.Sprintf(
		"gitlab-runner register --non-interactive --url %s --registration-token %s "+
			"--executor docker --docker-image alpine:latest",
		base, registrationToken,
	)
}
