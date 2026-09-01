package registries

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"sync"
	"time"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
)

// staticStorage is a full in-memory implementation of
// storage.RegistriesStorage.
//
// It is NOT dead code: local_storage (single-node / dev mode, no postgres)
// uses it as its Registries() backend, mirroring
// internal/storage/environments.staticStorage.
type staticStorage struct {
	m      *sync.RWMutex
	nextID *int64
	byID   map[int64]domain.Registry
}

// NewStatic builds an empty in-memory registries storage. Unlike
// environments there is no config-seeded default to carry over.
func NewStatic() storage.RegistriesStorage {
	var nextID int64 = 1

	return &staticStorage{
		m:      &sync.RWMutex{},
		nextID: &nextID,
		byID:   make(map[int64]domain.Registry),
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

func (s *staticStorage) CreateRegistry(_ context.Context, req domain.CreateRegistryReq) (domain.Registry, error) {
	s.m.Lock()
	defer s.m.Unlock()

	for _, reg := range s.byID {
		if reg.Name == req.Name {
			return domain.Registry{}, errors.Join(storage.ErrAlreadyExists,
				rerrors.New("registry already exists: "+req.Name))
		}
	}

	now := time.Now()

	reg := domain.Registry{
		ID:        *s.nextID,
		Name:      req.Name,
		Type:      req.Type,
		Url:       req.Url,
		Username:  req.Username,
		Secret:    req.Secret,
		IsDefault: req.IsDefault,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.byID[reg.ID] = reg
	*s.nextID++

	return reg, nil
}

func (s *staticStorage) UpdateRegistry(_ context.Context, req domain.UpdateRegistryReq) (domain.Registry, error) {
	s.m.Lock()
	defer s.m.Unlock()

	reg, ok := s.byID[req.ID]
	if !ok {
		return domain.Registry{}, rerrors.Wrap(storage.ErrNotFound)
	}

	if req.Name != nil {
		for id, other := range s.byID {
			if id != req.ID && other.Name == *req.Name {
				return domain.Registry{}, errors.Join(storage.ErrAlreadyExists,
					rerrors.New("registry already exists: "+*req.Name))
			}
		}

		reg.Name = *req.Name
	}

	if req.Type != nil {
		reg.Type = *req.Type
	}

	if req.Url != nil {
		reg.Url = *req.Url
	}

	if req.Username != nil {
		reg.Username = *req.Username
	}

	if req.Secret != nil {
		reg.Secret = *req.Secret
	}

	if req.IsDefault != nil {
		reg.IsDefault = *req.IsDefault
	}

	reg.UpdatedAt = time.Now()
	s.byID[reg.ID] = reg

	return reg, nil
}

func (s *staticStorage) DeleteRegistry(_ context.Context, id int64) error {
	s.m.Lock()
	defer s.m.Unlock()

	_, ok := s.byID[id]
	if !ok {
		return rerrors.Wrap(storage.ErrNotFound)
	}

	delete(s.byID, id)

	return nil
}

func (s *staticStorage) ClearDefaultRegistry(_ context.Context) error {
	s.m.Lock()
	defer s.m.Unlock()

	for id, reg := range s.byID {
		if reg.IsDefault {
			reg.IsDefault = false
			s.byID[id] = reg
		}
	}

	return nil
}

// WithTx is a no-op for the in-memory backend - single-node/dev mode has no
// database transaction to thread through, so the same storage is returned
// (mirrors local_storage's other WithTx stubs, e.g. local_storage/jobs.go).
func (s *staticStorage) WithTx(_ *sql.Tx) storage.RegistriesStorage {
	return s
}
