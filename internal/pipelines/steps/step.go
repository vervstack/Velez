package steps

import (
	"context"

	"go.redsock.ru/rerrors"
)

var ErrAlreadyExists = rerrors.New("pipeline result already exists")

type Step interface {
	Do(ctx context.Context) error
}

type RollbackableStep interface {
	Rollback(ctx context.Context) error
}

type singleFunc struct {
	f func(ctx context.Context) error
}

func SingleFunc(f func(ctx context.Context) error) Step {
	return &singleFunc{
		f: f,
	}
}

func (s *singleFunc) Do(ctx context.Context) error {
	return s.f(ctx)
}
