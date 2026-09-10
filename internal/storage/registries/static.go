package registries

import (
	"context"
	"database/sql"
	"sort"
	"sync"
	"time"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/user_errors"
)

// dockerHubRegistryName is the single builtin registry seeded by
// migrations/20260901120000_registries.sql - the in-memory storage seeds the
// same row so single-node/dev mode behaves like a freshly migrated cluster.
const (
	dockerHubRegistryName = "Docker Hub"
	dockerHubRegistryID   = int64(1)
)

// staticStorage is a read-only in-memory implementation of
// storage.RegistriesStorage, seeded with the same builtin registry
// migrations/20260901120000_registries.sql seeds in cluster mode. There is no
// way to add a custom registry in single-node/dev mode - see NewStatic.
//
// It is NOT dead code: local_storage (single-node / dev mode, no postgres)
// uses it as its Registries() backend, mirroring
// internal/storage/environments.staticStorage.
type staticStorage struct {
	m      *sync.RWMutex
	nextID int64
	byID   map[int64]domain.Registry
}

// NewStatic builds an in-memory registries storage seeded with the builtin
// Docker Hub registry. Every write method returns
// user_errors.ErrRequiresStatefullMode instead of mutating - single-node/dev
// mode offers this fixed set only, and points callers at statefull/postgres
// mode for real registry management.
//
// The one exception is UpsertBuiltinRegistry (builtin.go) - a narrow,
// non-interface escape hatch for the registry plugin's own system row, which
// still needs to allocate an ID for a row this seed didn't create.
func NewStatic() storage.RegistriesStorage {
	now := time.Now()

	dockerHub := domain.Registry{
		ID:        dockerHubRegistryID,
		Name:      dockerHubRegistryName,
		Type:      domain.RegistryTypeDockerHub,
		IsDefault: true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return &staticStorage{
		m:      &sync.RWMutex{},
		nextID: dockerHubRegistryID + 1,
		byID:   map[int64]domain.Registry{dockerHub.ID: dockerHub},
	}
}

func (s *staticStorage) ListRegistries(_ context.Context) ([]domain.Registry, error) {
	s.m.RLock()
	defer s.m.RUnlock()

	out := make([]domain.Registry, 0, len(s.byID))
	for _, reg := range s.byID {
		out = append(out, reg)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})

	return out, nil
}

func (s *staticStorage) GetRegistryByID(_ context.Context, id int64) (domain.Registry, error) {
	s.m.RLock()
	defer s.m.RUnlock()

	reg, ok := s.byID[id]
	if !ok {
		return domain.Registry{}, rerrors.Wrap(storage.ErrNotFound)
	}

	return reg, nil
}

func (s *staticStorage) CreateRegistry(_ context.Context, _ domain.CreateRegistryReq) (domain.Registry, error) {
	return domain.Registry{}, rerrors.Wrap(user_errors.ErrRequiresStatefullMode)
}

func (s *staticStorage) UpdateRegistry(_ context.Context, _ domain.UpdateRegistryReq) (domain.Registry, error) {
	return domain.Registry{}, rerrors.Wrap(user_errors.ErrRequiresStatefullMode)
}

func (s *staticStorage) DeleteRegistry(_ context.Context, _ int64) error {
	return rerrors.Wrap(user_errors.ErrRequiresStatefullMode)
}

func (s *staticStorage) ClearDefaultRegistry(_ context.Context) error {
	return rerrors.Wrap(user_errors.ErrRequiresStatefullMode)
}

// WithTx is a no-op for the in-memory backend - single-node/dev mode has no
// database transaction to thread through, so the same storage is returned
// (mirrors local_storage's other WithTx stubs, e.g. local_storage/jobs.go).
func (s *staticStorage) WithTx(_ *sql.Tx) storage.RegistriesStorage {
	return s
}
