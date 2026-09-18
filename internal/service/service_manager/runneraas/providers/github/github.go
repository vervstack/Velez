// Package github implements runneraas.Provider for GitHub Actions
// self-hosted runners - the first RunnerProvider. Ported from the
// now-superseded internal/jobs/enable_github_runner.go, which PR3 deletes
// once transport is recut onto RunnersAPI.
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/user_errors"
	"go.vervstack.ru/Velez/internal/utils/common"
)

const (
	githubApiBaseUrl = "https://api.github.com"

	descriptorName = "github_runner"

	// dataPath must match builtin/github_runner/deployment.yaml's volume
	// mount - the runner's persistent work/registration directory, so a
	// restart reuses cached registration instead of re-registering.
	dataPath = "/home/runner"

	envRunnerRepoUrl = "RUNNER_REPO_URL"
	envRunnerOrgUrl  = "RUNNER_ORG_URL"
	envRunnerToken   = "RUNNER_TOKEN"
	envRunnerName    = "RUNNER_NAME"
	envRunnerLabels  = "RUNNER_LABELS"
)

// Provider implements runneraas.Provider for GitHub Actions.
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

// MintRegistrationToken calls GitHub's REST API for a short-lived runner
// registration token, per
// https://docs.github.com/en/rest/actions/self-hosted-runners.
// MintRegistrationToken ignores baseUrl - GitHub has a single fixed API
// host (githubApiBaseUrl); the parameter only exists to satisfy
// runneraas.Provider for providers like GitLab that don't.
func (p *Provider) MintRegistrationToken(
	ctx context.Context, scope velez_api.RunnerScope, target, _, accessToken string,
) (string, error) {
	endpoint, err := registrationTokenEndpoint(scope, target)
	if err != nil {
		return "", rerrors.Wrap(err, "error building github registration token endpoint")
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return "", rerrors.Wrap(err, "error building github registration token request")
	}

	httpReq.Header.Set("Authorization", "token "+accessToken)
	httpReq.Header.Set("Accept", "application/vnd.github+json")

	//nolint:bodyclose // closed via common.CloseWithLog below
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", rerrors.Wrap(err, "error calling github registration token endpoint")
	}
	defer common.CloseWithLog(resp.Body.Close, "github registration token response body")

	if resp.StatusCode != http.StatusCreated {
		statusDetail := fmt.Sprintf("status: %d", resp.StatusCode)

		return "", rerrors.Wrap(user_errors.ErrGithubRegistrationTokenRequestFailed, statusDetail)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", rerrors.Wrap(err, "error reading github registration token response")
	}

	token, err := parseRegistrationTokenResponse(body)
	if err != nil {
		return "", rerrors.Wrap(err, "error parsing github registration token response")
	}

	return token, nil
}

// RegistrationEnv sets the actions-runner image's registration env vars -
// builtin/github_runner/deployment.yaml deliberately carries them all empty
// (no descriptor file ever contains a credential), so the caller overlays
// them at deploy time instead. baseUrl is ignored for the same reason as
// MintRegistrationToken.
func (p *Provider) RegistrationEnv(
	scope velez_api.RunnerScope, target, _, runnerName, registrationToken string, labels []string,
) map[string]string {
	env := make(map[string]string, 5) //nolint:mnd

	switch scope {
	case velez_api.RunnerScope_REPO:
		env[envRunnerRepoUrl] = "https://github.com/" + target
	case velez_api.RunnerScope_ORG:
		env[envRunnerOrgUrl] = "https://github.com/" + target
	case velez_api.RunnerScope_RUNNER_SCOPE_UNSPECIFIED:
	default:
	}

	env[envRunnerToken] = registrationToken
	env[envRunnerName] = runnerName
	env[envRunnerLabels] = strings.Join(labels, ",")

	return env
}

// RegisterCommand always returns "" - actions-runner self-registers from
// RegistrationEnv's env vars at container boot, there is no manual step.
func (p *Provider) RegisterCommand(_, _ string) string {
	return ""
}

// registrationTokenEndpoint builds the GitHub REST endpoint that mints a
// short-lived runner registration token.
func registrationTokenEndpoint(scope velez_api.RunnerScope, target string) (string, error) {
	switch scope {
	case velez_api.RunnerScope_REPO:
		return githubApiBaseUrl + "/repos/" + target + "/actions/runners/registration-token", nil
	case velez_api.RunnerScope_ORG:
		return githubApiBaseUrl + "/orgs/" + target + "/actions/runners/registration-token", nil
	case velez_api.RunnerScope_RUNNER_SCOPE_UNSPECIFIED:
		return "", user_errors.ErrRunnerScopeUnspecified
	default:
		return "", user_errors.ErrRunnerScopeUnspecified
	}
}

type registrationTokenResponse struct {
	Token string `json:"token"`
}

// parseRegistrationTokenResponse decodes GitHub's registration-token
// response body. Never called with an unbounded/untrusted-size body folded
// into an error message - a decode failure or a missing token only ever
// surfaces a sentinel, never the body content.
func parseRegistrationTokenResponse(body []byte) (string, error) {
	var parsed registrationTokenResponse

	err := json.Unmarshal(body, &parsed)
	if err != nil {
		return "", rerrors.Wrap(err, "error decoding github registration token response")
	}

	if parsed.Token == "" {
		return "", user_errors.ErrGithubRegistrationTokenMissing
	}

	return parsed.Token, nil
}
