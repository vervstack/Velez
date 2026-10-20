package postgres

import (
	"context"
	"sync"
	"time"

	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/storage/wakeup"
)

const (
	tasksChannel = "velez_tasks"

	listenerMinReconnectInterval = time.Second
	listenerMaxReconnectInterval = 30 * time.Second
)

type tasksStorage struct {
	*tasks_queries.Queries

	db  sqldb.DB
	dsn string

	created      *wakeup.Broadcaster
	listenerOnce sync.Once
}

func newTasksStorage(db sqldb.DB, dsn string) *tasksStorage {
	return &tasksStorage{
		Queries: tasks_queries.New(db),
		db:      db,
		dsn:     dsn,
		created: wakeup.New(),
	}
}

func (s *tasksStorage) NotifyTaskCreated(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "SELECT pg_notify($1, '')", tasksChannel)
	if err != nil {
		return rerrors.Wrap(err, "error sending task notification")
	}

	return nil
}

func (s *tasksStorage) SubscribeTaskCreated(ctx context.Context) <-chan struct{} {
	s.listenerOnce.Do(func() {
		go s.listen()
	})

	return s.created.Subscribe(ctx)
}

// listen lives as long as the process: it is shared by every subscriber, so no
// single subscriber's ctx can own it.
func (s *tasksStorage) listen() {
	listener := pq.NewListener(s.dsn, listenerMinReconnectInterval, listenerMaxReconnectInterval, logListenerEvent)

	err := listener.Listen(tasksChannel)
	if err != nil {
		log.Error().Err(err).Msg("error listening for task notifications")
	}

	// A nil notification means the connection was re-established; NOTIFYs sent
	// while it was down are lost, so it wakes subscribers like any other.
	for range listener.Notify {
		s.created.Notify()
	}
}

func logListenerEvent(event pq.ListenerEventType, err error) {
	if err != nil {
		log.Warn().Err(err).Int("event", int(event)).Msg("task notification listener event")
	}
}
