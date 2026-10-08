package jobs

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testDropRunnerName = "gh-runner"
)

func Test_DropRunnerHandler_ActionAndContext(t *testing.T) {
	h := NewDropRunnerHandler(nil, nil, nil)

	require.Equal(t, DropRunnerAction, h.Action())

	_, ok := h.NewContext().(*velez_api.DropRunnerTaskPayload)
	require.True(t, ok)
}

func Test_DropRunnerHandler_BuildJobsOrdersRowThenServiceThenSecret(t *testing.T) {
	stg := storage.NewStorageContainer(&fakeClusterStorage{})
	h := NewDropRunnerHandler(stg, nil, nil)

	namedJobs := h.BuildJobs(&velez_api.DropRunnerTaskPayload{Name: testDropRunnerName})

	names := make([]string, 0, len(namedJobs))
	for _, namedJob := range namedJobs {
		names = append(names, namedJob.Name)
	}

	require.Equal(t,
		[]string{stepDeleteRunnerRow, stepRemoveRunnerService, stepDeleteRunnerSecret},
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
