package gitlab

import (
	"go.redsock.ru/rerrors"
)

var (
	errRunnerScopeUnsupported = rerrors.New(
		"gitlab runner scope is not supported, only project (REPO) and group (ORG) runners can be created")

	errUnexpectedStatus = rerrors.New("gitlab api responded with an unexpected status")

	errRunnerTokenMissing = rerrors.New("gitlab api response carries no runner authentication token")

	errTargetIdMissing = rerrors.New("gitlab api response carries no project or group id")
)
