package jobs

import (
	"context"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
	"go.vervstack.ru/Velez/internal/gitlab_runner_config"
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
	stepWriteRunnerConfig   = "write_runner_config"
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

	// runnerRegistrationTokenEnvVar carries the registration token into the
	// runner container's own env, written regardless of provider - for GitLab
	// it is the minted glrt- runner token. Velez-internal, never read by the
	// runner process itself: the
	// single-node/dev secrets backend reads it back on GetRunnerCredentials/
	// reregister after a process restart wipes its in-memory copy of the
	// token, mirroring the precedent already established for pgaas's
	// password. See internal/storage/local_storage/secrets.go.
	runnerRegistrationTokenEnvVar = "VELEZ_RUNNER_REGISTRATION_TOKEN"

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

type runnerRegistrationTokenClearer interface {
	ClearRegistrationToken()
}

type runnerGitlabIdAccessor interface {
	GetGitlabRunnerId() int64
	SetGitlabRunnerId(v int64)
}

// runnerLoaderRef adapts the runner data volume to copyToVolumeRequestAccessor
// and containerIDAccessor for the loader-container jobs. The loader container
// id is its name, which Docker accepts wherever an id is expected, so a task
// resumed after create_loader_container was checkpointed still finds it.
type runnerLoaderRef struct {
	instanceName string
}

func (r runnerLoaderRef) GetVolumeName() string {
	return runnerVolumeName(r.instanceName)
}

func (runnerLoaderRef) GetPathToFiles() map[string][]byte {
	return nil
}

func (r runnerLoaderRef) GetContainerId() string {
	return r.GetVolumeName() + loaderContainerSuffix
}

func (runnerLoaderRef) SetContainerId(string) {}

type createRunnerHandler struct {
	nodeClients  node_clients.NodeClients
	dataStorage  storage.Storage
	secretsStore secrets.Store
	vervServices service.VervServicesService
	// jobsEngine lets waitForRunnerDeployJob watch the create_smerd task the
	// deploy watcher runs for this runner's container.
	jobsEngine Engine
	// runtimes resolves the ContainerRuntime the config-seeding and proxy-sync
	// jobs act through.
	runtimes container_runtime.RuntimeResolver
}

func NewCreateRunnerHandler(
	nodeClients node_clients.NodeClients,
	dataStorage storage.Storage,
	secretsStore secrets.Store,
	vervServices service.VervServicesService,
	jobsEngine Engine,
	runtimes container_runtime.RuntimeResolver,
) TaskHandler {
	return &createRunnerHandler{
		nodeClients:  nodeClients,
		dataStorage:  dataStorage,
		secretsStore: secretsStore,
		vervServices: vervServices,
		jobsEngine:   jobsEngine,
		runtimes:     runtimes,
	}
}

func (h *createRunnerHandler) Action() string {
	return CreateRunnerAction
}

func (h *createRunnerHandler) NewContext() TaskContext {
	return &velez_api.CreateRunnerTaskPayload{}
}

// BuildJobs mirrors create_registry_instance.go's shape, simplified: a
// runner has no ports/htpasswd/UI sidecar to resolve. GitHub self-registers
// from the container env, so its chain is mint/deploy/wait/row. GitLab has no
// mint or register step: CreateRunner (runneraas) creates the runner through
// the GitLab API before enqueueing, so the personal access token never reaches
// the persisted task payload, and the full config.toml is written into the
// data volume before the container first starts, so the container comes up
// already registered; sync_runner_proxy follows the deploy because the job
// proxy env is derived from the live container.
func (h *createRunnerHandler) BuildJobs(taskCtx TaskContext) []NamedJob {
	payload, ok := taskCtx.(*velez_api.CreateRunnerTaskPayload)
	if !ok {
		panic("create_runner: BuildJobs called with mismatched TaskContext type")
	}

	provider, _, _ := runnerProviderConfig(payload.GetRequest())
	instanceName := RunnerInstanceName(provider, payload.GetRequest().GetName())

	jobs := h.baseJobs(payload, provider, instanceName)

	if !payload.GetRequest().GetIsBuildkitEnabled() {
		return jobs
	}

	buildkit := runnerBuildkitJobs{dataStorage: h.dataStorage, runtimes: h.runtimes}
	target := createdRunnerTarget{req: payload, instanceName: instanceName}

	return append(jobs, buildkit.enable(target)...)
}

// baseJobs is the chain that creates and registers the runner itself.
func (h *createRunnerHandler) baseJobs(
	payload *velez_api.CreateRunnerTaskPayload, provider velez_api.RunnerProvider, instanceName string,
) []NamedJob {
	seederProvider := configSeederFor(provider)

	deployJob := NamedJob{
		Name: stepDeployRunner,
		Job: &deployRunnerJob{
			boxes:        h.dataStorage.ResourceBoxes(),
			vervServices: h.vervServices,
			secrets:      h.secretsStore,
			jobsEngine:   h.jobsEngine,
			req:          payload,
			ctx:          payload,
			instanceName: instanceName,
		},
	}

	waitJob := NamedJob{
		Name: stepWaitForRunnerDeploy,
		Job: &waitForRunnerDeployJob{
			jobsEngine:   h.jobsEngine,
			req:          payload,
			ctx:          payload,
			instanceName: instanceName,
		},
	}

	rowJob := NamedJob{
		Name: stepRegisterRunnerRow,
		Job: &registerRunnerRowJob{
			services:      h.dataStorage.Services(),
			runners:       h.dataStorage.Runners(),
			dindInstances: h.dataStorage.DindInstances(),
			instanceName:  instanceName,
			req:           payload,
			ctx:           payload,
		},
	}

	if provider != velez_api.RunnerProvider_GITLAB {
		mintJob := NamedJob{
			Name: stepMintRunnerToken,
			Job: &mintRunnerTokenJob{
				secrets:      h.secretsStore,
				req:          payload,
				ctx:          payload,
				instanceName: instanceName,
			},
		}

		return []NamedJob{mintJob, deployJob, waitJob, rowJob}
	}

	namedJobs := h.buildSeedConfigJobs(payload, instanceName, seederProvider)

	syncProxyJob := NamedJob{
		Name: stepSyncRunnerProxy,
		Job: &syncCreatedRunnerProxyJob{
			runtimes:     h.runtimes,
			req:          payload,
			instanceName: instanceName,
		},
	}

	return append(namedJobs, deployJob, waitJob, syncProxyJob, rowJob)
}

// buildSeedConfigJobs writes the rendered config.toml into the runner data
// volume through a throwaway loader container, the same way
// create_registry_instance writes its htpasswd file.
func (h *createRunnerHandler) buildSeedConfigJobs(
	payload *velez_api.CreateRunnerTaskPayload,
	instanceName string,
	seederProvider providers.ConfigSeeder,
) []NamedJob {
	loaderRef := runnerLoaderRef{instanceName: instanceName}
	folders := mountedFolders(loaderRef.GetVolumeName(), []string{gitlab_runner_config.ConfigPath})

	return []NamedJob{
		{
			Name: stepCreateLoaderContainer,
			Job: &createLoaderContainerJob{
				nodeClients: h.nodeClients,
				req:         loaderRef,
				folders:     folders,
				ctx:         loaderRef,
			},
		},
		{
			Name: stepStartSidecar,
			Job: &startLoaderContainerJob{
				dockerAPI: h.nodeClients.Docker().Client(),
				ctx:       loaderRef,
			},
		},
		{
			Name: stepWriteRunnerConfig,
			Job: &writeRunnerConfigJob{
				runtimes:     h.runtimes,
				copyAPI:      h.nodeClients.Docker().Client(),
				seeder:       seederProvider,
				req:          payload,
				token:        payload,
				loader:       loaderRef,
				instanceName: instanceName,
			},
		},
		{
			Name: stepDropContainer,
			Job: &dropLoaderContainerJob{
				docker: h.nodeClients.Docker(),
				ctx:    loaderRef,
			},
		},
	}
}

// configSeederFor returns the provider's ConfigSeeder, or nil for a provider
// that registers itself from the container env (GitHub) or is unspecified.
func configSeederFor(provider velez_api.RunnerProvider) providers.ConfigSeeder {
	runnerProvider, err := providers.For(provider)
	if err != nil {
		return nil
	}

	configSeeder, ok := runnerProvider.(providers.ConfigSeeder)
	if !ok {
		return nil
	}

	return configSeeder
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

func RunnerInstanceName(provider velez_api.RunnerProvider, name string) string {
	return runnerNamePrefix(provider) + name
}

// mintRunnerTokenJob stores the caller's access token and mints the GitHub
// registration token deploy_runner needs through the GitHub API. Skips minting
// on a resumed task that already has a token - see
// runnerRegistrationTokenAccessor's doc comment. GitLab runners are minted
// before the task is enqueued - see runneraas.CreateRunner.
type mintRunnerTokenJob struct {
	secrets secrets.Store

	req          createRunnerRequestAccessor
	ctx          runnerRegistrationTokenAccessor
	instanceName string
}

func (j *mintRunnerTokenJob) Do(ctx context.Context) error {
	request := j.req.GetRequest()

	provider, accessToken, baseUrl := runnerProviderConfig(request)

	runnerProvider, err := providers.For(provider)
	if err != nil {
		return rerrors.Wrap(err, "error resolving runner provider")
	}

	err = j.secrets.Put(ctx, domain.RunnerAccessTokenSecretRef(j.instanceName), accessToken)
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

	return j.storeRegistrationToken(ctx)
}

func (j *mintRunnerTokenJob) storeRegistrationToken(ctx context.Context) error {
	err := j.secrets.Put(ctx, domain.RunnerRegistrationTokenSecretRef(j.instanceName), j.ctx.GetRegistrationToken())
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
	jobsEngine   taskWatcher

	req createRunnerRequestAccessor
	ctx interface {
		runnerRegistrationTokenAccessor
		runnerGitlabIdAccessor
		deployBaselineAccessor
	}
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
		ctx, j.boxes, runnerProvider, request, j.ctx.GetRegistrationToken(), j.ctx.GetGitlabRunnerId(), name,
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
	switch {
	case request.GetDindName() != "":
		smerdRequest.Env[runnerDockerHostEnvVar] = domain.DindAddress(request.GetDindName())
		attachDindNetwork(smerdRequest, request.GetDindName())
	case request.GetDockerSocketAddress() != "":
		smerdRequest.Env[runnerDockerHostEnvVar] = request.GetDockerSocketAddress()
	default:
		err = j.secrets.Put(ctx, domain.DockerSocketGrantSecretRef(name), "true")
		if err != nil {
			return rerrors.Wrap(err, "error putting docker socket grant secret")
		}
	}

	err = recordDeployBaseline(ctx, j.jobsEngine, j.ctx, SmerdEntityID(request.GetEnvironment(), name))
	if err != nil {
		return err
	}

	deployReq := domain.CreateDeployReq{
		ServiceName:    name,
		DisplayName:    request.GetName(),
		VervDescriptor: &descriptor,
		LaunchSmerd:    domain.LaunchSmerd{CreateSmerd_Request: smerdRequest},
	}

	err = j.vervServices.CreateNewDeploy(ctx, deployReq)
	if err != nil {
		return rerrors.Wrap(err, "error creating runner deploy")
	}

	return nil
}

func attachDindNetwork(request *velez_api.CreateSmerd_Request, dindName string) {
	if request.GetSettings() == nil {
		request.Settings = &velez_api.Container_Settings{}
	}

	networkBind := &velez_api.NetworkBind{NetworkName: domain.DindNetworkName(dindName)}

	request.Settings.Network = append(request.Settings.Network, networkBind)
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
	gitlabRunnerId int64,
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
	smerdRequest.Env[runnerRegistrationTokenEnvVar] = registrationToken

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
	smerdRequest.Labels[labels.DisplayNameLabel] = request.GetName()
	smerdRequest.Labels[labels.RunnerInstanceLabel] = "true"
	smerdRequest.Labels[labels.RunnerProviderLabel] = provEnum.String()
	smerdRequest.Labels[labels.RunnerScopeLabel] = request.GetScope().String()
	smerdRequest.Labels[labels.RunnerTargetLabel] = request.GetTarget()
	smerdRequest.Labels[labels.RunnerLabelsLabel] = strings.Join(request.GetLabels(), ",")
	smerdRequest.Labels[labels.RunnerBaseUrlLabel] = baseUrl

	if gitlabRunnerId > 0 {
		smerdRequest.Labels[labels.RunnerGitlabIdLabel] = strconv.FormatInt(gitlabRunnerId, 10)
	}

	if request.GetDindName() != "" {
		smerdRequest.Labels[labels.RunnerDindLabel] = request.GetDindName()
	}

	return descriptor, smerdRequest, nil
}

// runnerVolumeName derives a per-instance Docker volume name from the
// instance's service name, so multiple runners launched from the same
// builtin descriptor (whose volume name is fixed to a placeholder) don't
// collide in Docker's global volume namespace.
func runnerVolumeName(instanceName string) string {
	return instanceName + "-data"
}

func runnerCacheVolumeName(instanceName string) string {
	return instanceName + "-cache"
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
	ctx          deployBaselineAccessor
	instanceName string
}

func (j *waitForRunnerDeployJob) Do(ctx context.Context) error {
	entityID := SmerdEntityID(j.req.GetRequest().GetEnvironment(), j.instanceName)

	watchCtx, cancel := context.WithTimeout(ctx, runnerDeployWaitTimeout)
	defer cancel()

	var finalTask tasks_queries.VelezTask

	baseline := j.ctx.GetDeployBaselineTaskId()

	for task := range j.jobsEngine.WatchAfter(watchCtx, entityID, CreateSmerdAction, baseline) {
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

// writeRunnerConfigJob renders the runner's config.toml and writes it into
// the data volume mounted in the loader container. Rendering and writing are
// one job so the glrt- token only ever lives in the task context, never in an
// intermediate buffer a resumed task would lose.
type writeRunnerConfigJob struct {
	runtimes container_runtime.RuntimeResolver
	copyAPI  copyAPI
	seeder   providers.ConfigSeeder

	req          createRunnerRequestAccessor
	token        runnerRegistrationTokenAccessor
	loader       containerIDAccessor
	instanceName string
}

func (j *writeRunnerConfigJob) Do(ctx context.Context) error {
	containerID := j.loader.GetContainerId()
	if containerID == "" {
		return user_errors.ErrContainerIdMissing
	}

	if j.seeder == nil {
		return rerrors.Wrap(user_errors.ErrRunnerProviderUnsupported)
	}

	request := j.req.GetRequest()
	_, _, baseUrl := runnerProviderConfig(request)

	renderReq := providers.RenderConfigReq{
		BaseUrl:         baseUrl,
		RunnerToken:     j.token.GetRegistrationToken(),
		RunnerName:      j.instanceName,
		DockerImage:     request.GetGitlab().GetDockerImage(),
		CacheVolumeName: runnerCacheVolumeName(j.instanceName),
		Concurrent:      effectiveConcurrent(request.GetGitlab().GetConcurrent()),
	}

	content, err := j.seeder.RenderConfig(renderReq)
	if err != nil {
		return rerrors.Wrap(err, "error rendering runner config")
	}

	containerRuntime, err := j.runtimes.Runtime(ctx, "")
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	err = mkdirInContainer(ctx, containerRuntime, containerID, path.Dir(gitlab_runner_config.ConfigPath))
	if err != nil {
		return err
	}

	configMode := int64(runnerConfigFileMode)

	err = writeFileToContainer(ctx, j.copyAPI, containerID, gitlab_runner_config.ConfigPath, content, configMode)
	if err != nil {
		return rerrors.Wrap(err, "error copying runner config to container")
	}

	return nil
}

// syncCreatedRunnerProxyJob mirrors the proxy env of the freshly deployed
// runner container into its config.toml - the job proxy env is derived from
// the live container, so it cannot be rendered when the file is seeded.
type syncCreatedRunnerProxyJob struct {
	runtimes container_runtime.RuntimeResolver

	req          createRunnerRequestAccessor
	instanceName string
}

func (j *syncCreatedRunnerProxyJob) Do(ctx context.Context) error {
	request := j.req.GetRequest()

	provider, _, _ := runnerProviderConfig(request)

	runnerProvider, err := providers.For(provider)
	if err != nil {
		return rerrors.Wrap(err, "error resolving runner provider")
	}

	containerRuntime, err := j.runtimes.Runtime(ctx, request.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	err = runnerProvider.SyncProxy(ctx, containerRuntime, j.instanceName)
	if err != nil {
		return rerrors.Wrap(err, "error syncing runner proxy")
	}

	return nil
}

// registerRunnerRowJob upserts the velez.runners row once the container is
// confirmed deployed - see waitForRunnerDeployJob's doc comment on why this
// must not run any earlier.
type registerRunnerRowJob struct {
	services      storage.ServicesStorage
	runners       storage.RunnersStorage
	dindInstances storage.DindInstancesStorage

	instanceName string
	req          createRunnerRequestAccessor
	ctx          interface {
		runnerGitlabIdAccessor
		runnerRegistrationTokenClearer
	}
}

func (j *registerRunnerRowJob) Do(ctx context.Context) error {
	svc, err := j.services.GetByName(ctx, j.instanceName)
	if err != nil {
		return rerrors.Wrap(err, "error getting runner service")
	}

	request := j.req.GetRequest()
	provider, _, baseUrl := runnerProviderConfig(request)

	dindServiceId, err := j.resolveDindServiceId(ctx, request.GetDindName())
	if err != nil {
		return err
	}

	upsertReq := domain.UpsertRunnerReq{
		ServiceID:           svc.ID,
		DindServiceId:       dindServiceId,
		Provider:            provider.String(),
		Scope:               request.GetScope().String(),
		Target:              request.GetTarget(),
		Labels:              request.GetLabels(),
		SecretRef:           domain.RunnerAccessTokenSecretRef(j.instanceName).String(),
		BaseUrl:             baseUrl,
		DockerImage:         request.GetGitlab().GetDockerImage(),
		DockerSocketAddress: request.GetDockerSocketAddress(),
		Concurrent:          effectiveConcurrent(request.GetGitlab().GetConcurrent()),
		GitlabRunnerId:      j.ctx.GetGitlabRunnerId(),
	}

	_, err = j.runners.UpsertRunner(ctx, upsertReq)
	if err != nil {
		return rerrors.Wrap(err, "error upserting runner row")
	}

	j.ctx.ClearRegistrationToken()

	return nil
}

func (j *registerRunnerRowJob) resolveDindServiceId(ctx context.Context, dindName string) (int64, error) {
	if dindName == "" {
		return 0, nil
	}

	dindSvc, err := j.services.GetByName(ctx, dindName)
	if err != nil {
		return 0, rerrors.Wrap(err, "error getting dind service")
	}

	_, err = j.dindInstances.GetDindInstanceByServiceId(ctx, dindSvc.ID)
	if err != nil {
		return 0, rerrors.Wrap(err, "error getting dind instance")
	}

	return dindSvc.ID, nil
}

// effectiveConcurrent maps an unset (0) concurrent to gitlab-runner's own
// default of 1.
func effectiveConcurrent(concurrent int32) int32 {
	if concurrent < 1 {
		return 1
	}

	return concurrent
}
