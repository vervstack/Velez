package jobs

import (
	"context"
	"maps"
	"strings"

	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/parser"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/container_derived"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	stepStoreRunnerSecrets = "store_runner_secrets"
	stepUpsertRunnerRow    = "upsert_runner_row"

	runnerPendingSecretScope          = "runneraas-pending"
	runnerPendingAccessTokenKey       = "access_token"
	runnerPendingRegistrationTokenKey = "registration_token"
)

type registerRunnerAccessor interface {
	GetRunnerProvider() velez_api.RunnerProvider
	GetRunnerScope() velez_api.RunnerScope
	GetRunnerTarget() string
	GetRunnerBaseUrl() string
	GetRunnerLabels() []string
	GetRunnerDockerImage() string
	GetRunnerConcurrent() int32
	GetRunnerPendingSecretOwner() string
}

// RegisteredRunnerPendingAccessTokenRef holds the access token a register
// request supplied until store_runner_secrets moves it to its final ref; owner
// is the task entity id.
func RegisteredRunnerPendingAccessTokenRef(owner string) domain.SecretRef {
	return domain.SecretRef{Scope: runnerPendingSecretScope, Owner: owner, Key: runnerPendingAccessTokenKey}
}

// RegisteredRunnerPendingRegistrationTokenRef is the registration-token
// counterpart of RegisteredRunnerPendingAccessTokenRef.
func RegisteredRunnerPendingRegistrationTokenRef(owner string) domain.SecretRef {
	return domain.SecretRef{Scope: runnerPendingSecretScope, Owner: owner, Key: runnerPendingRegistrationTokenKey}
}

// registeredRunnerLabels are the labels the local_storage backend recovers a
// runner's facts from - see labels.RunnerInstanceLabel.
func registeredRunnerLabels(runner registerRunnerAccessor) map[string]string {
	return map[string]string{
		labels.RunnerInstanceLabel: labelTrueValue,
		labels.RunnerProviderLabel: runner.GetRunnerProvider().String(),
		labels.RunnerScopeLabel:    runner.GetRunnerScope().String(),
		labels.RunnerTargetLabel:   runner.GetRunnerTarget(),
		labels.RunnerLabelsLabel:   strings.Join(runner.GetRunnerLabels(), ","),
		labels.RunnerBaseUrlLabel:  runner.GetRunnerBaseUrl(),
	}
}

func readOptionalSecret(ctx context.Context, store secrets.Store, ref domain.SecretRef) (string, error) {
	value, err := store.Get(ctx, ref)
	if rerrors.Is(err, user_errors.ErrSecretNotFound) {
		return "", nil
	}

	if err != nil {
		return "", rerrors.Wrap(err, "error reading runner secret")
	}

	return value, nil
}

func deletePendingRunnerSecrets(ctx context.Context, store secrets.Store, owner string) error {
	if owner == "" {
		return nil
	}

	refs := []domain.SecretRef{
		RegisteredRunnerPendingAccessTokenRef(owner),
		RegisteredRunnerPendingRegistrationTokenRef(owner),
	}

	for _, ref := range refs {
		err := store.Delete(ctx, ref)
		if err != nil && !rerrors.Is(err, user_errors.ErrSecretNotFound) {
			return rerrors.Wrap(err, "error deleting pending runner secret")
		}
	}

	return nil
}

// loadRunnerRegistrationToken resolves the registration token: the request's
// pending one, then the stored one, then the container env.
func loadRunnerRegistrationToken(
	ctx context.Context,
	store secrets.Store,
	serviceName string,
	runner registerRunnerAccessor,
	containerEnv map[string]string,
) (string, error) {
	owner := runner.GetRunnerPendingSecretOwner()
	if owner != "" {
		pending, err := readOptionalSecret(ctx, store, RegisteredRunnerPendingRegistrationTokenRef(owner))
		if err != nil {
			return "", rerrors.Wrap(err)
		}

		if pending != "" {
			return pending, nil
		}
	}

	stored, err := readOptionalSecret(ctx, store, domain.RunnerRegistrationTokenSecretRef(serviceName))
	if err != nil {
		return "", rerrors.Wrap(err)
	}

	if stored != "" {
		return stored, nil
	}

	return container_derived.RunnerRegistrationToken(runner.GetRunnerProvider(), containerEnv), nil
}

// runnerStampEnv is the env a recreated container needs so Velez reads the
// registration token back from it: empty when the container already has it.
func runnerStampEnv(containerEnv map[string]string, token string) map[string]string {
	stamp := make(map[string]string)

	if token != "" && containerEnv[runnerRegistrationTokenEnvVar] != token {
		stamp[runnerRegistrationTokenEnvVar] = token
	}

	return stamp
}

// storeRunnerSecretsJob writes the final runner secrets - the registration
// token from the request or else the container env, and the optional access
// token - and drops the pending ones, also when it fails.
type storeRunnerSecretsJob struct {
	runtimes container_runtime.RuntimeResolver
	secrets  secrets.Store

	req    registerRequestAccessor
	runner registerRunnerAccessor
}

func (j *storeRunnerSecretsJob) Do(ctx context.Context) error {
	err := j.store(ctx)
	if err == nil {
		return nil
	}

	deleteErr := deletePendingRunnerSecrets(ctx, j.secrets, j.runner.GetRunnerPendingSecretOwner())
	if deleteErr != nil {
		log.Error().Err(deleteErr).Msg("error deleting pending runner secrets after failed store")
	}

	return rerrors.Wrap(err)
}

func (j *storeRunnerSecretsJob) store(ctx context.Context) error {
	info, err := inspectRegisteredContainer(ctx, j.runtimes, j.req)
	if err != nil {
		return rerrors.Wrap(err)
	}

	serviceName := j.req.GetServiceName()

	token, err := loadRunnerRegistrationToken(
		ctx, j.secrets, serviceName, j.runner, parser.ToDockerEnv(info.Config.Env))
	if err != nil {
		return rerrors.Wrap(err)
	}

	if token != "" {
		err = j.secrets.Put(ctx, domain.RunnerRegistrationTokenSecretRef(serviceName), token)
		if err != nil {
			return rerrors.Wrap(err, "error storing runner registration token")
		}
	}

	owner := j.runner.GetRunnerPendingSecretOwner()
	if owner != "" {
		err = j.storePendingAccessToken(ctx, owner, serviceName)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	err = deletePendingRunnerSecrets(ctx, j.secrets, owner)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

func (j *storeRunnerSecretsJob) storePendingAccessToken(ctx context.Context, owner, serviceName string) error {
	accessToken, err := readOptionalSecret(ctx, j.secrets, RegisteredRunnerPendingAccessTokenRef(owner))
	if err != nil {
		return rerrors.Wrap(err)
	}

	if accessToken == "" {
		return nil
	}

	err = j.secrets.Put(ctx, domain.RunnerAccessTokenSecretRef(serviceName), accessToken)
	if err != nil {
		return rerrors.Wrap(err, "error storing runner access token")
	}

	return nil
}

// upsertRunnerRowJob runs right after the bind transaction: Runners has no
// WithTx, and the upsert is idempotent.
type upsertRunnerRowJob struct {
	dataStorage storage.Storage

	req     registerRequestAccessor
	runner  registerRunnerAccessor
	service registeredServiceAccessor
}

func (j *upsertRunnerRowJob) Do(ctx context.Context) error {
	upsertReq := domain.UpsertRunnerReq{
		ServiceID:   j.service.GetServiceId(),
		Provider:    j.runner.GetRunnerProvider().String(),
		Scope:       j.runner.GetRunnerScope().String(),
		Target:      j.runner.GetRunnerTarget(),
		Labels:      j.runner.GetRunnerLabels(),
		SecretRef:   domain.RunnerAccessTokenSecretRef(j.req.GetServiceName()).String(),
		BaseUrl:     j.runner.GetRunnerBaseUrl(),
		DockerImage: j.runner.GetRunnerDockerImage(),
		Concurrent:  effectiveConcurrent(j.runner.GetRunnerConcurrent()),
	}

	_, err := j.dataStorage.Runners().UpsertRunner(ctx, upsertReq)
	if err != nil {
		return rerrors.Wrap(err, "error upserting runner row")
	}

	return nil
}

func (j *recreateWithLabelsJob) applyRunnerOverlay(
	ctx context.Context, payload *velez_api.UpgradeSmerdTaskPayload, containerEnv []string,
) error {
	env := parser.ToDockerEnv(containerEnv)

	token, err := loadRunnerRegistrationToken(ctx, j.secrets, j.req.GetServiceName(), j.runner, env)
	if err != nil {
		return rerrors.Wrap(err)
	}

	maps.Copy(payload.GetExtraLabels(), registeredRunnerLabels(j.runner))

	payload.ExtraEnv = runnerStampEnv(env, token)

	return nil
}
