package transfer_in

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"go.vervstack.ru/Velez/internal/storage"
)

const (
	progressPersistInterval = time.Second
)

type progressTracker struct {
	transfers   storage.TransfersStorage
	transferId  uuid.UUID
	totalBytes  int64
	doneBytes   int64
	lastSavedAt time.Time
}

func newProgressTracker(transfers storage.TransfersStorage, transferId uuid.UUID, totalBytes int64) *progressTracker {
	return &progressTracker{
		transfers:  transfers,
		transferId: transferId,
		totalBytes: totalBytes,
	}
}

func (p *progressTracker) add(ctx context.Context, bytes int) {
	p.doneBytes += int64(bytes)

	if time.Since(p.lastSavedAt) < progressPersistInterval {
		return
	}

	p.flush(ctx)
}

func (p *progressTracker) flush(ctx context.Context) {
	p.lastSavedAt = time.Now()

	err := p.transfers.UpdateProgress(ctx, p.transferId, p.totalBytes, p.doneBytes)
	if err != nil {
		log.Warn().
			Str("transfer_uuid", p.transferId.String()).
			Err(err).
			Msg("error persisting transfer progress")
	}
}
