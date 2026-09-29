package local_storage

import (
	"go.redsock.ru/rerrors"
)

var errBindingsUnsupported = rerrors.New("container bindings are not supported in single-node mode")
