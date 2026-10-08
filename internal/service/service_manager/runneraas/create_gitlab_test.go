package runneraas

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/service/service_manager/runneraas/providers"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testRunnerName   = "artel"
	testInstanceName = "gitlab_runner_artel"
	testPat          = "glpat-secret"
	testMintedToken  = "glrt-minted"
	testMintedId     = int64(77)
	testBaseUrl      = "https://gitlab.example.com"

	testDockerSocketAddress = "tcp://dind:2375"
)

var (
	errGitlabDown  = rerrors.New("gitlab is down")
	errEnqueueDown = rerrors.New("engine is down")
)

type fakeSeeder struct {
	createCalls []providers.CreateRunnerReq
	createErr   error
	deleteCalls []string
}

func (f *fakeSeeder) CreateRunner(_ context.Context, req providers.CreateRunnerReq) (providers.CreatedRunner, error) {
	f.createCalls = append(f.createCalls, req)

	if f.createErr != nil {
		return providers.CreatedRunner{}, f.createErr
	}

	return providers.CreatedRunner{Id: testMintedId, Token: testMintedToken}, nil
}

func (f *fakeSeeder) RenderConfig(providers.RenderConfigReq) ([]byte, error) {
	return nil, nil
}

func (f *fakeSeeder) DeleteRunnerByToken(_ context.Context, _, runnerToken string) error {
	f.deleteCalls = append(f.deleteCalls, runnerToken)

	return nil
}

func (f *fakeSeeder) ResetRunnerToken(context.Context, string, string) (string, error) {
	return "", nil
}

type fakeEngine struct {
	jobs.Engine

	enqueued   []any
	enqueueErr error
}

func (f *fakeEngine) EnqueueReplacing(
	_ context.Context, _, _ string, initialContext any,
) (tasks_queries.VelezTask, error) {
	if f.enqueueErr != nil {
		return tasks_queries.VelezTask{}, f.enqueueErr
	}

	f.enqueued = append(f.enqueued, initialContext)

	return tasks_queries.VelezTask{}, nil
}

type fakeSecrets struct {
	values map[domain.SecretRef]string
}

func newFakeSecrets() *fakeSecrets {
	return &fakeSecrets{values: map[domain.SecretRef]string{}}
}

func (f *fakeSecrets) Put(_ context.Context, ref domain.SecretRef, value string) error {
	f.values[ref] = value

	return nil
}

func (f *fakeSecrets) Get(_ context.Context, ref domain.SecretRef) (string, error) {
	value, ok := f.values[ref]
	if !ok {
		return "", user_errors.ErrSecretNotFound
	}

	return value, nil
}

func (f *fakeSecrets) Delete(_ context.Context, ref domain.SecretRef) error {
	delete(f.values, ref)

	return nil
}

func (f *fakeSecrets) ListRefs(context.Context, string, string) ([]domain.SecretRef, error) {
	return nil, nil
}

func newGitlabRunnerReq() domain.CreateRunnerReq {
	return domain.CreateRunnerReq{
		Name:        testRunnerName,
		Provider:    velez_api.RunnerProvider_GITLAB,
		Scope:       velez_api.RunnerScope_ORG,
		Target:      "artel",
		Labels:      []string{"docker"},
		AccessToken: testPat,
		BaseUrl:     testBaseUrl,
		Concurrent:  1,

		DockerSocketAddress: testDockerSocketAddress,
	}
}

func newGithubRunnerReq() domain.CreateRunnerReq {
	return domain.CreateRunnerReq{
		Name:        testRunnerName,
		Provider:    velez_api.RunnerProvider_GITHUB,
		Scope:       velez_api.RunnerScope_REPO,
		Target:      "owner/repo",
		AccessToken: "ghp-secret",

		DockerSocketAddress: testDockerSocketAddress,
	}
}

func newTestService(seeder *fakeSeeder, engine *fakeEngine, store *fakeSecrets) *RunneraasService {
	return &RunneraasService{
		secrets:      store,
		jobsEngine:   engine,
		gitlabSeeder: seeder,
	}
}

func Test_CreateRunner_Gitlab_MintsBeforeEnqueueAndKeepsPatOutOfPayload(t *testing.T) {
	seeder := &fakeSeeder{}
	engine := &fakeEngine{}
	store := newFakeSecrets()
	svc := newTestService(seeder, engine, store)

	err := svc.CreateRunner(t.Context(), newGitlabRunnerReq())
	require.NoError(t, err)

	require.Len(t, seeder.createCalls, 1)
	require.Equal(t, testPat, seeder.createCalls[0].PersonalAccessToken)
	require.Equal(t, testInstanceName, seeder.createCalls[0].Description)
	require.Equal(t, testBaseUrl, seeder.createCalls[0].BaseUrl)

	require.Len(t, engine.enqueued, 1)

	payload, ok := engine.enqueued[0].(*velez_api.CreateRunnerTaskPayload)
	require.True(t, ok)
	require.Empty(t, payload.GetRequest().GetGitlab().GetAccessToken())
	require.Equal(t, testMintedToken, payload.GetRegistrationToken())
	require.Equal(t, testMintedId, payload.GetGitlabRunnerId())

	stored, err := store.Get(t.Context(), domain.RunnerRegistrationTokenSecretRef(testInstanceName))
	require.NoError(t, err)
	require.Equal(t, testMintedToken, stored)
}

func Test_CreateRunner_Gitlab_SeederErrorEnqueuesAndStoresNothing(t *testing.T) {
	seeder := &fakeSeeder{createErr: errGitlabDown}
	engine := &fakeEngine{}
	store := newFakeSecrets()
	svc := newTestService(seeder, engine, store)

	err := svc.CreateRunner(t.Context(), newGitlabRunnerReq())
	require.ErrorIs(t, err, errGitlabDown)

	require.Empty(t, engine.enqueued)
	require.Empty(t, store.values)
}

func Test_CreateRunner_Gitlab_EnqueueFailureDropsMintedRunnerAndSecret(t *testing.T) {
	seeder := &fakeSeeder{}
	engine := &fakeEngine{enqueueErr: errEnqueueDown}
	store := newFakeSecrets()
	svc := newTestService(seeder, engine, store)

	err := svc.CreateRunner(t.Context(), newGitlabRunnerReq())
	require.ErrorIs(t, err, errEnqueueDown)

	require.Equal(t, []string{testMintedToken}, seeder.deleteCalls)
	require.Empty(t, store.values)
}

func Test_CreateRunner_Github_KeepsPatInPayloadAndSkipsSeeder(t *testing.T) {
	seeder := &fakeSeeder{}
	engine := &fakeEngine{}
	store := newFakeSecrets()
	svc := newTestService(seeder, engine, store)

	err := svc.CreateRunner(t.Context(), newGithubRunnerReq())
	require.NoError(t, err)

	require.Empty(t, seeder.createCalls)
	require.Empty(t, store.values)

	require.Len(t, engine.enqueued, 1)

	payload, ok := engine.enqueued[0].(*velez_api.CreateRunnerTaskPayload)
	require.True(t, ok)
	require.Equal(t, "ghp-secret", payload.GetRequest().GetGithub().GetAccessToken())
	require.Empty(t, payload.GetRegistrationToken())
}
