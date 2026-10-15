package transfer_out

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/transfers_queries"
)

func (s *Service) recordTransfer(
	ctx context.Context,
	req domain.StartTransferReq,
	totalBytes int64,
) (domain.Transfer, error) {
	transfer := domain.Transfer{
		IsSource:      true,
		ContainerName: req.ContainerName,
		PeerAddress:   req.Address,
		Status:        transfers_queries.VelezTransferStatusRUNNING,
		TotalBytes:    totalBytes,
	}

	inserted, err := s.dataStorage.Transfers().InsertTransfer(ctx, transfer)
	if err != nil {
		return domain.Transfer{}, rerrors.Wrap(err, "error recording transfer")
	}

	return inserted, nil
}
