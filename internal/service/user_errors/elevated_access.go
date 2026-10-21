package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	// ErrReservedSettingsNotAllowed is returned by the public API when a request carries bind mounts, elevated
	// access or the elevated-access allowance label: those are set only by server-internal code.
	ErrReservedSettingsNotAllowed = rerrors.New(
		"bind mounts, elevated access and the elevated access label cannot be set through the public API",
		codes.InvalidArgument,
	)

	// ErrWhitelistNotSupported is returned when the image whitelist is edited outside cluster mode.
	ErrWhitelistNotSupported = rerrors.New("image whitelist is managed only in cluster mode", codes.FailedPrecondition)

	// ErrWhitelistedImageIdInvalid is returned by RemoveWhitelistedImage for an id that is not a uuid.
	ErrWhitelistedImageIdInvalid = rerrors.New("whitelisted image id is invalid", codes.InvalidArgument)
)
