package transfer_out

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	progressInterval = time.Second
)

type progress struct {
	doneBytes atomic.Int64
}

func (s *TransferOutServiceImpl) startProgressReport(
	ctx context.Context,
	id uuid.UUID,
	totalBytes int64,
	counter *progress,
) context.CancelFunc {
	reportCtx, stop := context.WithCancel(ctx)

	go s.reportProgress(reportCtx, id, totalBytes, counter)

	return stop
}

func (s *TransferOutServiceImpl) reportProgress(
	ctx context.Context, id uuid.UUID, totalBytes int64, counter *progress,
) {
	ticker := time.NewTicker(progressInterval)
	defer ticker.Stop()

	var reported int64

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		current := counter.doneBytes.Load()
		if current == reported {
			continue
		}

		err := s.dataStorage.Transfers().UpdateProgress(ctx, id, totalBytes, current)
		if err != nil {
			log.Warn().
				Str("transfer_uuid", id.String()).
				Err(err).
				Msg("error persisting transfer progress")

			continue
		}

		reported = current
	}
}
