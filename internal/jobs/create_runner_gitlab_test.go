package jobs

import (
	"context"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/service_manager/runneraas/providers"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testGitlabRunnerName  = "gitlab_runner_artel"
	testGitlabPat         = "glpat-secret"
	testGitlabRunnerToken = "glrt-minted"
	testGitlabRunnerId    = int64(77)
	testGitlabBaseUrl     = "https://gitlab.example.com"
	testRenderedConfig    = "rendered-config"
)

var (
	errSeederDown                                    = rerrors.New("gitlab is down")
	_             container_runtime.ContainerRuntime = (*seededRunnerRuntime)(nil)
	_             storage.RunnersStorage             = (*fakeRunnersStorage)(nil)
)

type fakeConfigSeeder struct {
	createCalls []providers.CreateRunnerReq
	createErr   error

	renderCalls []providers.RenderConfigReq

	deleteCalls []string
	deleteErr   error

	resetCalls []string
	resetToken string
	resetErr   error
}

func (f *fakeConfigSeeder) CreateRunner(
	_ context.Context, req providers.CreateRunnerReq,
) (providers.CreatedRunner, error) {
	f.createCalls = append(f.createCalls, req)

	if f.createErr != nil {
		return providers.CreatedRunner{}, f.createErr
	}

	return providers.CreatedRunner{Id: testGitlabRunnerId, Token: testGitlabRunnerToken}, nil
}

func (f *fakeConfigSeeder) RenderConfig(req providers.RenderConfigReq) ([]byte, error) {
	f.renderCalls = append(f.renderCalls, req)

	return []byte(testRenderedConfig), nil
}

func (f *fakeConfigSeeder) DeleteRunnerByToken(_ context.Context, _, runnerToken string) error {
	f.deleteCalls = append(f.deleteCalls, runnerToken)

	return f.deleteErr
}

func (f *fakeConfigSeeder) ResetRunnerToken(_ context.Context, _, runnerToken string) (string, error) {
	f.resetCalls = append(f.resetCalls, runnerToken)

	return f.resetToken, f.resetErr
}

// seededRunnerRuntime keeps one config.toml in memory on top of fakeContainerRuntime.
type seededRunnerRuntime struct {
	*fakeContainerRuntime

	config      []byte
	copyFromErr error

	written     []byte
	writtenMode fs.FileMode
}

func (r *seededRunnerRuntime) CopyFromContainer(context.Context, string, string) ([]byte, error) {
	return r.config, r.copyFromErr
}

func (r *seededRunnerRuntime) CopyToContainer(
	_ context.Context, _, path string, content []byte, mode fs.FileMode,
) error {
	r.written = content
	r.writtenMode = mode

	return nil
}

func newSeededRunnerResolver(runtime *seededRunnerRuntime) proxyRunnerResolver {
	return proxyRunnerResolver{runtime: runtime}
}

type fakeRunnersStorage struct {
	upserted []domain.UpsertRunnerReq
	runner   domain.Runner
	deleted  []int64
}

func (f *fakeRunnersStorage) UpsertRunner(_ context.Context, req domain.UpsertRunnerReq) (domain.Runner, error) {
	f.upserted = append(f.upserted, req)

	return domain.Runner{}, nil
}

func (f *fakeRunnersStorage) GetRunnerByServiceID(context.Context, int64) (domain.Runner, error) {
	if f.runner.Provider == "" {
		return domain.Runner{}, user_errors.ErrStorageNotFound
	}

	return f.runner, nil
}

func (f *fakeRunnersStorage) ListRunners(context.Context) ([]domain.Runner, error) {
	return nil, nil
}

func (f *fakeRunnersStorage) DeleteRunner(_ context.Context, serviceID int64) error {
	f.deleted = append(f.deleted, serviceID)

	return nil
}

func (f *fakeRunnersStorage) SetRunnerBuildkit(context.Context, int64, bool) error {
	return nil
}

func newGitlabRunnerPayload() *velez_api.CreateRunnerTaskPayload {
	baseUrl := testGitlabBaseUrl

	return &velez_api.CreateRunnerTaskPayload{
		Request: &velez_api.CreateRunner_Request{
			Name:   testRunnerArtel,
			Scope:  velez_api.RunnerScope_ORG,
			Target: "artel",
			Labels: []string{"docker"},
			ProviderConfig: &velez_api.CreateRunner_Request_Gitlab{
				Gitlab: &velez_api.GitlabConfig{AccessToken: testGitlabPat, BaseUrl: &baseUrl},
			},
		},
	}
}

func Test_WriteRunnerConfigJob_RendersAndWritesIntoLoaderContainer(t *testing.T) {
	seeder := &fakeConfigSeeder{}
	payload := newGitlabRunnerPayload()
	payload.SetRegistrationToken(testGitlabRunnerToken)

	docker := newFakeDocker()
	containerAPI := newFakeContainerAPI()
	docker.withClient(containerAPI)

	loader := runnerLoaderRef{instanceName: testGitlabRunnerName}
	runtime := &seededRunnerRuntime{fakeContainerRuntime: &fakeContainerRuntime{docker: docker}}

	job := &writeRunnerConfigJob{
		runtimes:     newSeededRunnerResolver(runtime),
		copyAPI:      containerAPI,
		seeder:       seeder,
		req:          payload,
		token:        payload,
		loader:       loader,
		instanceName: testGitlabRunnerName,
	}

	err := job.Do(t.Context())
	require.NoError(t, err)

	require.Len(t, seeder.renderCalls, 1)

	renderReq := seeder.renderCalls[0]
	require.Equal(t, testGitlabRunnerToken, renderReq.RunnerToken)
	require.Equal(t, testGitlabBaseUrl, renderReq.BaseUrl)
	require.Equal(t, testGitlabRunnerName+"-cache", renderReq.CacheVolumeName)
	require.Equal(t, int32(1), renderReq.Concurrent)

	require.Len(t, containerAPI.copyCalledWith, 1)
	require.Equal(t, testGitlabRunnerName+"-data_loader", containerAPI.copyCalledWith[0].containerID)
	require.Equal(t, "/etc/gitlab-runner", containerAPI.copyCalledWith[0].dstPath)
}

func Test_RunnerLoaderRef_ContainerIdIsLoaderName(t *testing.T) {
	loader := runnerLoaderRef{instanceName: testGitlabRunnerName}

	require.Equal(t, runnerVolumeName(testGitlabRunnerName), loader.GetVolumeName())
	require.Equal(t, loader.GetVolumeName()+loaderContainerSuffix, loader.GetContainerId())
}

func Test_RegisterRunnerRowJob_StoresGitlabIdAndClearsRegistrationToken(t *testing.T) {
	payload := newGitlabRunnerPayload()
	payload.SetRegistrationToken(testGitlabRunnerToken)
	payload.SetGitlabRunnerId(testGitlabRunnerId)

	services := newFakeServicesStorage()
	err := services.UpsertService(t.Context(), testGitlabRunnerName, "")
	require.NoError(t, err)

	runners := &fakeRunnersStorage{}

	job := &registerRunnerRowJob{
		services:     services,
		runners:      runners,
		instanceName: testGitlabRunnerName,
		req:          payload,
		ctx:          payload,
	}

	err = job.Do(t.Context())
	require.NoError(t, err)

	require.Len(t, runners.upserted, 1)
	require.Equal(t, testGitlabRunnerId, runners.upserted[0].GitlabRunnerId)
	require.Equal(t, domain.RunnerAccessTokenSecretRef(testGitlabRunnerName).String(), runners.upserted[0].SecretRef)
	require.Empty(t, payload.GetRegistrationToken())
}

func Test_CreateRunnerHandler_BuildJobs_GitlabSeedsConfigWithoutMintOrRegisterStep(t *testing.T) {
	stg := storage.NewStorageContainer(&fakeClusterStorage{})
	handler := NewCreateRunnerHandler(newFakeNodeClients(newFakeDocker()), stg, nil, nil, nil, nil)

	namedJobs := handler.BuildJobs(newGitlabRunnerPayload())

	require.Equal(t,
		[]string{
			stepCreateLoaderContainer, stepStartSidecar, stepWriteRunnerConfig, stepDropContainer,
			stepDeployRunner, stepWaitForRunnerDeploy, stepSyncRunnerProxy, stepRegisterRunnerRow,
		},
		namedJobNames(namedJobs),
	)
}

func Test_CreateRunnerHandler_BuildJobs_GithubKeepsSelfRegisteringChain(t *testing.T) {
	stg := storage.NewStorageContainer(&fakeClusterStorage{})
	handler := NewCreateRunnerHandler(newFakeNodeClients(newFakeDocker()), stg, nil, nil, nil, nil)

	payload := &velez_api.CreateRunnerTaskPayload{
		Request: &velez_api.CreateRunner_Request{
			Name:           "artel",
			ProviderConfig: &velez_api.CreateRunner_Request_Github{Github: &velez_api.GithubConfig{AccessToken: "ghp"}},
		},
	}

	namedJobs := handler.BuildJobs(payload)

	require.Equal(t,
		[]string{stepMintRunnerToken, stepDeployRunner, stepWaitForRunnerDeploy, stepRegisterRunnerRow},
		namedJobNames(namedJobs),
	)
}

func Test_ConfigSeederFor_OnlyGitlabSeeds(t *testing.T) {
	require.NotNil(t, configSeederFor(velez_api.RunnerProvider_GITLAB))
	require.Nil(t, configSeederFor(velez_api.RunnerProvider_GITHUB))
}

func namedJobNames(namedJobs []NamedJob) []string {
	names := make([]string, 0, len(namedJobs))
	for _, namedJob := range namedJobs {
		names = append(names, namedJob.Name)
	}

	return names
}
