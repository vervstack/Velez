package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	// ErrElevatedAccessImageNotWhitelisted is returned by the container runtime
	// when a container requests elevated host access (privileged mode, extra
	// capabilities, devices, host namespaces, docker socket) and its image is
	// not on the developer-approved whitelist.
	ErrElevatedAccessImageNotWhitelisted = rerrors.New(
		"image is not allowed to request elevated host access", codes.PermissionDenied)
)
