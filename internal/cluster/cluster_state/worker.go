package cluster_state

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/state"
)

const (
	defaultRetries    = 3
	defaultRetryDelay = time.Second * 6
)

func SetupWorkerPg(ctx context.Context, workerNodeDsn string, sm cluster_clients.ClusterStateManagerContainer) {
	pg, err := state.NewPgStateManager(ctx, workerNodeDsn)
	if err == nil {
		log.Info().Msg("Successfully connected to Pg cluster state as node")
		sm.Set(pg)
		return
	}
	log.Err(err).
		Msg("error connecting to PgRootDsn. Starting retry loop")
	ticker := time.NewTicker(5 * time.Second)

	for try := range defaultRetries {
		select {
		case <-ticker.C:
			pg, err = state.NewPgStateManager(ctx, workerNodeDsn)
			if err == nil {
				sm.Set(pg)
				return
			}

			delay := defaultRetryDelay * time.Duration(try)
			log.Debug().
				Err(err).
				Int("try", try).
				Str("next_delay", delay.String()).
				Msg("Failed to connect to pg cluster state as node")

			ticker.Reset(delay)
		}
	}

	log.Warn().
		Err(err).
		Msg("Can't connect to pg cluster as working node. Retry loop is stopped. " +
			"You can try connecting to cluster again via UI or API.")

}
