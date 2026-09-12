package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	// ErrNoVervonomiconDescriptor is returned when a deploy explicitly
	// requested FromVervonomicon but the given image carries no .verv/
	// descriptor at all. Unlike GetVervonomicon (where absence of a
	// descriptor is never an error - see docs/features/vervonomicon.md), a
	// deploy that names this exact mechanism and finds nothing to resolve
	// genuinely has nothing to do.
	ErrNoVervonomiconDescriptor = rerrors.NewUserError(
		"image has no vervonomicon descriptor to deploy from",
		codes.FailedPrecondition,
	)

	// ErrVervonomiconDescriptorNotFound reports that a service simply has no
	// vervonomicon descriptor. Per docs/features/vervonomicon.md: "Absence of
	// a descriptor is never an error - the service simply deploys the way it
	// does today." Callers must be able to tell this apart from a descriptor
	// that exists but is broken - check with errors.Is. Returned by
	// vervonomicon.Parse and vervonomicon.ImageSource.Read.
	ErrVervonomiconDescriptorNotFound = rerrors.New("no vervonomicon descriptor found")
)
