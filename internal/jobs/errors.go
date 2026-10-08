package jobs

import (
	"go.redsock.ru/rerrors"
)

var (
	errBuildkitNotRunning = rerrors.New("runner buildkit container is not running")

	errBinfmtInstallFailed = rerrors.New("binfmt installer exited non-zero")

	errBinfmtNotFinished = rerrors.New("binfmt installer is still running")
)
