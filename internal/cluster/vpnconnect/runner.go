// Package vpnconnect connects a node-level Verv service (matreshka, makosh)
// to the Verv Closed Network by launching its tailscale sidecar container.
//
// It is the one survivor of internal/pipelines. docs/jobs_migration.md
// deliberately keeps this flow off the jobs engine: both call sites
// (internal/cluster/configuration, internal/cluster/service_discovery) branch
// on typed sentinel errors - user_errors.ErrVpnResultAlreadyExists here and
// cluster_clients.ErrServiceIsDisabled - which Engine.Enqueue/Watch would
// flatten into an opaque error string. The pipeline runner and the steps this
// one flow needs were moved here verbatim when internal/pipelines was deleted;
// behavior is unchanged.
//
// internal/jobs/connect_service_to_vpn.go is the jobs-engine implementation of
// the same flow, used for user-triggered service connections over the API.
package vpnconnect

import (
	"context"
	"time"

	"go.redsock.ru/rerrors"
)

const (
	defaultRollbackTimeout = 30 * time.Second
)

//nolint:forbidigo // package-private sentinel, not shared/user-facing
var errNoGetResultFunction = rerrors.New("no get result function")

// Runner executes an ordered list of steps, rolling back the ones that
// support it when a later step fails.
type Runner[T any] interface {
	Run(ctx context.Context) error
	Result() (res *T, err error)
}

// step is one unit of work in a Runner. It mirrors the deleted
// internal/pipelines/steps.Step, and is deliberately identical in shape to
// jobs.Job (both are just Do(ctx) error).
type step interface {
	Do(ctx context.Context) error
}

// rollbackableStep is a step that can undo its own side effects.
type rollbackableStep interface {
	Rollback(ctx context.Context) error
}

type runner[T any] struct {
	steps     []step
	getResult func() (res *T, err error)
	stepIdx   int
}

func (p *runner[T]) Run(ctx context.Context) (err error) {
	runErr := p.run(ctx)
	if runErr == nil {
		return nil
	}

	err = rerrors.Wrap(runErr)

	rollbackCtx, cancel := context.WithTimeout(context.Background(), defaultRollbackTimeout)
	defer cancel()

	rollbackErr := p.rollback(rollbackCtx)
	if rollbackErr != nil {
		err = rerrors.Join(err, rerrors.Wrap(rollbackErr))
	}

	return err
}

func (p *runner[T]) Result() (res *T, err error) {
	if p.getResult != nil {
		return p.getResult()
	}

	return nil, rerrors.Wrap(errNoGetResultFunction)
}

func (p *runner[T]) run(ctx context.Context) error {
	var s step

	for p.stepIdx, s = range p.steps {
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
		rollbackable, ok := p.steps[p.stepIdx].(rollbackableStep)
		if ok {
			err := rollbackable.Rollback(ctx)
			if err != nil {
				globalErr = rerrors.Join(globalErr, rerrors.Wrapf(err, "error during rollback step: %v ", rollbackable))
			}
		}
	}

	return globalErr
}
