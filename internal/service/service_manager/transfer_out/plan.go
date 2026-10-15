package transfer_out

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

func (s *Service) Plan(ctx context.Context, containerName string) (domain.TransferPlan, error) {
	res, err := s.resolve(ctx, containerName)
	if err != nil {
		return domain.TransferPlan{}, rerrors.Wrap(err)
	}

	return res.toPlan(), nil
}
