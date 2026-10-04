package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	// ErrRunnerDockerSourceRequired is returned by the runneraas package when
	// a request carries neither dind_name nor docker_socket_address.
	ErrRunnerDockerSourceRequired = rerrors.New(
		"runner requires either dind_name or docker_socket_address", codes.InvalidArgument)

	// ErrRunnerDockerSourceAmbiguous is returned by the runneraas package when
	// a request carries both dind_name and docker_socket_address.
	ErrRunnerDockerSourceAmbiguous = rerrors.New(
		"runner accepts only one of dind_name and docker_socket_address", codes.InvalidArgument)
)
