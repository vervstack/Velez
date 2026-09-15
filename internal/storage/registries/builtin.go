package registries

import (
	"context"
	"time"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

// BuiltinRegistryUpserter is a narrow, non-interface escape hatch for a
// system row a job maintains on the node's own behalf - currently just
// internal/jobs.registerRegistryRowJob's velez.registries row for the
// self-hosted registry plugin (docs/features/pgaas_and_registry_plugin.md
// section 4). It is deliberately NOT part of storage.RegistriesStorage: an
// ordinary caller still goes through CreateRegistry/UpdateRegistry, which
// single-node mode's staticStorage always rejects with
// user_errors.ErrRequiresStatefullMode.
//
// Both concrete registries storages satisfy it. pgStorage's version is just
// its existing public Create/Update path, upserted by name - postgres never
// rejects a write. staticStorage's version bypasses that rejection, but only
// for this one call path.
type BuiltinRegistryUpserter interface {
	UpsertBuiltinRegistry(ctx context.Context, req domain.CreateRegistryReq) (domain.Registry, error)
}

var (
	_ BuiltinRegistryUpserter = (*staticStorage)(nil)
	_ BuiltinRegistryUpserter = (*pgStorage)(nil)
)

// findRegistryByName is shared by both UpsertBuiltinRegistry implementations.
func findRegistryByName(all []domain.Registry, name string) *domain.Registry {
	for i := range all {
		if all[i].Name == name {
			return &all[i]
		}
	}

	return nil
}

func (s *staticStorage) UpsertBuiltinRegistry(
	_ context.Context,
	req domain.CreateRegistryReq,
) (domain.Registry, error) {
	s.m.Lock()
	defer s.m.Unlock()

	for id, existing := range s.byID {
		if existing.Name != req.Name {
			continue
		}

		existing.Type = req.Type
		existing.Url = req.Url
		existing.Username = req.Username
		existing.Secret = req.Secret
		existing.UpdatedAt = time.Now()

		s.byID[id] = existing

		return existing, nil
	}

	now := time.Now()

	reg := domain.Registry{
		Id:        s.nextID,
		Name:      req.Name,
		Type:      req.Type,
		Url:       req.Url,
		Username:  req.Username,
		Secret:    req.Secret,
		IsDefault: req.IsDefault,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.byID[reg.Id] = reg
	s.nextID++

	return reg, nil
}

func (p *pgStorage) UpsertBuiltinRegistry(
	ctx context.Context,
	req domain.CreateRegistryReq,
) (domain.Registry, error) {
	all, err := p.ListRegistries(ctx)
	if err != nil {
		return domain.Registry{}, rerrors.Wrap(err, "error listing registries")
	}

	existing := findRegistryByName(all, req.Name)
	if existing == nil {
		created, createErr := p.CreateRegistry(ctx, req)
		if createErr != nil {
			return domain.Registry{}, rerrors.Wrap(createErr, "error creating registry")
		}

		return created, nil
	}

	updateReq := domain.UpdateRegistryReq{
		Id:       existing.Id,
		Type:     &req.Type,
		Url:      &req.Url,
		Username: &req.Username,
		Secret:   &req.Secret,
	}

	updated, err := p.UpdateRegistry(ctx, updateReq)
	if err != nil {
		return domain.Registry{}, rerrors.Wrap(err, "error updating registry")
	}

	return updated, nil
}
