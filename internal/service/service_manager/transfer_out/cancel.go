package transfer_out

import (
	"context"

	"github.com/google/uuid"
	"go.redsock.ru/rerrors"
)

func (s *TransferOutServiceImpl) Cancel(_ context.Context, id uuid.UUID) error {
	cancel, isFound := s.lookupCancel(id)
	if !isFound {
		return rerrors.Wrap(errTransferNotFound)
	}

	cancel()

	return nil
}

func (s *TransferOutServiceImpl) registerCancel(id uuid.UUID, cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cancels[id] = cancel
}

func (s *TransferOutServiceImpl) unregisterCancel(id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.cancels, id)
}

func (s *TransferOutServiceImpl) lookupCancel(id uuid.UUID) (context.CancelFunc, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cancel, isFound := s.cancels[id]

	return cancel, isFound
}
