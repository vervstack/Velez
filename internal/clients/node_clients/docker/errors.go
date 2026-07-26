package docker

import (
	"strings"

	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

const (
	NoSuchContainerError = "No such container"

	subjectContainerName = "The container name"

	problemInUseByOtherContainer = "is already in use by container"
)

var ErrNameIsTaken = rerrors.New("container name is taken", codes.AlreadyExists)

func handleConflictMessage(err error) error {
	msg := err.Error()

	if containsAll(msg, subjectContainerName, problemInUseByOtherContainer) {
		return ErrNameIsTaken
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
