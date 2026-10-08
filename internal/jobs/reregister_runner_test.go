package jobs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
)

const (
	testReregisterConfig = "[[runners]]\n  token = \"glrt-old\"\n"
	testReregisterNewKey = "glrt-new"
)

type runnersClusterStorage struct {
	*fakeClusterStorage

	runners storage.RunnersStorage
}

func (s *runnersClusterStorage) Runners() storage.RunnersStorage {
	return s.runners
}

func newReregisterRunnerJob(
	seeder *fakeConfigSeeder, runtime *seededRunnerRuntime, runner domain.Runner,
) (*reregisterRunnerJob, *fakeServicesStorage) {
	services := newFakeServicesStorage()

	_ = services.UpsertService(context.Background(), testDropRunnerName, "")

	clusterStorage := &runnersClusterStorage{
		fakeClusterStorage: &fakeClusterStorage{services: services},
		runners:            &fakeRunnersStorage{runner: runner},
	}

	payload := &velez_api.ReregisterRunnerTaskPayload{
		Request: &velez_api.ReregisterRunner_Request{Name: testDropRunnerName},
	}

	job := &reregisterRunnerJob{
		dataStorage: clusterStorage,
		secrets:     newStaticSecretsStore(),
		runtimes:    newSeededRunnerResolver(runtime),
		seeder:      seeder,
		req:         payload,
	}

	return job, services
}

func Test_ReregisterRunnerJob_GitlabRunnerWithIdRotatesTokenInConfig(t *testing.T) {
	seeder := &fakeConfigSeeder{resetToken: testReregisterNewKey}
	runtime := &seededRunnerRuntime{fakeContainerRuntime: &fakeContainerRuntime{}, config: []byte(testReregisterConfig)}
	job, _ := newReregisterRunnerJob(seeder, runtime, newGitlabRunnerRow(testGitlabRunnerId))

	err := job.Do(t.Context())
	require.NoError(t, err)

	require.Equal(t, []string{"glrt-old"}, seeder.resetCalls)
	require.Contains(t, string(runtime.written), testReregisterNewKey)
	require.NotContains(t, string(runtime.written), "glrt-old")
	require.Equal(t, runnerConfigFileMode, runtime.writtenMode)

	stored, err := job.secrets.Get(t.Context(), domain.RunnerRegistrationTokenSecretRef(testDropRunnerName))
	require.NoError(t, err)
	require.Equal(t, testReregisterNewKey, stored)
}

func Test_ReregisterRunnerJob_ResetFailureLeavesConfigUntouched(t *testing.T) {
	seeder := &fakeConfigSeeder{resetErr: errSeederDown}
	runtime := &seededRunnerRuntime{fakeContainerRuntime: &fakeContainerRuntime{}, config: []byte(testReregisterConfig)}
	job, _ := newReregisterRunnerJob(seeder, runtime, newGitlabRunnerRow(testGitlabRunnerId))

	err := job.Do(t.Context())
	require.ErrorIs(t, err, errSeederDown)
	require.Nil(t, runtime.written)
}
