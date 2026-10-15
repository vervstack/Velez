package transfer_in

import (
	"context"

	"github.com/rs/zerolog/log"

	"go.vervstack.ru/Velez/internal/storage/postgres/generated/transfers_queries"
)

// abortTransfer removes what the failed run created and records the failure.
// Loaded images stay: they are harmless and may be shared with other containers.
func (s *Service) abortTransfer(ctx context.Context, run *receiveRun, cause error) {
	cleanupCtx := context.WithoutCancel(ctx)

	log.Error().
		Str("transfer_uuid", run.transferUuid.String()).
		Err(cause).
		Msg("incoming container transfer failed")

	for _, id := range run.createdContainerIds {
		err := run.runtime.Remove(cleanupCtx, id)
		if err != nil {
			log.Warn().
				Str("container_id", id).
				Err(err).
				Msg("error removing container of failed transfer")
		}
	}

	for _, name := range run.createdVolumes {
		err := run.runtime.RemoveVolume(cleanupCtx, name)
		if err != nil {
			log.Warn().
				Str("volume_name", name).
				Err(err).
				Msg("error removing volume of failed transfer")
		}
	}

	err := s.dataStorage.Transfers().Finish(
		cleanupCtx, run.transferUuid, transfers_queries.VelezTransferStatusFAILED, cause.Error())
	if err != nil {
		log.Warn().
			Str("transfer_uuid", run.transferUuid.String()).
			Err(err).
			Msg("error recording failed transfer")
	}
}
