package jobs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testDropRunnerName   = "gh-runner"
	testDropGitlabConfig = "[[runners]]\n  token = \"glrt-live\"\n"
)

func Test_DropRunnerHandler_ActionAndContext(t *testing.T) {
	h := NewDropRunnerHandler(nil, nil, nil, nil)

	require.Equal(t, DropRunnerAction, h.Action())

	_, ok := h.NewContext().(*velez_api.DropRunnerTaskPayload)
	require.True(t, ok)
}

func Test_DropRunnerHandler_BuildJobsOrdersUnregisterThenBuildkitThenRowThenServiceThenSecret(t *testing.T) {
	stg := storage.NewStorageContainer(&fakeClusterStorage{})
	h := NewDropRunnerHandler(stg, nil, nil, nil)

	namedJobs := h.BuildJobs(&velez_api.DropRunnerTaskPayload{Name: testDropRunnerName})

	names := make([]string, 0, len(namedJobs))
	for _, namedJob := range namedJobs {
		names = append(names, namedJob.Name)
	}

	require.Equal(t,
		[]string{stepUnregisterRunner, stepDropRunnerBuildkit, stepDeleteRunnerRow, stepRemoveRunnerService,
			stepDeleteRunnerSecret, stepDeleteRunnerRegistrationToken},
		names,
	)
}

func Test_DeleteRunnerSecretJob_RunTwiceDeletesSecretOnce(t *testing.T) {
	store := newStaticSecretsStore()
	ref := domain.RunnerAccessTokenSecretRef(testDropRunnerName)

	err := store.Put(t.Context(), ref, "token")
	require.NoError(t, err)

	job := &deleteRunnerSecretJob{secrets: store, secretRef: ref}

	err = job.Do(t.Context())
	require.NoError(t, err)

	err = job.Do(t.Context())
	require.NoError(t, err)

	_, err = store.Get(t.Context(), ref)
	require.ErrorIs(t, err, user_errors.ErrSecretNotFound)
}

func newUnregisterRunnerJob(
	seeder *fakeConfigSeeder, runtime *seededRunnerRuntime, runner domain.Runner,
) (*unregisterRunnerJob, secrets.Store) {
	services := newFakeServicesStorage()

	_ = services.UpsertService(context.Background(), testDropRunnerName, "")

	store := newStaticSecretsStore()

	job := &unregisterRunnerJob{
		services: services,
		runners:  &fakeRunnersStorage{runner: runner},
		secrets:  store,
		runtimes: newSeededRunnerResolver(runtime),
		seeder:   seeder,
		name:     testDropRunnerName,
	}

	return job, store
}

func newGitlabRunnerRow(gitlabRunnerId int64) domain.Runner {
	return domain.Runner{
		Provider:       velez_api.RunnerProvider_GITLAB.String(),
		BaseUrl:        testGitlabBaseUrl,
		GitlabRunnerId: gitlabRunnerId,
	}
}

func Test_UnregisterRunnerJob_DeletesRunnerWithTokenFromConfig(t *testing.T) {
	seeder := &fakeConfigSeeder{}
	runtime := &seededRunnerRuntime{fakeContainerRuntime: &fakeContainerRuntime{}, config: []byte(testDropGitlabConfig)}
	job, _ := newUnregisterRunnerJob(seeder, runtime, newGitlabRunnerRow(testGitlabRunnerId))

	err := job.Do(t.Context())
	require.NoError(t, err)
	require.Equal(t, []string{"glrt-live"}, seeder.deleteCalls)
}

func Test_UnregisterRunnerJob_FallsBackToStoredTokenWhenContainerMissing(t *testing.T) {
	seeder := &fakeConfigSeeder{}
	runtime := &seededRunnerRuntime{fakeContainerRuntime: &fakeContainerRuntime{}, copyFromErr: errSeederDown}
	job, store := newUnregisterRunnerJob(seeder, runtime, newGitlabRunnerRow(testGitlabRunnerId))

	err := store.Put(t.Context(), domain.RunnerRegistrationTokenSecretRef(testDropRunnerName), "glrt-stored")
	require.NoError(t, err)

	err = job.Do(t.Context())
	require.NoError(t, err)
	require.Equal(t, []string{"glrt-stored"}, seeder.deleteCalls)
}

func Test_UnregisterRunnerJob_GitlabFailureDoesNotBlockDrop(t *testing.T) {
	seeder := &fakeConfigSeeder{deleteErr: errSeederDown}
	runtime := &seededRunnerRuntime{fakeContainerRuntime: &fakeContainerRuntime{}, config: []byte(testDropGitlabConfig)}
	job, _ := newUnregisterRunnerJob(seeder, runtime, newGitlabRunnerRow(testGitlabRunnerId))

	err := job.Do(t.Context())
	require.NoError(t, err)
	require.Len(t, seeder.deleteCalls, 1)
}

func Test_UnregisterRunnerJob_RunnerWithoutGitlabIdIsSkipped(t *testing.T) {
	seeder := &fakeConfigSeeder{}
	runtime := &seededRunnerRuntime{fakeContainerRuntime: &fakeContainerRuntime{}, config: []byte(testDropGitlabConfig)}
	job, _ := newUnregisterRunnerJob(seeder, runtime, newGitlabRunnerRow(0))

	err := job.Do(t.Context())
	require.NoError(t, err)
	require.Empty(t, seeder.deleteCalls)
}

func Test_UnregisterRunnerJob_MissingRowIsSkipped(t *testing.T) {
	seeder := &fakeConfigSeeder{}
	runtime := &seededRunnerRuntime{fakeContainerRuntime: &fakeContainerRuntime{}}
	job, _ := newUnregisterRunnerJob(seeder, runtime, domain.Runner{})

	err := job.Do(t.Context())
	require.NoError(t, err)
	require.Empty(t, seeder.deleteCalls)
}
