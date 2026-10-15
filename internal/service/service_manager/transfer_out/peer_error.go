package transfer_out

import (
	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/status"
)

// peerError carries the gRPC code of a target's failure through rerrors, which would
// otherwise reduce a foreign status error to Internal.
func peerError(err error, msg string) error {
	return rerrors.Wrap(err, msg, status.Code(err))
}
