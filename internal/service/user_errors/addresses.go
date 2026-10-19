package user_errors

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/codes"
)

var (
	// ErrAddressesRebuildInProgress is returned when an address registry
	// rebuild is requested while another one is still running.
	ErrAddressesRebuildInProgress = rerrors.New("addresses rebuild is already in progress", codes.AlreadyExists)
)
