package velez_api_impl

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/user_errors"
)

func (impl *Impl) RegisterContainer(
	ctx context.Context,
	req *velez_api.RegisterContainer_Request,
) (*velez_api.RegisterContainer_Response, error) {
	if req.GetContainerId() == "" {
		return nil, rerrors.Wrap(user_errors.ErrRegisterContainerIdRequired)
	}

	if req.GetServiceName() == "" {
		return nil, rerrors.Wrap(user_errors.ErrRegisterServiceNameRequired)
	}

	if req.GetKeepPortMapping() && len(req.GetPorts()) != 0 {
		return nil, rerrors.Wrap(user_errors.ErrRegisterPortsWithKeepMapping)
	}

	_, err := impl.resolveEnvironment(ctx, req.GetEnvironment())
	if err != nil {
		return nil, err
	}

	err = impl.rejectUnregistrable(ctx, req.GetEnvironment(), req.GetContainerId())
	if err != nil {
		return nil, err
	}

	isPg := req.GetPg() != nil
	isRunner := req.GetRunner() != nil
	isRegistry := req.GetRegistry() != nil

	if isRunner {
		err = validateRunnerPattern(req.GetRunner())
		if err != nil {
			return nil, err
		}
	}

	if isRegistry {
		err = validateRegistryPattern(req.GetRegistry())
		if err != nil {
			return nil, err
		}
	}

	payload := &velez_api.RegisterContainerTaskPayload{
		ContainerId: req.GetContainerId(),
		Environment: req.GetEnvironment(),
		ServiceName: req.GetServiceName(),

		BindMountLinks:  req.GetBindMountLinks(),
		KeepPortMapping: req.GetKeepPortMapping(),
		Ports:           req.GetPorts(),
	}

	payload.ImageTag = req.ImageTag

	if isPg {
		payload.Pattern = velez_api.ServicePattern_SERVICE_PATTERN_POSTGRES
		payload.PgSuperuser = req.GetPg().GetSuperuser()
	}

	if isRunner {
		fillRunnerPayload(payload, req.GetRunner())
	}

	if isRegistry {
		payload.Pattern = velez_api.ServicePattern_SERVICE_PATTERN_REGISTRY
		payload.RegistryUsername = req.GetRegistry().GetUsername()
	}

	entityId := req.GetServiceName() + "/" + uuid.NewString()

	pgPassword := req.GetPg().GetPassword()
	if isPg && pgPassword != "" {
		ref := jobs.RegisteredPgPendingSecretRef(entityId)

		err = impl.secrets.Put(ctx, ref, pgPassword)
		if err != nil {
			return nil, rerrors.Wrap(err, "error storing pending pg password")
		}

		payload.PgPendingSecretOwner = entityId
	}

	if isRunner {
		err = impl.putPendingRunnerSecrets(ctx, entityId, req.GetRunner())
		if err != nil {
			return nil, err
		}

		payload.RunnerPendingSecretOwner = entityId
	}

	if isRegistry && req.GetRegistry().GetPassword() != "" {
		ref := jobs.RegisteredRegistryPendingSecretRef(entityId)

		err = impl.secrets.Put(ctx, ref, req.GetRegistry().GetPassword())
		if err != nil {
			return nil, rerrors.Wrap(err, "error storing pending registry password")
		}

		payload.RegistryPendingSecretOwner = entityId
	}

	_, err = impl.jobsEngine.Enqueue(ctx, entityId, jobs.RegisterContainerAction, payload)
	if err != nil {
		impl.dropPendingPgSecret(ctx, entityId)
		impl.dropPendingRunnerSecrets(ctx, entityId)
		impl.dropPendingRegistrySecret(ctx, entityId)

		return nil, rerrors.Wrap(err, "error enqueuing register_container task")
	}

	resp := &velez_api.RegisterContainer_Response{
		EntityId: entityId,
		Action:   jobs.RegisterContainerAction,
	}

	return resp, nil
}

func (impl *Impl) rejectUnregistrable(ctx context.Context, environment, containerId string) error {
	listReq := &velez_api.ListContainers_Request{Environment: environment}

	list, err := impl.smerdService.ListContainers(ctx, listReq)
	if err != nil {
		return rerrors.Wrap(err, "error listing containers")
	}

	for _, cont := range list.GetContainers() {
		isRequested := strings.HasPrefix(cont.GetId(), containerId)
		if !isRequested {
			continue
		}

		if domain.IsVelezImage(cont.GetImageName()) {
			return rerrors.Wrap(user_errors.ErrContainerIsVelez)
		}

		if cont.ReplacedByContainerId != nil {
			return rerrors.Wrap(user_errors.ErrContainerIsOnboardingLeftover)
		}

		if cont.NetworkOwnerContainerId != nil {
			return rerrors.Wrap(user_errors.ErrContainerSharesNetwork)
		}
	}

	return nil
}

func (impl *Impl) dropPendingPgSecret(ctx context.Context, owner string) {
	err := impl.secrets.Delete(ctx, jobs.RegisteredPgPendingSecretRef(owner))
	if err != nil && !rerrors.Is(err, user_errors.ErrSecretNotFound) {
		log.Ctx(ctx).Error().Err(err).Msg("error deleting pending pg secret after failed enqueue")
	}
}

func (impl *Impl) dropPendingRegistrySecret(ctx context.Context, owner string) {
	err := impl.secrets.Delete(ctx, jobs.RegisteredRegistryPendingSecretRef(owner))
	if err != nil && !rerrors.Is(err, user_errors.ErrSecretNotFound) {
		log.Ctx(ctx).Error().Err(err).Msg("error deleting pending registry secret after failed enqueue")
	}
}

func validateRegistryPattern(registry *velez_api.RegisterContainer_Request_RegistryPattern) error {
	hasUsername := registry.GetUsername() != ""
	hasPassword := registry.GetPassword() != ""

	if hasUsername != hasPassword {
		return rerrors.Wrap(user_errors.ErrRegistryCredentialsRequired)
	}

	return nil
}

func validateRunnerPattern(runner *velez_api.RegisterContainer_Request_RunnerPattern) error {
	if runner.GetTarget() == "" {
		return rerrors.Wrap(user_errors.ErrRunnerTargetRequired)
	}

	if runner.GetProvider() == velez_api.RunnerProvider_RUNNER_PROVIDER_UNSPECIFIED {
		return rerrors.Wrap(user_errors.ErrRunnerProviderUnsupported)
	}

	return nil
}

func fillRunnerPayload(
	payload *velez_api.RegisterContainerTaskPayload, runner *velez_api.RegisterContainer_Request_RunnerPattern,
) {
	if payload.GetPattern() == velez_api.ServicePattern_SERVICE_PATTERN_UNSPECIFIED {
		payload.Pattern = runnerServicePattern(runner.GetProvider())
	}

	payload.RunnerProvider = runner.GetProvider()
	payload.RunnerScope = runner.GetScope()
	payload.RunnerTarget = runner.GetTarget()
	payload.RunnerBaseUrl = runner.GetBaseUrl()
	payload.RunnerLabels = runner.GetLabels()
	payload.RunnerDockerImage = runner.GetDockerImage()
	payload.RunnerConcurrent = runner.GetConcurrent()
}

func runnerServicePattern(provider velez_api.RunnerProvider) velez_api.ServicePattern {
	if provider == velez_api.RunnerProvider_GITLAB {
		return velez_api.ServicePattern_SERVICE_PATTERN_GITLAB_RUNNER
	}

	return velez_api.ServicePattern_SERVICE_PATTERN_GITHUB_RUNNER
}

func (impl *Impl) putPendingRunnerSecrets(
	ctx context.Context, owner string, runner *velez_api.RegisterContainer_Request_RunnerPattern,
) error {
	pending := map[domain.SecretRef]string{
		jobs.RegisteredRunnerPendingAccessTokenRef(owner):       runner.GetAccessToken(),
		jobs.RegisteredRunnerPendingRegistrationTokenRef(owner): runner.GetRegistrationToken(),
	}

	for ref, value := range pending {
		if value == "" {
			continue
		}

		err := impl.secrets.Put(ctx, ref, value)
		if err != nil {
			impl.dropPendingRunnerSecrets(ctx, owner)

			return rerrors.Wrap(err, "error storing pending runner secret")
		}
	}

	return nil
}

func (impl *Impl) dropPendingRunnerSecrets(ctx context.Context, owner string) {
	refs := []domain.SecretRef{
		jobs.RegisteredRunnerPendingAccessTokenRef(owner),
		jobs.RegisteredRunnerPendingRegistrationTokenRef(owner),
	}

	for _, ref := range refs {
		err := impl.secrets.Delete(ctx, ref)
		if err != nil && !rerrors.Is(err, user_errors.ErrSecretNotFound) {
			log.Ctx(ctx).Error().Err(err).Msg("error deleting pending runner secret")
		}
	}
}
