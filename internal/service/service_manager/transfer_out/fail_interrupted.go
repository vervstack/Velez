package transfer_out

import (
	"context"

	"go.redsock.ru/rerrors"
)

func (s *TransferOutServiceImpl) FailInterrupted(ctx context.Context) error {
	err := s.dataStorage.Transfers().FailRunning(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error failing interrupted transfers")
	}

	return nil
}
