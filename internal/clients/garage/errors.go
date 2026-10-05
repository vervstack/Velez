package garage

import (
	"go.redsock.ru/rerrors"
)

var (
	ErrNotFound   = rerrors.New("garage entity not found")
	ErrBadRequest = rerrors.New("garage rejected the request")
	ErrConflict   = rerrors.New("garage entity already exists")

	errUnexpectedStatus  = rerrors.New("garage returned an unexpected status")
	errBuildRequest      = rerrors.New("error building garage request")
	errDoRequest         = rerrors.New("error sending garage request")
	errDecodeResponse    = rerrors.New("error decoding garage response")
	errEncodeRequest     = rerrors.New("error encoding garage request")
	errContainerNotFound = rerrors.New("garage container not found")
)
