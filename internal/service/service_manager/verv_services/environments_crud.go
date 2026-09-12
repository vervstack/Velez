package verv_services

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/environments"
	"go.vervstack.ru/Velez/internal/user_errors"
)

func (v *VervService) ListEnvironments(ctx context.Context) ([]domain.Environment, error) {
	envs, err := v.environments().ListEnvironments(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing environments from storage")
	}

	return envs, nil
}

func (v *VervService) GetEnvironment(ctx context.Context, name string) (domain.Environment, error) {
	if name == "" {
		return domain.Environment{}, rerrors.Wrap(user_errors.ErrEnvironmentRequired)
	}

	env, err := v.environments().GetEnvironmentByName(ctx, name)
	if err != nil {
		return domain.Environment{}, rerrors.Wrapf(user_errors.ErrEnvironmentNotFound, "unknown environment %q", name)
	}

	return env, nil
}

// ResolveEnvironmentSuffix maps an environment NAME (what callers pass on the
// wire) to the SUFFIX actually used for Docker naming/labels
// (labels.SuffixLabel). It doubles as the validation entry point for the six
// request messages that carry an `environment` field: an unknown name is
// rejected here.
//
// An EMPTY name is not an error - it silently resolves to the default
// environment (environments.DefaultEnvironmentName, the row seeded by
// migrations/20260802120000_environments.sql carrying this node's previously
// configured ContainerSuffix). Velez must keep working as a single-environment
// node for every caller that predates environments (existing UI code paths,
// e2e tests, internally built requests), so "no environment" means "the
// default one", not "bad request".
//
// GetEnvironment itself stays strict on purpose: it's the "look up THIS named
// environment" primitive, where an empty name really is a caller error.
func (v *VervService) ResolveEnvironmentSuffix(ctx context.Context, name string) (string, error) {
	if name == "" {
		name = environments.DefaultEnvironmentName
	}

	env, err := v.GetEnvironment(ctx, name)
	if err != nil {
		return "", err
	}

	return env.Suffix, nil
}

// CreateEnvironment persists a new environment. An omitted/empty suffix
// defaults to the environment's own name; an omitted/empty DockerHost
// defaults to this node's own configured Docker connection
// (container_runtime.resolver.Runtime treats that as "shared daemon", not
// "dedicated").
func (v *VervService) CreateEnvironment(
	ctx context.Context,
	req domain.CreateEnvironmentReq,
) (domain.Environment, error) {
	if req.Name == "" {
		return domain.Environment{}, rerrors.Wrap(user_errors.ErrEnvironmentRequired)
	}

	if req.Suffix == "" {
		req.Suffix = req.Name
	}

	if req.DockerHost == "" {
		req.DockerHost = v.docker.Host()
	}

	env, err := v.environments().CreateEnvironment(ctx, req)
	if err != nil {
		return domain.Environment{}, rerrors.Wrap(err, "error creating environment")
	}

	return env, nil
}

func (v *VervService) UpdateEnvironment(
	ctx context.Context,
	req domain.UpdateEnvironmentReq,
) (domain.Environment, error) {
	if req.ID == 0 {
		return domain.Environment{}, user_errors.ErrEnvironmentIdRequired
	}

	env, err := v.environments().UpdateEnvironment(ctx, req)
	if err != nil {
		return domain.Environment{}, rerrors.Wrap(err, "error updating environment")
	}

	return env, nil
}

// DeleteEnvironment removes every Docker resource tagged with the
// environment's suffix before dropping the DB row - see
// cascadeRemoveEnvironmentResources.
func (v *VervService) DeleteEnvironment(ctx context.Context, req domain.DeleteEnvironmentReq) error {
	env, err := v.resolveDeleteTarget(ctx, req)
	if err != nil {
		return err
	}

	err = v.cascadeRemoveEnvironmentResources(ctx, env.Suffix)
	if err != nil {
		return rerrors.Wrap(err, "error removing environment's docker resources")
	}

	err = v.environments().DeleteEnvironment(ctx, env.ID)
	if err != nil {
		return rerrors.Wrap(err, "error deleting environment")
	}

	return nil
}

func (v *VervService) resolveDeleteTarget(
	ctx context.Context,
	req domain.DeleteEnvironmentReq,
) (domain.Environment, error) {
	switch {
	case req.ID != nil && *req.ID != 0:
		env, err := v.environments().GetEnvironmentByID(ctx, *req.ID)
		if err != nil {
			return domain.Environment{}, rerrors.Wrap(user_errors.ErrEnvironmentNotFound, "unknown environment id")
		}

		return env, nil
	case req.Name != nil && *req.Name != "":
		return v.GetEnvironment(ctx, *req.Name)
	default:
		return domain.Environment{}, user_errors.ErrEnvironmentIdOrNameRequired
	}
}
