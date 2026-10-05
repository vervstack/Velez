package container_runtime

import (
	"go.redsock.ru/rerrors"
)

var errDaemonHostUnknown = rerrors.New("docker daemon host is unknown")
