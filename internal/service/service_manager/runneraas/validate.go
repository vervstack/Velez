package runneraas

import (
	"strings"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/user_errors"
)

// validateRunnerTarget is a pure, table-tested check of the RunnerScope/
// target combination the request carries - REPO requires "owner/repo",
// ORG requires a bare "owner" with no slash. Provider-agnostic: every
// provider under this contract shares the same repo/org target shape.
func validateRunnerTarget(scope velez_api.RunnerScope, target string) error {
	if target == "" {
		return user_errors.ErrRunnerTargetEmpty
	}

	switch scope {
	case velez_api.RunnerScope_REPO:
		owner, repo, found := strings.Cut(target, "/")
		if !found || owner == "" || repo == "" || strings.Contains(repo, "/") {
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
