package environments

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
)

// DefaultEnvironmentName is the name of the environment seeded by
// migrations/20260802120000_environments.sql, whose suffix matches the node's
// configured ContainerSuffix. The in-memory storage seeds the same row so
// single-node/dev mode behaves like a freshly migrated cluster.
const (
	DefaultEnvironmentName = "PROD"
)

// staticStorage is a full in-memory implementation of
// storage.EnvironmentsStorage.
//
// It is NOT dead code: local_storage (single-node / dev mode, no postgres)
// uses it as its Environments() backend, and it doubles as a test fixture for
// anything that needs a real, mutable environments storage without a database.
type staticStorage struct {
	m      sync.RWMutex
	nextID int64
	byID   map[int64]domain.Environment
}

// NewStatic builds an in-memory environments storage seeded with one
// environment per name in envs plus a DefaultEnvironmentName row carrying
// defaultSuffix. Each named environment's suffix defaults to its own name,
// matching CreateEnvironment's "empty suffix defaults to name" rule.
func NewStatic(envs []string, defaultSuffix string) storage.EnvironmentsStorage {
	s := &staticStorage{
		nextID: 1,
		byID:   make(map[int64]domain.Environment, len(envs)+1),
	}

	s.seed(DefaultEnvironmentName, defaultSuffix)

	for _, name := range envs {
		if name == "" || name == DefaultEnvironmentName {
			continue
		}

		s.seed(name, name)
	}

	return s
}

func (s *staticStorage) ListEnvironments(_ context.Context) ([]domain.Environment, error) {
	s.m.RLock()
	defer s.m.RUnlock()

	out := make([]domain.Environment, 0, len(s.byID))
	for _, env := range s.byID {
		out = append(out, env)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})

	return out, nil
}

func (s *staticStorage) GetEnvironmentByID(_ context.Context, id int64) (domain.Environment, error) {
	s.m.RLock()
	defer s.m.RUnlock()

	env, ok := s.byID[id]
	if !ok {
		return domain.Environment{}, rerrors.Wrap(storage.ErrNotFound)
	}

	return env, nil
}

func (s *staticStorage) GetEnvironmentByName(_ context.Context, name string) (domain.Environment, error) {
	s.m.RLock()
	defer s.m.RUnlock()

	for _, env := range s.byID {
		if env.Name == name {
			return env, nil
		}
	}

	return domain.Environment{}, rerrors.Wrap(storage.ErrNotFound)
}

func (s *staticStorage) CreateEnvironment(
	_ context.Context,
	req domain.CreateEnvironmentReq,
) (domain.Environment, error) {
	s.m.Lock()
	defer s.m.Unlock()

	for _, env := range s.byID {
		if env.Name == req.Name {
			return domain.Environment{}, errors.Join(storage.ErrAlreadyExists,
				rerrors.New("environment already exists: "+req.Name))
		}
	}

	now := time.Now()

	env := domain.Environment{
		ID:        s.nextID,
		Name:      req.Name,
		Suffix:    req.Suffix,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.byID[env.ID] = env
	s.nextID++

	return env, nil
}

func (s *staticStorage) UpdateEnvironment(
	_ context.Context,
	req domain.UpdateEnvironmentReq,
) (domain.Environment, error) {
	s.m.Lock()
	defer s.m.Unlock()

	env, ok := s.byID[req.ID]
	if !ok {
		return domain.Environment{}, rerrors.Wrap(storage.ErrNotFound)
	}

	if req.Name != nil {
		for id, other := range s.byID {
			if id != req.ID && other.Name == *req.Name {
				return domain.Environment{}, errors.Join(storage.ErrAlreadyExists,
					rerrors.New("environment already exists: "+*req.Name))
			}
		}

		env.Name = *req.Name
	}

	if req.Suffix != nil {
		env.Suffix = *req.Suffix
	}

	env.UpdatedAt = time.Now()
	s.byID[env.ID] = env

	return env, nil
}

func (s *staticStorage) DeleteEnvironment(_ context.Context, id int64) error {
	s.m.Lock()
	defer s.m.Unlock()

	_, ok := s.byID[id]
	if !ok {
		return rerrors.Wrap(storage.ErrNotFound)
	}

	delete(s.byID, id)

	return nil
}

func (s *staticStorage) seed(name, suffix string) {
	now := time.Now()

	env := domain.Environment{
		ID:        s.nextID,
		Name:      name,
		Suffix:    suffix,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.byID[env.ID] = env
	s.nextID++
}
