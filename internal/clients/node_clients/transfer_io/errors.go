package transfer_io

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	errVolumeExists     = rerrors.New("volume already exists", codes.AlreadyExists)
	errImageLoadFailed  = rerrors.New("image load failed", codes.Internal)
	errHelperNotStarted = rerrors.New("transfer helper container id is empty", codes.Internal)
)
