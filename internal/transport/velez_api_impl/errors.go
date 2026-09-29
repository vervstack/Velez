package velez_api_impl

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var errPatternNotImplemented = rerrors.New("registration pattern is not implemented yet", codes.Unimplemented)
