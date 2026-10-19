package transfer_out

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/transfers_queries"
)

// transferJob is everything the transfer goroutine needs once Start has returned.
type transferJob struct {
	id         uuid.UUID
	totalBytes int64
	items      []*velez_api.Transfer_ItemManifest
	order      sourceOrder
	runtime    container_runtime.ContainerRuntime
	session    *session
}

// run owns the session until the transfer settles; the source containers stay stopped on success.
func (s *TransferOutServiceImpl) run(ctx context.Context, job transferJob) {
	defer s.unregisterCancel(job.id)
	defer job.session.close()

	counter := &progress{}
	stopReport := s.startProgressReport(ctx, job.id, job.totalBytes, counter)

	err := s.transferSource(ctx, job, counter)

	stopReport()
	s.settle(ctx, job, counter, err)
}

func (s *TransferOutServiceImpl) transferSource(ctx context.Context, job transferJob, counter *progress) error {
	err := stopSource(ctx, job.runtime, job.order)
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = s.streamItems(ctx, job.session, job.items, counter)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

// settle records the outcome. A cancelled ctx at this point means the user cancelled: the
// transfer's own cancel only fires after run returns.
func (s *TransferOutServiceImpl) settle(ctx context.Context, job transferJob, counter *progress, runErr error) {
	settleCtx := context.WithoutCancel(ctx)

	if runErr == nil {
		s.finishDone(settleCtx, job, counter)

		return
	}

	log.Error().
		Str("transfer_uuid", job.id.String()).
		Err(runErr).
		Msg("container transfer did not complete")

	restoreSource(settleCtx, job.runtime, job.order)

	status, text := failureOutcome(runErr, ctx.Err() != nil)
	s.finish(settleCtx, job.id, status, text)
}

func (s *TransferOutServiceImpl) finishDone(ctx context.Context, job transferJob, counter *progress) {
	err := s.dataStorage.Transfers().UpdateProgress(ctx, job.id, job.totalBytes, counter.doneBytes.Load())
	if err != nil {
		log.Warn().
			Str("transfer_uuid", job.id.String()).
			Err(err).
			Msg("error persisting final transfer progress")
	}

	s.finish(ctx, job.id, transfers_queries.VelezTransferStatusDONE, "")
}

func (s *TransferOutServiceImpl) finish(
	ctx context.Context,
	id uuid.UUID,
	status transfers_queries.VelezTransferStatus,
	errMessage string,
) {
	err := s.dataStorage.Transfers().Finish(ctx, id, status, errMessage)
	if err != nil {
		log.Error().
			Str("transfer_uuid", id.String()).
			Err(err).
			Msg("error finishing transfer")
	}
}
