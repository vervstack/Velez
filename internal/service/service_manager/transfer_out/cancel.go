package transfer_out

import (
	"context"

	"github.com/google/uuid"
	"go.redsock.ru/rerrors"
)

func (s *Service) Cancel(_ context.Context, id uuid.UUID) error {
	cancel, isFound := s.lookupCancel(id)
	if !isFound {
		return rerrors.Wrap(errTransferNotFound)
	}

	cancel()

	return nil
}

func (s *Service) registerCancel(id uuid.UUID, cancel context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cancels[id] = cancel
}

func (s *Service) unregisterCancel(id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.cancels, id)
}

func (s *Service) lookupCancel(id uuid.UUID) (context.CancelFunc, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cancel, isFound := s.cancels[id]

	return cancel, isFound
}
