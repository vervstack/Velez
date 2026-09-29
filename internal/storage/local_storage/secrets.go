package local_storage

import (
	"context"
	"sync"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/container_derived"
	"go.vervstack.ru/Velez/internal/storage/secrets"
	"go.vervstack.ru/Velez/internal/user_errors"
)

// dockerSecrets is the single-node/dev storage.SecretsStorage. Secrets in the
// pgaas scope are read straight back from the instance's running
// container env (pgaas writes POSTGRES_PASSWORD there at deploy time), so they
// survive a Velez restart the same way the container does; Put/Delete for
// that scope are no-ops with a small overlay that only bridges the window
// before the container exists. The runner registration-token ref
// (domain.RunnerRegistrationTokenSecretRef) gets the same
// restart-recovery treatment on Get only - see registrationTokenFromContainer
// - because create_runner.go writes it into the runner container's own env
// (container_derived.RunnerRegistrationTokenEnvVar) regardless of provider. Put/Delete for it,
// and every other scope/key, keep the plain in-memory behaviour of
// secrets.NewStatic.
type dockerSecrets struct {
	docker   node_clients.Docker
	fallback storage.SecretsStorage

	mu             sync.Mutex
	pendingPgaasPw map[domain.SecretRef]string
}

func newSecretsStorage(docker node_clients.Docker) *dockerSecrets {
	return &dockerSecrets{
		docker:         docker,
		fallback:       secrets.NewStatic(),
		pendingPgaasPw: make(map[domain.SecretRef]string),
	}
}

func (d *dockerSecrets) PutSecret(ctx context.Context, ref domain.SecretRef, value string) error {
	if ref.Scope != container_derived.PgaasSecretScope {
		err := d.fallback.PutSecret(ctx, ref, value)
		if err != nil {
			return rerrors.Wrap(err, "error putting secret")
		}

		return nil
	}

	d.mu.Lock()

	d.pendingPgaasPw[ref] = value

	d.mu.Unlock()

	return nil
}

func (d *dockerSecrets) GetSecret(ctx context.Context, ref domain.SecretRef) (string, error) {
	if ref.Scope != container_derived.PgaasSecretScope {
		if ref == domain.RunnerRegistrationTokenSecretRef(ref.Owner) {
			token := d.registrationTokenFromContainer(ctx, ref.Owner)
			if token != "" {
				return token, nil
			}
		}

		value, err := d.fallback.GetSecret(ctx, ref)
		if err != nil {
			return "", rerrors.Wrap(err, "error getting secret")
		}

		return value, nil
	}

	password := d.passwordFromContainer(ctx, ref.Owner)
	if password != "" {
		return password, nil
	}

	d.mu.Lock()

	pending, ok := d.pendingPgaasPw[ref]

	d.mu.Unlock()

	if ok {
		return pending, nil
	}

	return "", rerrors.Wrap(user_errors.ErrStorageNotFound)
}

func (d *dockerSecrets) DeleteSecret(ctx context.Context, ref domain.SecretRef) error {
	if ref.Scope != container_derived.PgaasSecretScope {
		err := d.fallback.DeleteSecret(ctx, ref)
		if err != nil {
			return rerrors.Wrap(err, "error deleting secret")
		}

		return nil
	}

	d.mu.Lock()
	delete(d.pendingPgaasPw, ref)
	d.mu.Unlock()

	return nil
}

func (d *dockerSecrets) ListSecretRefs(ctx context.Context, scope, owner string) ([]domain.SecretRef, error) {
	if scope != container_derived.PgaasSecretScope {
		refs, err := d.fallback.ListSecretRefs(ctx, scope, owner)
		if err != nil {
			return nil, rerrors.Wrap(err, "error listing secret refs")
		}

		return refs, nil
	}

	if d.passwordFromContainer(ctx, owner) == "" {
		return []domain.SecretRef{}, nil
	}

	return []domain.SecretRef{container_derived.PgInstanceSecretRef(owner)}, nil
}

// passwordFromContainer returns the POSTGRES_PASSWORD of the pgaas instance's
// container, or "" if there is no such container yet or it carries no
// password env. The container name is the instance name (pgaas launches it
// that way), and it must carry labels.PgaasInstanceLabel to be trusted here.
func (d *dockerSecrets) passwordFromContainer(ctx context.Context, name string) string {
	if name == "" {
		return ""
	}

	listReq := &pb.ListSmerds_Request{
		Name:  &name,
		Label: map[string]string{labels.PgaasInstanceLabel: boolLabelValue},
	}

	containers, err := d.docker.ListContainers(ctx, listReq, allEnvironments)
	if err != nil || len(containers) == 0 {
		return ""
	}

	info, err := d.docker.Client().ContainerInspect(ctx, containers[0].ID)
	if err != nil {
		return ""
	}

	return container_derived.EnvValue(info.Config.Env, container_derived.PgaasEnvPassword)
}

// registrationTokenFromContainer returns the runner container's
// container_derived.RunnerRegistrationTokenEnvVar value, or "" if there is no such container
// yet or it carries no such env. Mirrors passwordFromContainer's pattern for
// pgaas - see container_derived.RunnerRegistrationTokenEnvVar for
// why the token is written there regardless of provider.
func (d *dockerSecrets) registrationTokenFromContainer(ctx context.Context, name string) string {
	if name == "" {
		return ""
	}

	listReq := &pb.ListSmerds_Request{
		Name:  &name,
		Label: map[string]string{labels.RunnerInstanceLabel: boolLabelValue},
	}

	containers, err := d.docker.ListContainers(ctx, listReq, allEnvironments)
	if err != nil || len(containers) == 0 {
		return ""
	}

	info, err := d.docker.Client().ContainerInspect(ctx, containers[0].ID)
	if err != nil {
		return ""
	}

	return container_derived.EnvValue(info.Config.Env, container_derived.RunnerRegistrationTokenEnvVar)
}
