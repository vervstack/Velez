package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	// ErrRunnerScopeUnspecified is returned by the runneraas package when a
	// request carries no RunnerScope.
	ErrRunnerScopeUnspecified = rerrors.New("runner scope is unspecified", codes.InvalidArgument)

	// ErrRunnerTargetEmpty is returned by the runneraas package when a
	// request carries an empty target.
	ErrRunnerTargetEmpty = rerrors.New("runner target is empty", codes.InvalidArgument)

	// ErrRunnerTargetInvalidRepoFormat is returned by the runneraas package
	// when a REPO-scoped target is not in "owner/repo" form.
	ErrRunnerTargetInvalidRepoFormat = rerrors.New(
		"runner target must be in \"owner/repo\" form for repo scope", codes.InvalidArgument)

	// ErrRunnerTargetInvalidOrgFormat is returned by the runneraas package
	// when an ORG-scoped target contains a "/".
	ErrRunnerTargetInvalidOrgFormat = rerrors.New(
		"runner target must be a bare org name for org scope", codes.InvalidArgument)

	// ErrGithubRegistrationTokenRequestFailed is returned by the runneraas
	// package's github provider when GitHub's registration-token endpoint
	// responds with an unexpected status code.
	ErrGithubRegistrationTokenRequestFailed = rerrors.New("github registration token request failed")

	// ErrGithubRegistrationTokenMissing is returned by the runneraas
	// package's github provider when GitHub's registration-token response
	// decodes but carries no token.
	ErrGithubRegistrationTokenMissing = rerrors.New("github registration token response carries no token")

	// ErrRunnerProviderUnsupported is returned by the runneraas package for
	// an unrecognized or unspecified RunnerProvider.
	ErrRunnerProviderUnsupported = rerrors.New("runner provider is not supported", codes.InvalidArgument)

	// ErrRunnerDockerSocketAddressInvalid is returned by the runneraas
	// package when a request's docker_socket_address is non-empty but not a
	// tcp:// address.
	ErrRunnerDockerSocketAddressInvalid = rerrors.New(
		"runner docker socket address must be a tcp:// address", codes.InvalidArgument)

	// ErrGitlabAccessTokenEmpty is returned by the runneraas package's
	// gitlab provider when a request carries an empty access token. v1
	// trusts the caller's access token as an already-valid registration
	// token (see the gitlab package doc comment), so this is the only
	// validation MintRegistrationToken performs.
	ErrGitlabAccessTokenEmpty = rerrors.New("gitlab access token is empty", codes.InvalidArgument)

	// ErrGitlabLegacyRegistrationToken is returned by the runneraas package
	// when a gitlab access token is a runner registration or authentication
	// token instead of a personal access token.
	ErrGitlabLegacyRegistrationToken = rerrors.New(
		"gitlab access token must be a personal access token with the create_runner scope, "+
			"not a runner registration or authentication token", codes.InvalidArgument)

	// ErrGitlabAccessTokenRejected is returned by the runneraas package's
	// gitlab provider when gitlab answers 401/403 to the personal access
	// token.
	ErrGitlabAccessTokenRejected = rerrors.New(
		"gitlab rejected the personal access token or it lacks the create_runner scope",
		codes.PermissionDenied)

	// ErrGitlabTargetNotFound is returned by the runneraas package's gitlab
	// provider when the requested project or group does not exist or is not
	// visible to the personal access token.
	ErrGitlabTargetNotFound = rerrors.New(
		"gitlab project or group not found or not visible to the token", codes.NotFound)

	// ErrGitlabRunnerRegisterFailed is returned by the runneraas package's
	// gitlab provider when `gitlab-runner register` exits non-zero inside
	// the deployed container.
	ErrGitlabRunnerRegisterFailed = rerrors.New("gitlab-runner register exited non-zero")

	// ErrGitlabRunnerUnregisterFailed is returned by the runneraas package's
	// gitlab provider when `gitlab-runner unregister --all-runners` exits
	// non-zero inside the deployed container.
	ErrGitlabRunnerUnregisterFailed = rerrors.New("gitlab-runner unregister exited non-zero")

	// ErrGitlabRunnerEntryCount is returned by the gitlab provider when
	// config.toml does not hold exactly one [[runners]] entry right after
	// `gitlab-runner register`.
	ErrGitlabRunnerEntryCount = rerrors.New("gitlab-runner config.toml must hold exactly one [[runners]] entry")

	// ErrRunnerConcurrentInvalid is returned by the runneraas package when a
	// concurrent value below 1 is requested.
	ErrRunnerConcurrentInvalid = rerrors.NewUserError("concurrent must be at least 1")

	// ErrRunnerSettingsInvalid is returned by the runneraas package when a
	// requested pull policy, log level or interval is not supported.
	ErrRunnerSettingsInvalid = rerrors.NewUserError(
		"runner settings are invalid: unsupported pull policy, log level or negative interval")

	// ErrRunnerNotRunning is returned by the runneraas package when a runner's
	// config file is needed but its container is absent or stopped.
	ErrRunnerNotRunning = rerrors.NewUserError("runner container is not running")

	// ErrRunnerBuildkitRequiresDind is returned by the runneraas package when
	// BuildKit is requested for a runner that is not a GitLab runner backed by
	// a DinD service: buildkitd runs inside that DinD daemon.
	ErrRunnerBuildkitRequiresDind = rerrors.New(
		"buildkit needs a DinD-backed GitLab runner", codes.FailedPrecondition)
)
