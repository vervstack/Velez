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

	// ErrGitlabRunnerRegisterFailed is returned by the runneraas package's
	// gitlab provider when `gitlab-runner register` exits non-zero inside
	// the deployed container.
	ErrGitlabRunnerRegisterFailed = rerrors.New("gitlab-runner register exited non-zero")
)
