package transfer_out

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/google/uuid"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func (s *Service) Start(ctx context.Context, req domain.StartTransferReq) (uuid.UUID, error) {
	err := validateRequest(req)
	if err != nil {
		return uuid.Nil, rerrors.Wrap(err)
	}

	res, err := s.resolve(ctx, req.ContainerName)
	if err != nil {
		return uuid.Nil, rerrors.Wrap(err)
	}

	selected, err := selectSidecars(res.sidecars, req.SidecarNames)
	if err != nil {
		return uuid.Nil, rerrors.Wrap(err)
	}

	err = checkTransferable(res.root, selected, res.consumers, req.IsConsumersAcknowledged)
	if err != nil {
		return uuid.Nil, rerrors.Wrap(err)
	}

	begin, err := s.buildBegin(ctx, req.Token, res.root, selected)
	if err != nil {
		return uuid.Nil, rerrors.Wrap(err)
	}

	// The transfer outlives the request, so its context keeps only the request's values.
	transferCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))

	sess, err := openSession(transferCtx, cancel, req.Address, begin)
	if err != nil {
		cancel()

		return uuid.Nil, rerrors.Wrap(err)
	}

	return s.launch(ctx, transferCtx, sess, req, res, selected, begin)
}

func (s *Service) launch(
	ctx context.Context,
	transferCtx context.Context,
	sess *session,
	req domain.StartTransferReq,
	res resolved,
	selected []container.InspectResponse,
	begin *velez_api.Transfer_Begin,
) (uuid.UUID, error) {
	total := totalBytes(begin.GetItems())

	transfer, err := s.recordTransfer(ctx, req, total)
	if err != nil {
		sess.close()

		return uuid.Nil, rerrors.Wrap(err)
	}

	job := transferJob{
		id:         transfer.Uuid,
		totalBytes: total,
		items:      begin.GetItems(),
		order:      newSourceOrder(res.root, selected),
		runtime:    res.runtime,
		session:    sess,
	}

	s.registerCancel(job.id, sess.cancel)

	go s.run(transferCtx, job)

	return job.id, nil
}
