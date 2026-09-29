package velez_api_impl

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var errRegisterContainerNotImplemented = rerrors.New("register container is not implemented yet", codes.Unimplemented)
