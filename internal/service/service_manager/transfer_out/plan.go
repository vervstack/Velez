package transfer_out

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

func (s *Service) Plan(_ context.Context, _ string) (domain.TransferPlan, error) {
	return domain.TransferPlan{}, rerrors.Wrap(errNotImplemented)
}
