package transfer_out

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

func (s *TransferOutServiceImpl) List(ctx context.Context) ([]domain.Transfer, error) {
	transfers, err := s.dataStorage.Transfers().List(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing transfers")
	}

	return transfers, nil
}
