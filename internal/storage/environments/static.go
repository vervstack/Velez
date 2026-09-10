package environments

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

// DefaultEnvironmentName is the name of the environment seeded by
// migrations/20260802120000_environments.sql, whose suffix matches the node's
// configured ContainerSuffix. The in-memory storage seeds the same row so
// single-node/dev mode behaves like a freshly migrated cluster.
const (
	DefaultEnvironmentName = "PROD"
)

// staticStorage is a read-only in-memory implementation of
// storage.EnvironmentsStorage - every write method returns
// user_errors.ErrRequiresStatefullMode instead of mutating.
//
// It is NOT dead code: local_storage (single-node / dev mode, no postgres)
// uses it as its Environments() backend.
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
	_ domain.CreateEnvironmentReq,
) (domain.Environment, error) {
	return domain.Environment{}, rerrors.Wrap(user_errors.ErrRequiresStatefullMode)
}

func (s *staticStorage) UpdateEnvironment(
	_ context.Context,
	_ domain.UpdateEnvironmentReq,
) (domain.Environment, error) {
	return domain.Environment{}, rerrors.Wrap(user_errors.ErrRequiresStatefullMode)
}

func (s *staticStorage) DeleteEnvironment(_ context.Context, _ int64) error {
	return rerrors.Wrap(user_errors.ErrRequiresStatefullMode)
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
