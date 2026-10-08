package jobs

import (
	"go.redsock.ru/rerrors"
)

var (
	errBuildkitNetworkMissing = rerrors.New("runner buildkit network does not exist")

	errBuildkitNotRunning = rerrors.New("runner buildkit container is not running")

	errBinfmtInstallFailed = rerrors.New("binfmt installer exited non-zero")

	errBinfmtNotFinished = rerrors.New("binfmt installer is still running")
)
