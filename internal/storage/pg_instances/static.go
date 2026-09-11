package pg_instances

import (
	"context"
	"sort"
	"sync"
	"time"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/user_errors"
)

// staticStorage is a full in-memory implementation of
// storage.PgInstancesStorage.
//
// It is NOT dead code: local_storage (single-node / dev mode, no postgres)
// uses it as its PgInstances() backend, mirroring
// internal/storage/registries.staticStorage.
type staticStorage struct {
	m           *sync.RWMutex
	byServiceID map[int64]domain.PgInstance
}

// NewStatic builds an empty in-memory pg instances storage.
func NewStatic() storage.PgInstancesStorage {
	return &staticStorage{
		m:           &sync.RWMutex{},
		byServiceID: make(map[int64]domain.PgInstance),
	}
}

func (s *staticStorage) UpsertPgInstance(_ context.Context, req domain.UpsertPgInstanceReq) (domain.PgInstance, error) {
	s.m.Lock()
	defer s.m.Unlock()

	now := time.Now()

	existing, ok := s.byServiceID[req.ServiceID]

	createdAt := now
	if ok {
		createdAt = existing.CreatedAt
	}

	instance := domain.PgInstance{
		ServiceID: req.ServiceID,
		DbName:    req.DbName,
		Username:  req.Username,
		SecretRef: req.SecretRef,
		Port:      req.Port,
		CreatedAt: createdAt,
		UpdatedAt: now,
	}

	s.byServiceID[req.ServiceID] = instance

	return instance, nil
}

func (s *staticStorage) GetPgInstanceByServiceID(_ context.Context, serviceID int64) (domain.PgInstance, error) {
	s.m.RLock()
	defer s.m.RUnlock()

	instance, ok := s.byServiceID[serviceID]
	if !ok {
		return domain.PgInstance{}, rerrors.Wrap(user_errors.ErrStorageNotFound)
	}

	return instance, nil
}

func (s *staticStorage) ListPgInstances(_ context.Context) ([]domain.PgInstance, error) {
	s.m.RLock()
	defer s.m.RUnlock()

	out := make([]domain.PgInstance, 0, len(s.byServiceID))
	for _, instance := range s.byServiceID {
		out = append(out, instance)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].ServiceID < out[j].ServiceID
	})

	return out, nil
}

func (s *staticStorage) DeletePgInstance(_ context.Context, serviceID int64) error {
	s.m.Lock()
	defer s.m.Unlock()

	_, ok := s.byServiceID[serviceID]
	if !ok {
		return rerrors.Wrap(user_errors.ErrStorageNotFound)
	}

	delete(s.byServiceID, serviceID)

	return nil
}
