package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	// ErrContainerIsOnboardingLeftover is returned by RegisterContainer for a container that already has an
	// onboarded replacement and waits for FinishOnboarding.
	ErrContainerIsOnboardingLeftover = rerrors.New(
		"container was already onboarded, finish its onboarding instead", codes.FailedPrecondition)

	// ErrContainerNotOnboardingLeftover is returned by FinishOnboarding for a container no onboarded
	// replacement points at.
	ErrContainerNotOnboardingLeftover = rerrors.New(
		"container is not waiting for onboarding end", codes.FailedPrecondition)

	// ErrPortOccupied is returned by register_container when a requested host port is bound by a container
	// other than the one being registered.
	ErrPortOccupied = rerrors.New("host port is already occupied", codes.AlreadyExists)
)
