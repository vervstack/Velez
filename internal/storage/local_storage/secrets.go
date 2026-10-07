package local_storage

import (
	"context"
	"strconv"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
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

const (
	dockerSocketPath      = "/var/run/docker.sock"
	sysboxRuntimeName     = "sysbox-runc"
	runningContainerState = "running"
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
// secrets.NewStatic, except for the two refs below.
//
// The docker-socket grant (domain.DockerSocketGrantSecretRef) and the
// container isolation (domain.ContainerIsolationSecretRef) also get
// restart-recovery on Get only: without a postgres-backed store they would
// otherwise vanish with the process, and deployWatcher.upgrade would recreate a
// runner without its socket or a dind without sysbox/privileged. They are
// derived from the service's live container - the socket grant only from a
// real host-socket bind mount (never from labels or env, which a client can
// set), the isolation from HostConfig.Runtime/Privileged - and fall back to
// memory for the window before the container exists.
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
		return d.getNonPgaasSecret(ctx, ref)
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

func (d *dockerSecrets) getNonPgaasSecret(ctx context.Context, ref domain.SecretRef) (string, error) {
	if ref == domain.RunnerRegistrationTokenSecretRef(ref.Owner) {
		token := d.registrationTokenFromContainer(ctx, ref.Owner)
		if token != "" {
			return token, nil
		}
	}

	derived, ok := d.derivedFromServiceContainer(ctx, ref)
	if ok {
		return derived, nil
	}

	value, err := d.fallback.GetSecret(ctx, ref)
	if err != nil {
		return "", rerrors.Wrap(err, "error getting secret")
	}

	return value, nil
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

// derivedFromServiceContainer derives the docker-socket grant and the
// container isolation of ref.Key's live container. ok is false when ref is
// neither of those, there is no such container, docker fails, or the
// container shows nothing to derive - the caller then falls back to memory.
func (d *dockerSecrets) derivedFromServiceContainer(ctx context.Context, ref domain.SecretRef) (string, bool) {
	isSocketGrant := ref == domain.DockerSocketGrantSecretRef(ref.Key)
	isIsolation := ref == domain.ContainerIsolationSecretRef(ref.Key)

	if !isSocketGrant && !isIsolation {
		return "", false
	}

	info, ok := d.inspectServiceContainer(ctx, ref.Key)
	if !ok {
		return "", false
	}

	if isSocketGrant {
		return socketGrantFromContainer(info)
	}

	return isolationFromContainer(info)
}

// inspectServiceContainer inspects the container labelled
// labels.VervServiceLabel=name, preferring a running one - the docker name may
// carry an environment suffix, so the label is the identity, as in
// dockerServices.GetByName.
func (d *dockerSecrets) inspectServiceContainer(ctx context.Context, name string) (container.InspectResponse, bool) {
	if name == "" {
		return container.InspectResponse{}, false
	}

	listReq := &pb.ListSmerds_Request{
		Label: map[string]string{labels.VervServiceLabel: name},
	}

	containers, err := d.docker.ListContainers(ctx, listReq, allEnvironments)
	if err != nil || len(containers) == 0 {
		return container.InspectResponse{}, false
	}

	chosen := containers[0]

	for _, c := range containers {
		if c.State == runningContainerState {
			chosen = c

			break
		}
	}

	info, err := d.docker.Client().ContainerInspect(ctx, chosen.ID)
	if err != nil {
		return container.InspectResponse{}, false
	}

	return info, true
}

func socketGrantFromContainer(info container.InspectResponse) (string, bool) {
	for _, m := range info.Mounts {
		if m.Type == mount.TypeBind && m.Destination == dockerSocketPath {
			return strconv.FormatBool(true), true
		}
	}

	return "", false
}

func isolationFromContainer(info container.InspectResponse) (string, bool) {
	if info.HostConfig == nil {
		return "", false
	}

	if info.HostConfig.Runtime == sysboxRuntimeName {
		return domain.ContainerIsolationSysbox, true
	}

	if info.HostConfig.Privileged {
		return domain.ContainerIsolationPrivileged, true
	}

	return "", false
}
