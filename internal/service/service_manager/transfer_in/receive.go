package transfer_in

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/transfers_queries"
)

func (s *TransferInServiceImpl) Receive(ctx context.Context, stream velez_api.TransferAPI_TransferServer) error {
	begin, err := s.readBegin(ctx, stream)
	if err != nil {
		return rerrors.Wrap(err)
	}

	manifest, err := parseManifest(begin)
	if err != nil {
		return rerrors.Wrap(err)
	}

	runtime, err := s.runtimes.Runtime(ctx, "")
	if err != nil {
		return rerrors.Wrap(err, "error resolving runtime")
	}

	err = s.preflight(ctx, runtime, manifest)
	if err != nil {
		return rerrors.Wrap(err)
	}

	transfer, err := s.startTransfer(ctx, manifest)
	if err != nil {
		return rerrors.Wrap(err)
	}

	run := &receiveRun{
		transferUuid: transfer.Uuid,
		manifest:     manifest,
		runtime:      runtime,
	}

	err = s.runTransfer(ctx, stream, run)
	if err != nil {
		s.abortTransfer(ctx, run, err)

		return rerrors.Wrap(err)
	}

	return nil
}

func (s *TransferInServiceImpl) runTransfer(
	ctx context.Context,
	stream velez_api.TransferAPI_TransferServer,
	run *receiveRun,
) error {
	err := stream.Send(newAcceptedResponse())
	if err != nil {
		return rerrors.Wrap(err, "error sending accepted")
	}

	progress := newProgressTracker(s.dataStorage.Transfers(), run.transferUuid, run.manifest.totalBytes)

	err = s.receiveItems(ctx, stream, run, progress)
	if err != nil {
		return rerrors.Wrap(err)
	}

	progress.flush(ctx)

	err = s.createContainers(ctx, run)
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = s.dataStorage.Transfers().Finish(ctx, run.transferUuid, transfers_queries.VelezTransferStatusDONE, "")
	if err != nil {
		return rerrors.Wrap(err, "error finishing transfer")
	}

	err = stream.Send(newFinishedResponse())
	if err != nil {
		return rerrors.Wrap(err, "error sending finished")
	}

	return nil
}
