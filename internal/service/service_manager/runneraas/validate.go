package runneraas

import (
	"strings"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/user_errors"
)

// validateRunnerTarget is a pure, table-tested check of the RunnerScope/
// target combination the request carries - REPO requires at least one "/"
// with a non-empty segment on both sides ("owner/repo" for GitHub, and a
// GitLab nested "group/subgroup/project" path is equally valid), ORG
// requires a bare name with no slash. ORG doesn't allow a nested path today
// (a GitLab nested group as an ORG target) - out of scope for now.
func validateRunnerTarget(scope velez_api.RunnerScope, target string) error {
	if target == "" {
		return user_errors.ErrRunnerTargetEmpty
	}

	switch scope {
	case velez_api.RunnerScope_REPO:
		owner, repo, found := strings.Cut(target, "/")
		if !found || owner == "" || repo == "" {
			return user_errors.ErrRunnerTargetInvalidRepoFormat
		}

		return nil
	case velez_api.RunnerScope_ORG:
		if strings.Contains(target, "/") {
			return user_errors.ErrRunnerTargetInvalidOrgFormat
		}

		return nil
	case velez_api.RunnerScope_RUNNER_SCOPE_UNSPECIFIED:
		return user_errors.ErrRunnerScopeUnspecified
	default:
		return user_errors.ErrRunnerScopeUnspecified
	}
}

const (
	gitlabLegacyRegistrationTokenPrefix = "GR1348941"
	gitlabRunnerAuthTokenPrefix         = "glrt-"
)

// validateGitlabAccessToken rejects a runner registration or authentication
// token where a personal access token is required. Any other value is
// accepted - self-managed instances may use custom personal access token
// prefixes.
func validateGitlabAccessToken(token string) error {
	if token == "" {
		return user_errors.ErrGitlabAccessTokenEmpty
	}

	isLegacy := strings.HasPrefix(token, gitlabLegacyRegistrationTokenPrefix)
	isRunnerAuth := strings.HasPrefix(token, gitlabRunnerAuthTokenPrefix)

	if isLegacy || isRunnerAuth {
		return user_errors.ErrGitlabLegacyRegistrationToken
	}

	return nil
}

// validateDockerSocketAddress checks a request's docker_socket_address.
// Empty is fine - it means fallback to the default host socket grant. A
// non-empty value must be tcp:// - unix:// and bare filesystem paths are
// rejected so a caller can never hand Velez a bind-mount source that could
// be pointed at Velez's own host socket.
func validateDockerSocketAddress(addr string) error {
	if addr == "" {
		return nil
	}

	if !strings.HasPrefix(addr, "tcp://") {
		return user_errors.ErrRunnerDockerSocketAddressInvalid
	}

	return nil
}

// validateDockerSource enforces that exactly one of dindName and
// dockerSocketAddress is set.
func validateDockerSource(dindName, dockerSocketAddress string) error {
	isDindSet := dindName != ""
	isSocketSet := dockerSocketAddress != ""

	if isDindSet && isSocketSet {
		return user_errors.ErrRunnerDockerSourceAmbiguous
	}

	if !isDindSet && !isSocketSet {
		return user_errors.ErrRunnerDockerSourceRequired
	}

	return nil
}
