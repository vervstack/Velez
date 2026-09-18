package jobs

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/service/service_manager/runneraas/providers"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon/builtin"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	CreateRunnerAction = "create_runner"

	stepMintRunnerToken     = "mint_runner_token"
	stepDeployRunner        = "deploy_runner"
	stepWaitForRunnerDeploy = "wait_for_runner_deploy"
	stepRegisterRunnerRow   = "register_runner_row"

	// runnerDeployWaitTimeout mirrors registryDeployWaitTimeout - bounds how
	// long create_runner's task blocks inside waitForRunnerDeployJob, waiting
	// for the deploy watcher (internal/workers/deploy_watcher.go, 5s ticker)
	// to run and finish the create_smerd task it schedules for the runner
	// container.
	runnerDeployWaitTimeout = 600 * time.Second

	// runnerDockerHostEnvVar overlays DOCKER_HOST when the request names a
	// custom docker_socket_address, instead of granting the host bind-mount.
	runnerDockerHostEnvVar = "DOCKER_HOST"

	// runnerNameSuffixLen is the number of hex characters of a fresh
	// uuid.NewString() used to disambiguate the runner name a fresh
	// registration derives - a repeated CreateRunner call for the same
	// target must not collide with a still-running prior runner of the same
	// name inside the provider's own runner registry.
	runnerNameSuffixLen = 8
)

// createRunnerRequestAccessor is the narrow TaskContext slice every
// create_runner job needs to read the original request.
// *velez_api.CreateRunnerTaskPayload satisfies it.
type createRunnerRequestAccessor interface {
	GetRequest() *velez_api.CreateRunner_Request
}

// runnerRegistrationTokenAccessor lets mintRunnerTokenJob persist the minted/
// pass-through registration token so a crash-resumed task reuses it instead
// of re-minting one that would mismatch the token already written into the
// runner container's env and secret store.
type runnerRegistrationTokenAccessor interface {
	GetRegistrationToken() string
	SetRegistrationToken(token string)
}

type createRunnerHandler struct {
	dataStorage  storage.Storage
	secretsStore secrets.Store
	vervServices service.VervServicesService
	// jobsEngine lets waitForRunnerDeployJob watch the create_smerd task the
	// deploy watcher runs for this runner's container.
	jobsEngine Engine
}

func NewCreateRunnerHandler(
	dataStorage storage.Storage,
	secretsStore secrets.Store,
	vervServices service.VervServicesService,
	jobsEngine Engine,
) TaskHandler {
	return &createRunnerHandler{
		dataStorage:  dataStorage,
		secretsStore: secretsStore,
		vervServices: vervServices,
		jobsEngine:   jobsEngine,
	}
}

func (h *createRunnerHandler) Action() string {
	return CreateRunnerAction
}

func (h *createRunnerHandler) NewContext() TaskContext {
	return &velez_api.CreateRunnerTaskPayload{}
}

// BuildJobs mirrors create_registry_instance.go's shape, simplified: a
// runner has no ports/htpasswd/UI sidecar to resolve, so mint_runner_token/
// deploy_runner/wait_for_runner_deploy/register_runner_row is the full
// chain. deploy_runner and register_runner_row split apart (never done
// inline in one job, unlike the retired synchronous runneraas.CreateRunner)
// because register_runner_row must not run until wait_for_runner_deploy
// confirms the container actually exists - see that job's doc comment.
func (h *createRunnerHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.CreateRunnerTaskPayload)
	if !ok {
		panic("create_runner: BuildJobs called with mismatched TaskContext type")
	}

	provider, _, _ := runnerProviderConfig(payload.GetRequest())
	instanceName := runnerNamePrefix(provider) + payload.GetRequest().GetName()

	return []NamedJob{
		{
			Name: stepMintRunnerToken,
			Job: &mintRunnerTokenJob{
				secrets:      h.secretsStore,
				req:          payload,
				ctx:          payload,
				instanceName: instanceName,
			},
		},
		{
			Name: stepDeployRunner,
			Job: &deployRunnerJob{
				boxes:        h.dataStorage.ResourceBoxes(),
				vervServices: h.vervServices,
				secrets:      h.secretsStore,
				req:          payload,
				ctx:          payload,
				instanceName: instanceName,
			},
		},
		{
			Name: stepWaitForRunnerDeploy,
			Job: &waitForRunnerDeployJob{
				jobsEngine:   h.jobsEngine,
				req:          payload,
				instanceName: instanceName,
			},
		},
		{
			Name: stepRegisterRunnerRow,
			Job: &registerRunnerRowJob{
				services:     h.dataStorage.Services(),
				runners:      h.dataStorage.Runners(),
				instanceName: instanceName,
				req:          payload,
			},
		},
	}
}

// runnerProviderConfig resolves the RunnerProvider implied by which
// provider_config oneof variant is set (CreateRunner.Request carries no
// separate provider field), plus that variant's access token and base url.
func runnerProviderConfig(request *velez_api.CreateRunner_Request) (
	provider velez_api.RunnerProvider, accessToken, baseUrl string,
) {
	switch cfg := request.GetProviderConfig().(type) {
	case *velez_api.CreateRunner_Request_Github:
		return velez_api.RunnerProvider_GITHUB, cfg.Github.GetAccessToken(), ""
	case *velez_api.CreateRunner_Request_Gitlab:
		return velez_api.RunnerProvider_GITLAB, cfg.Gitlab.GetAccessToken(), cfg.Gitlab.GetBaseUrl()
	default:
		return velez_api.RunnerProvider_RUNNER_PROVIDER_UNSPECIFIED, "", ""
	}
}

// runnerNamePrefix picks the container-name prefix for the runner's
// provider, mirroring providers.For's dispatch. Falls back to no prefix for
// an unspecified provider - runnerProviderConfig already surfaces that as a
// mint-time error before deploy_runner ever reads instanceName.
func runnerNamePrefix(provider velez_api.RunnerProvider) string {
	switch provider {
	case velez_api.RunnerProvider_GITHUB:
		return labels.GithubRunnerNamePrefix
	case velez_api.RunnerProvider_GITLAB:
		return labels.GitlabRunnerNamePrefix
	default:
		return ""
	}
}

// mintRunnerTokenJob stores the caller's access token, then mints (GitHub)
// or passes through (GitLab) the runner registration token deploy_runner
// needs. Skips minting on a resumed task that already has one - see
// runnerRegistrationTokenAccessor's doc comment.
type mintRunnerTokenJob struct {
	secrets secrets.Store

	req          createRunnerRequestAccessor
	ctx          runnerRegistrationTokenAccessor
	instanceName string
}

func (j *mintRunnerTokenJob) Do(ctx context.Context) error {
	request := j.req.GetRequest()
	name := j.instanceName

	provider, accessToken, baseUrl := runnerProviderConfig(request)

	runnerProvider, err := providers.For(provider)
	if err != nil {
		return rerrors.Wrap(err, "error resolving runner provider")
	}

	err = j.secrets.Put(ctx, domain.RunnerAccessTokenSecretRef(name), accessToken)
	if err != nil {
		return rerrors.Wrap(err, "error storing runner access token")
	}

	if j.ctx.GetRegistrationToken() == "" {
		token, mintErr := runnerProvider.MintRegistrationToken(
			ctx, request.GetScope(), request.GetTarget(), baseUrl, accessToken,
		)
		if mintErr != nil {
			return rerrors.Wrap(mintErr, "error minting runner registration token")
		}

		j.ctx.SetRegistrationToken(token)
	}

	err = j.secrets.Put(ctx, domain.RunnerRegistrationTokenSecretRef(name), j.ctx.GetRegistrationToken())
	if err != nil {
		return rerrors.Wrap(err, "error storing runner registration token")
	}

	return nil
}

// deployRunnerJob resolves the builtin descriptor named by the request's
// provider, overlays the instance's shape (unique volume name, registration
// env, docker-socket grant), and hands off to the existing CreateNewDeploy
// path - the deploy watcher creates the actual container, this job never
// does. Mirrors deployRegistryInstanceJob.
type deployRunnerJob struct {
	boxes        vervonomicon.BoxLookup
	vervServices service.VervServicesService
	secrets      secrets.Store

	req          createRunnerRequestAccessor
	ctx          runnerRegistrationTokenAccessor
	instanceName string
}

func (j *deployRunnerJob) Do(ctx context.Context) error {
	request := j.req.GetRequest()
	name := j.instanceName

	provider, _, _ := runnerProviderConfig(request)

	runnerProvider, err := providers.For(provider)
	if err != nil {
		return rerrors.Wrap(err, "error resolving runner provider")
	}

	descriptor, smerdRequest, err := buildRunnerDeployRequest(
		ctx, j.boxes, runnerProvider, request, j.ctx.GetRegistrationToken(), name,
	)
	if err != nil {
		return rerrors.Wrap(err, "error building runner deploy request")
	}

	// A caller-supplied docker_socket_address means the runner talks to its
	// own daemon over the network - set as DOCKER_HOST, the same mechanism
	// buildRunnerDeployRequest already uses for registration env vars, and
	// it never gets the host bind-mount grant too. Otherwise, the
	// docker-socket grant MUST be written before CreateNewDeploy schedules
	// the deploy - deploy_watcher.go reads it, by the same
	// DockerSocketGrantSecretRef derivation, right before it enqueues the
	// create_smerd task that actually creates the container. See
	// create_smerd.go's dockerSocketAccessor gate. Never derived from, or
	// settable via, any other field on CreateRunner.Request.
	if request.GetDockerSocketAddress() != "" {
		smerdRequest.Env[runnerDockerHostEnvVar] = request.GetDockerSocketAddress()
	} else {
		err = j.secrets.Put(ctx, domain.DockerSocketGrantSecretRef(name), "true")
		if err != nil {
			return rerrors.Wrap(err, "error putting docker socket grant secret")
		}
	}

	deployReq := domain.CreateDeployReq{
		ServiceName:    name,
		VervDescriptor: &descriptor,
		LaunchSmerd:    domain.LaunchSmerd{CreateSmerd_Request: smerdRequest},
	}

	err = j.vervServices.CreateNewDeploy(ctx, deployReq)
	if err != nil {
		return rerrors.Wrap(err, "error creating runner deploy")
	}

	return nil
}

// buildRunnerDeployRequest reads the builtin descriptor named by provider,
// overlays the instance's shape (unique volume name, registration env),
// resolves it into a CreateSmerd.Request via the box resolver, then overlays
// the instance's generated name onto the *resolved request* - never onto the
// descriptor. Mirrors buildRegistryDeployRequest / the retired
// runneraas.buildDeployRequest's "no descriptor file ever contains a
// credential" discipline.
func buildRunnerDeployRequest(
	ctx context.Context,
	boxes vervonomicon.BoxLookup,
	provider providers.Provider,
	request *velez_api.CreateRunner_Request,
	registrationToken string,
	name string,
) (verv.Descriptor, *velez_api.CreateSmerd_Request, error) {
	files, err := builtin.Read(provider.DescriptorName())
	if err != nil {
		return verv.Descriptor{}, nil, rerrors.Wrap(err, "error reading builtin runner descriptor")
	}

	descriptor, err := vervonomicon.MergeEnvironment(files, request.GetEnvironment())
	if err != nil {
		return verv.Descriptor{}, nil, rerrors.Wrap(err, "error merging builtin runner descriptor environment")
	}

	descriptor.Source = verv.SourceKindBuiltin

	for i := range descriptor.Deployment.App.Volumes {
		if descriptor.Deployment.App.Volumes[i].Path != provider.DataPath() {
			continue
		}

		descriptor.Deployment.App.Volumes[i].Name = runnerVolumeName(name)
	}

	resolver := vervonomicon.NewBoxResolver(boxes)

	smerdRequest, err := resolver.ResolveRequest(ctx, descriptor, request.GetEnvironment(), "")
	if err != nil {
		return verv.Descriptor{}, nil, rerrors.Wrap(err, "error resolving runner deploy request")
	}

	smerdRequest.Name = name

	provEnum, _, baseUrl := runnerProviderConfig(request)
	runnerName := deriveRunnerName(request.GetTarget())

	smerdRequest.Env = provider.RegistrationEnv(
		request.GetScope(), request.GetTarget(), baseUrl, runnerName, registrationToken, request.GetLabels())

	if smerdRequest.Labels == nil {
		smerdRequest.Labels = make(map[string]string)
	}

	// VervServiceLabel makes the instance a first-class entry in the node's
	// service list (listDistinctServices keys on it); the Runner*Label set
	// lets the single-node local_storage backend recover the instance's
	// facts from the running container directly, without parsing any
	// provider-specific env var - see
	// internal/storage/local_storage/runners.go and
	// internal/domain/labels.RunnerInstanceLabel's doc comment. All inert in
	// cluster mode, where velez.runners is authoritative.
	smerdRequest.Labels[labels.VervServiceLabel] = name
	smerdRequest.Labels[labels.RunnerInstanceLabel] = "true"
	smerdRequest.Labels[labels.RunnerProviderLabel] = provEnum.String()
	smerdRequest.Labels[labels.RunnerScopeLabel] = request.GetScope().String()
	smerdRequest.Labels[labels.RunnerTargetLabel] = request.GetTarget()
	smerdRequest.Labels[labels.RunnerLabelsLabel] = strings.Join(request.GetLabels(), ",")
	smerdRequest.Labels[labels.RunnerBaseUrlLabel] = baseUrl

	return descriptor, smerdRequest, nil
}

// runnerVolumeName derives a per-instance Docker volume name from the
// instance's service name, so multiple runners launched from the same
// builtin descriptor (whose volume name is fixed to a placeholder) don't
// collide in Docker's global volume namespace.
func runnerVolumeName(instanceName string) string {
	return instanceName + "-data"
}

// deriveRunnerName builds a unique-enough runner name from the target plus
// a short random suffix, so re-registering the same target never collides
// with a still-registered prior runner in the provider's own runner list.
func deriveRunnerName(target string) string {
	suffix := uuid.NewString()[:runnerNameSuffixLen]

	return sanitizeRunnerTarget(target) + "-" + suffix
}

// sanitizeRunnerTarget replaces every "/" with "-" and lowercases the
// result - the only character an owner/repo or owner target string can
// carry that a runner name cannot.
func sanitizeRunnerTarget(target string) string {
	return strings.ToLower(strings.ReplaceAll(target, "/", "-"))
}

// waitForRunnerDeployJob blocks until the create_smerd task that the deploy
// watcher (internal/workers/deploy_watcher.go) runs for this runner's
// container reaches a terminal status - the exact same task and entity id
// deployWatcher.deploy's runTask awaits. deployRunnerJob only schedules a
// velez.deployments row; without this job create_runner's own task reaches
// DONE the moment that row is written, well before the container actually
// exists. Reuses the taskWatcher interface declared in
// create_registry_instance.go.
type waitForRunnerDeployJob struct {
	jobsEngine taskWatcher

	req          createRunnerRequestAccessor
	instanceName string
}

func (j *waitForRunnerDeployJob) Do(ctx context.Context) error {
	entityID := SmerdEntityID(j.req.GetRequest().GetEnvironment(), j.instanceName)

	watchCtx, cancel := context.WithTimeout(ctx, runnerDeployWaitTimeout)
	defer cancel()

	var finalTask tasks_queries.VelezTask

	for task := range j.jobsEngine.Watch(watchCtx, entityID, CreateSmerdAction) {
		finalTask = task
	}

	isDone := finalTask.Status == tasks_queries.VelezTaskStatusDONE
	isFailed := finalTask.Status == tasks_queries.VelezTaskStatusFAILED

	if !isDone && !isFailed && watchCtx.Err() != nil {
		return rerrors.Wrapf(
			watchCtx.Err(),
			"timed out waiting for runner container to deploy, last status: %q",
			finalTask.Status,
		)
	}

	if isFailed {
		return rerrors.Wrap(user_errors.ErrTaskFailed, finalTask.Error.String)
	}

	return nil
}

// registerRunnerRowJob upserts the velez.runners row once the container is
// confirmed deployed - see waitForRunnerDeployJob's doc comment on why this
// must not run any earlier.
type registerRunnerRowJob struct {
	services storage.ServicesStorage
	runners  storage.RunnersStorage

	instanceName string
	req          createRunnerRequestAccessor
}

func (j *registerRunnerRowJob) Do(ctx context.Context) error {
	svc, err := j.services.GetByName(ctx, j.instanceName)
	if err != nil {
		return rerrors.Wrap(err, "error getting runner service")
	}

	request := j.req.GetRequest()
	provider, _, baseUrl := runnerProviderConfig(request)

	upsertReq := domain.UpsertRunnerReq{
		ServiceID: svc.ID,
		Provider:  provider.String(),
		Scope:     request.GetScope().String(),
		Target:    request.GetTarget(),
		Labels:    request.GetLabels(),
		SecretRef: domain.RunnerAccessTokenSecretRef(j.instanceName).String(),
		BaseUrl:   baseUrl,
	}

	_, err = j.runners.UpsertRunner(ctx, upsertReq)
	if err != nil {
		return rerrors.Wrap(err, "error upserting runner row")
	}

	return nil
}
