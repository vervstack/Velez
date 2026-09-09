package secrets

import (
	"context"
	"sort"
	"sync"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
)

// staticStorage is a full in-memory implementation of
// storage.SecretsStorage.
//
// It is NOT dead code: local_storage (single-node / dev mode, no postgres)
// uses it as its Secrets() backend, mirroring
// internal/storage/registries.staticStorage.
type staticStorage struct {
	m     *sync.RWMutex
	byRef map[domain.SecretRef]string
}

// NewStatic builds an empty in-memory secrets storage.
func NewStatic() storage.SecretsStorage {
	return &staticStorage{
		m:     &sync.RWMutex{},
		byRef: make(map[domain.SecretRef]string),
	}
}

func (s *staticStorage) PutSecret(_ context.Context, ref domain.SecretRef, value string) error {
	s.m.Lock()
	defer s.m.Unlock()

	s.byRef[ref] = value

	return nil
}

func (s *staticStorage) GetSecret(_ context.Context, ref domain.SecretRef) (string, error) {
	s.m.RLock()
	defer s.m.RUnlock()

	value, ok := s.byRef[ref]
	if !ok {
		return "", rerrors.Wrap(storage.ErrNotFound)
	}

	return value, nil
}

func (s *staticStorage) DeleteSecret(_ context.Context, ref domain.SecretRef) error {
	s.m.Lock()
	defer s.m.Unlock()

	_, ok := s.byRef[ref]
	if !ok {
		return rerrors.Wrap(storage.ErrNotFound)
	}

	delete(s.byRef, ref)

	return nil
}

func (s *staticStorage) ListSecretRefs(_ context.Context, scope, owner string) ([]domain.SecretRef, error) {
	s.m.RLock()
	defer s.m.RUnlock()

	out := make([]domain.SecretRef, 0)

	for ref := range s.byRef {
		if ref.Scope != scope || ref.Owner != owner {
			continue
		}

		out = append(out, ref)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Key < out[j].Key
	})

	return out, nil
}
