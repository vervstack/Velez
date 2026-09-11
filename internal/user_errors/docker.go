package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

// ErrNameIsTaken is returned by docker.HandleConflictMessage when a
// ContainerCreate 409 conflict is a container-name collision. Checked with
// rerrors.Is by internal/cluster/vpnconnect/steps.go to tolerate an existing
// container of the same name.
var ErrNameIsTaken = rerrors.New("container name is taken", codes.AlreadyExists)
