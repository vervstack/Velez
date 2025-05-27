package pipelines

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/pipelines/steps"
)

type runner[T any] struct {
	Steps     []steps.Step
	getResult func() (res *T, err error)
	stepIdx   int
}

func (p *runner[T]) Run(ctx context.Context) (err error) {
	runErr := p.run(ctx)
	if runErr == nil {
		return nil
	}

	err = rerrors.Wrap(runErr)

	rollbackErr := p.rollback(ctx)
	if rollbackErr != nil {
		err = rerrors.Join(err, rerrors.Wrap(rollbackErr))
	}

	return err
}

func (p *runner[T]) Result() (res *T, err error) {
	return p.getResult()
}

func (p *runner[T]) run(ctx context.Context) error {
	var s steps.Step

	for p.stepIdx, s = range p.Steps {
		err := s.Do(ctx)
		if err != nil {
			return rerrors.Wrapf(err, "error during execution of step: %T", s)
		}

	}

	return nil
}

func (p *runner[T]) rollback(ctx context.Context) error {
	var globalErr error

	for ; p.stepIdx >= 0; p.stepIdx-- {
		rollbackable, ok := p.Steps[p.stepIdx].(steps.RollbackableStep)
		if ok {
			err := rollbackable.Rollback(ctx)
			if err != nil {
				globalErr = rerrors.Join(globalErr, rerrors.Wrap(err, "error during rollback step: %v ", rollbackable))
			}
		}
	}

	return globalErr
}
