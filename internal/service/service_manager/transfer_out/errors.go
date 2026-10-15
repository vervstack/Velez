package transfer_out

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	errContainerNotFound        = rerrors.New("container not found", codes.NotFound)
	errSidecarNotMovable        = rerrors.New("a sidecar cannot be transferred on its own", codes.FailedPrecondition)
	errUnsupportedMounts        = rerrors.New("container has host bind mounts", codes.FailedPrecondition)
	errAddressRequired          = rerrors.New("target address is required", codes.InvalidArgument)
	errTokenRequired            = rerrors.New("transfer token is required", codes.InvalidArgument)
	errUnknownSidecar           = rerrors.New("selected sidecar does not belong to the container", codes.InvalidArgument)
	errConsumersNotAcknowledged = rerrors.New("other containers use this network", codes.FailedPrecondition)
	errTransferNotFound         = rerrors.New("transfer is not running", codes.NotFound)
	errTargetNotAccepted        = rerrors.New("target did not accept the transfer", codes.Aborted)
	errUnexpectedFrame          = rerrors.New("unexpected frame from target", codes.Internal)
	errTargetClosedEarly        = rerrors.New("target closed the stream before finishing", codes.Aborted)
	errUnknownItemKind          = rerrors.New("unknown transfer item kind", codes.Internal)
)
