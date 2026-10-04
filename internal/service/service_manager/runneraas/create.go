package runneraas

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/user_errors"
)

// CreateRunner validates the request, then enqueues the multi-step
// create_runner task (mint/store the registration token, deploy the runner
// container, wait for it, register the velez.runners row -
// internal/jobs/create_runner.go) and returns immediately - mirrors
// registryaas.CreateRegistryInstance: creating a runner means deploying a
// container through the ordinary (asynchronous) CreateNewDeploy/deploy
// watcher path, not something a single service-layer call can do by itself.
// Callers watch progress through TasksApi.WatchTask(req.Name,
// jobs.CreateRunnerAction) and refetch ListRunners once the task reaches
// DONE.
func (s *RunneraasService) CreateRunner(ctx context.Context, req domain.CreateRunnerReq) error {
	err := validateRunnerTarget(req.Scope, req.Target)
	if err != nil {
		return rerrors.Wrap(err)
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

	initialContext := &velez_api.CreateRunnerTaskPayload{
		Request: runnerRequestToPb(req),
	}

	_, err = s.jobsEngine.Enqueue(ctx, req.Name, jobs.CreateRunnerAction, initialContext)
	if err != nil {
		return rerrors.Wrap(err, "error enqueuing create runner task")
	}

	return nil
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
