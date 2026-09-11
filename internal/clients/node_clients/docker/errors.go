package docker

import (
	"strings"

	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"

	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	NoSuchContainerError = "No such container"

	subjectContainerName = "The container name"

	problemInUseByOtherContainer = "is already in use by container"
)

// HandleConflictMessage maps a Docker 409 into Velez's own
// user_errors.ErrNameIsTaken when the conflict is a container-name
// collision, and into a wrapped internal error otherwise.
//
// Exported because container_runtime's implementations issue ContainerCreate
// against the raw Docker API themselves and must surface the exact same error
// as Docker.ContainerCreate always has.
func HandleConflictMessage(err error) error {
	msg := err.Error()

	if containsAll(msg, subjectContainerName, problemInUseByOtherContainer) {
		return user_errors.ErrNameIsTaken
	}

	return rerrors.Wrap(err, "unhandled error", codes.Internal)
}

func containsAll(msg string, subs ...string) bool {
	for _, s := range subs {
		if !strings.Contains(msg, s) {
			return false
		}
	}

	return true
}
