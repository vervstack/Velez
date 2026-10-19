package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	// ErrNameIsTaken is returned by docker.HandleConflictMessage when a
	// ContainerCreate 409 conflict is a container-name collision. Checked with
	// rerrors.Is by internal/cluster/vpnconnect/steps.go to tolerate an existing
	// container of the same name.
	ErrNameIsTaken = rerrors.New("container name is taken", codes.AlreadyExists)

	// ErrContainerConfigRequired names the invariant that a
	// ContainerCreateRequest's Config/Config.Config must be set - kept here
	// as the sanctioned error for any future caller-side validation.
	ErrContainerConfigRequired = rerrors.New("container config is required")

	// ErrDirReadSizeExceeded is returned by dockerutils.ReadFromContainer
	// when a container's directory contents exceed the maximum allowed
	// in-memory read size.
	ErrDirReadSizeExceeded = rerrors.New("directory contents exceed the maximum allowed read size")
)
