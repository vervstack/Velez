package local_storage

import (
	"go.redsock.ru/rerrors"
)

var (
	errBindingsUnsupported = rerrors.New("container bindings are not supported in single-node mode")

	errAddressResolverNotBound = rerrors.New("container address resolver is not bound to the registries storage")
)
