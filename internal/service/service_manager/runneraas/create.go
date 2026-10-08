package runneraas

import (
	"context"

	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/service/service_manager/runneraas/providers"
	"go.vervstack.ru/Velez/internal/user_errors"
)

// CreateRunner validates the request, mints a GitLab runner from the personal
// access token (so the token never reaches the persisted task payload), then
// enqueues the multi-step create_runner task (mint/store the GitHub
// registration token, deploy the runner container, wait for it, register the
// velez.runners row -
// internal/jobs/create_runner.go) and returns immediately - mirrors
// registryaas.CreateRegistryInstance: creating a runner means deploying a
// container through the ordinary (asynchronous) CreateNewDeploy/deploy
// watcher path, not something a single service-layer call can do by itself.
// Callers watch progress through TasksApi.WatchTask(req.Name,
// jobs.CreateRunnerAction) and refetch ListRunners once the task reaches
// DONE.
func (s *RunneraasService) CreateRunner(ctx context.Context, req domain.CreateRunnerReq) error {
	err := domain.ValidateInstanceName(req.Name)
	if err != nil {
		return rerrors.Wrap(err, "error validating runner name")
	}

	err = validateRunnerTarget(req.Scope, req.Target)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if req.Provider == velez_api.RunnerProvider_GITLAB {
		err = validateGitlabAccessToken(req.AccessToken)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	if req.IsBuildkitEnabled {
		err = validateBuildkitRunner(req.Provider, req.DindName != "")
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	err = validateDockerSource(req.DindName, req.DockerSocketAddress)
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = validateDockerSocketAddress(req.DockerSocketAddress)
	if err != nil {
		return rerrors.Wrap(err)
	}

	if req.DindName != "" {
		err = s.ensureDindExists(ctx, req.DindName)
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	isGitlab := req.Provider == velez_api.RunnerProvider_GITLAB
	instanceName := jobs.RunnerInstanceName(req.Provider, req.Name)

	var created providers.CreatedRunner

	payloadReq := req

	if isGitlab {
		created, err = s.mintGitlabRunner(ctx, req, instanceName)
		if err != nil {
			return rerrors.Wrap(err)
		}

		payloadReq.AccessToken = ""
	}

	initialContext := &velez_api.CreateRunnerTaskPayload{
		Request: runnerRequestToPb(payloadReq),
	}

	if isGitlab {
		initialContext.SetRegistrationToken(created.Token)
		initialContext.SetGitlabRunnerId(created.Id)
	}

	_, err = s.jobsEngine.EnqueueReplacing(ctx, req.Name, jobs.CreateRunnerAction, initialContext)
	if err != nil {
		if isGitlab {
			s.dropMintedGitlabRunner(ctx, req, instanceName, created.Token)
		}

		return rerrors.Wrap(err, "error enqueuing create runner task")
	}

	return nil
}

func (s *RunneraasService) mintGitlabRunner(
	ctx context.Context, req domain.CreateRunnerReq, instanceName string,
) (providers.CreatedRunner, error) {
	createReq := providers.CreateRunnerReq{
		BaseUrl:             req.BaseUrl,
		PersonalAccessToken: req.AccessToken,
		Scope:               req.Scope,
		Target:              req.Target,
		Description:         instanceName,
		Labels:              req.Labels,
	}

	created, err := s.gitlabSeeder.CreateRunner(ctx, createReq)
	if err != nil {
		return providers.CreatedRunner{}, rerrors.Wrap(err, "error creating gitlab runner")
	}

	err = s.secrets.Put(ctx, domain.RunnerRegistrationTokenSecretRef(instanceName), created.Token)
	if err != nil {
		s.dropMintedGitlabRunner(ctx, req, instanceName, created.Token)

		return providers.CreatedRunner{}, rerrors.Wrap(err, "error storing runner registration token")
	}

	return created, nil
}

func (s *RunneraasService) dropMintedGitlabRunner(
	ctx context.Context, req domain.CreateRunnerReq, instanceName, runnerToken string,
) {
	err := s.gitlabSeeder.DeleteRunnerByToken(ctx, req.BaseUrl, runnerToken)
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Str("runner", req.Name).Msg("error cleaning up minted gitlab runner")
	}

	err = s.secrets.Delete(ctx, domain.RunnerRegistrationTokenSecretRef(instanceName))
	if err != nil {
		log.Ctx(ctx).Warn().Err(err).Str("runner", req.Name).Msg("error cleaning up minted gitlab runner")
	}
}

func (s *RunneraasService) ensureDindExists(ctx context.Context, dindName string) error {
	svc, err := s.dataStorage.Services().GetByName(ctx, dindName)
	if err != nil {
		if rerrors.Is(err, user_errors.ErrStorageNotFound) {
			return rerrors.Wrap(user_errors.ErrDindNotFound)
		}

		return rerrors.Wrap(err, "error getting dind service")
	}

	_, err = s.dataStorage.DindInstances().GetDindInstanceByServiceId(ctx, svc.ID)
	if err != nil {
		if rerrors.Is(err, user_errors.ErrStorageNotFound) {
			return rerrors.Wrap(user_errors.ErrDindNotFound)
		}

		return rerrors.Wrap(err, "error getting dind instance")
	}

	return nil
}

// runnerRequestToPb converts the domain request into the wire request the
// create_runner task persists as its own context -
// internal/jobs/create_runner.go reads request fields straight off
// *velez_api.CreateRunner_Request, so the task payload carries the proto
// message rather than a second, parallel domain-shaped copy. The
// provider_config oneof is rebuilt from req.Provider/AccessToken/BaseUrl/Concurrent,
// which the transport layer already resolved out of the wire oneof once.
func runnerRequestToPb(req domain.CreateRunnerReq) *velez_api.CreateRunner_Request {
	pbReq := &velez_api.CreateRunner_Request{
		Name:   req.Name,
		Scope:  req.Scope,
		Target: req.Target,
		Labels: req.Labels,

		IsBuildkitEnabled: req.IsBuildkitEnabled,
	}

	if req.Environment != "" {
		pbReq.Environment = &req.Environment
	}

	if req.DockerSocketAddress != "" {
		pbReq.DockerSocketAddress = &req.DockerSocketAddress
	}

	if req.DindName != "" {
		pbReq.DindName = &req.DindName
	}

	switch req.Provider {
	case velez_api.RunnerProvider_GITLAB:
		pbReq.ProviderConfig = &velez_api.CreateRunner_Request_Gitlab{
			Gitlab: &velez_api.GitlabConfig{
				AccessToken: req.AccessToken,
				BaseUrl:     &req.BaseUrl,
				Concurrent:  &req.Concurrent,
			},
		}
	default:
		pbReq.ProviderConfig = &velez_api.CreateRunner_Request_Github{
			Github: &velez_api.GithubConfig{
				AccessToken: req.AccessToken,
			},
		}
	}

	return pbReq
}
