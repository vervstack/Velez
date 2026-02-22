package state

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres"
)

type pgState struct {
	storage.Storage
}

const PgName = "verv-cluster-state"

func NewPgStateManager(ctx context.Context, dsn string) (cluster_clients.ClusterStateManager, error) {
	conn, err := sqldb.New(dsn)
	if err != nil {
		return nil, rerrors.Wrap(err, "error connecting to database")
	}

	state := &pgState{
		Storage: postgres.New(conn),
	}

	go state.doHeartbeat(ctx)

	return state, nil
}

func (s *pgState) doHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5)

	for {
		select {
		case <-ticker.C:
			err := s.Nodes().UpdateOnline(ctx)
			if err != nil {
				log.Error().
					Err(err).
					Msg("error updating this node's last online on pg cluster")
			}
		case <-ctx.Done():
			return
		}

	}

}
