package transfer_in

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (s *Service) Receive(_ context.Context, _ velez_api.TransferAPI_TransferServer) error {
	return rerrors.Wrap(errNotImplemented)
}
