package transfer_in

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	errTokenInvalid = rerrors.New("transfer token is invalid or expired", codes.PermissionDenied)

	errBeginExpected        = rerrors.New("first frame must be begin", codes.InvalidArgument)
	errManifestEmpty        = rerrors.New("manifest has no containers", codes.InvalidArgument)
	errNoRootContainer      = rerrors.New("manifest has no root container", codes.InvalidArgument)
	errInvalidInspect       = rerrors.New("container inspect is invalid", codes.InvalidArgument)
	errDuplicateContainer   = rerrors.New("container is listed twice", codes.InvalidArgument)
	errUnknownNetworkOwner  = rerrors.New("network owner is not a root container of the manifest", codes.InvalidArgument)
	errBindMountUnsupported = rerrors.New("host bind mounts are not supported", codes.InvalidArgument)
	errInvalidItem          = rerrors.New("manifest item is invalid", codes.InvalidArgument)
	errDuplicateItem        = rerrors.New("manifest item is listed twice", codes.InvalidArgument)

	errContainerExists = rerrors.New("container already exists on this node", codes.AlreadyExists)
	errVolumeExists    = rerrors.New("volume already exists on this node", codes.AlreadyExists)

	errUnknownItem     = rerrors.New("item is not in the manifest", codes.InvalidArgument)
	errItemRepeated    = rerrors.New("item was already received", codes.InvalidArgument)
	errNoOpenItem      = rerrors.New("batch arrived without an open item", codes.InvalidArgument)
	errUnexpectedFrame = rerrors.New("unexpected frame", codes.InvalidArgument)

	errItemsIncomplete   = rerrors.New("stream ended before every item was applied", codes.FailedPrecondition)
	errStreamClosedEarly = rerrors.New("stream closed before end frame", codes.FailedPrecondition)

	errRootNotCreated = rerrors.New("root container was not created before its sidecar", codes.Internal)
)
