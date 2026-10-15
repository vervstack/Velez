package transfer_api_impl

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var errInvalidTransferId = rerrors.New("invalid transfer id", codes.InvalidArgument)
