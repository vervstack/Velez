package transfer_io

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var errNotImplemented = rerrors.New("not implemented", codes.Unimplemented)
