package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

// ErrNoVervonomiconDescriptor is returned when a deploy explicitly requested
// FromVervonomicon but the given image carries no .verv/ descriptor at all.
// Unlike GetVervonomicon (where absence of a descriptor is never an error -
// see docs/features/vervonomicon.md), a deploy that names this exact
// mechanism and finds nothing to resolve genuinely has nothing to do.
var ErrNoVervonomiconDescriptor = rerrors.NewUserError(
	"image has no vervonomicon descriptor to deploy from",
	codes.FailedPrecondition,
)
