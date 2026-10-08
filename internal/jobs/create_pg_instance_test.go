package jobs

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/service/secrets"
	"go.vervstack.ru/Velez/internal/service/service_manager/pgaas/pgdeploy"
	"go.vervstack.ru/Velez/internal/storage"
	secretsstorage "go.vervstack.ru/Velez/internal/storage/secrets"
)

const (
	testPgInstanceName = "my-pg"
)

type staticSecretsStorage struct {
	storage.Storage

	secretsStorage storage.SecretsStorage
}

func (s *staticSecretsStorage) Secrets() storage.SecretsStorage {
	return s.secretsStorage
}

func newStaticSecretsStore() secrets.Store {
	stg := &staticSecretsStorage{secretsStorage: secretsstorage.NewStatic()}

	return secrets.New(stg)
}

func newCreatePgInstancePayload(ownerService string) *velez_api.CreatePgInstanceTaskPayload {
	request := &velez_api.CreatePgInstance_Request{Name: testPgInstanceName}
	if ownerService != "" {
		request.OwnerService = &ownerService
	}

	return &velez_api.CreatePgInstanceTaskPayload{Request: request}
}

func Test_CreatePgInstanceHandler_Action(t *testing.T) {
	h := NewCreatePgInstanceHandler(nil, nil, nil, nil)

	require.Equal(t, CreatePgInstanceAction, h.Action())
}

func Test_CreatePgInstanceHandler_NewContextIsPayload(t *testing.T) {
	h := NewCreatePgInstanceHandler(nil, nil, nil, nil)

	_, ok := h.NewContext().(*velez_api.CreatePgInstanceTaskPayload)
	require.True(t, ok)
}

func Test_CreatePgInstanceHandler_BuildJobsWithoutOwnerHasNoBindStep(t *testing.T) {
	stg := storage.NewStorageContainer(&fakeClusterStorage{})
	h := NewCreatePgInstanceHandler(stg, nil, nil, nil)

	namedJobs := h.BuildJobs(newCreatePgInstancePayload(""))

	names := make([]string, 0, len(namedJobs))
	for _, namedJob := range namedJobs {
		names = append(names, namedJob.Name)
	}

	require.Equal(t,
		[]string{stepGenerateCredentials, stepDeployPg, stepWaitPg, stepRegisterPgInstanceRow},
		names,
	)
}

func Test_CreatePgInstanceHandler_BuildJobsWithOwnerEndsWithBindStep(t *testing.T) {
	stg := storage.NewStorageContainer(&fakeClusterStorage{})
	h := NewCreatePgInstanceHandler(stg, nil, nil, nil)

	namedJobs := h.BuildJobs(newCreatePgInstancePayload("owner-svc"))

	require.Len(t, namedJobs, 5)
	require.Equal(t, stepBindPgOwnerService, namedJobs[4].Name)
}

func Test_GeneratePgCredentialsJob_StoresPasswordWhenAbsent(t *testing.T) {
	store := newStaticSecretsStore()
	ref := pgSecretRef(testPgInstanceName)
	job := &generatePgCredentialsJob{secrets: store, secretRef: ref}

	err := job.Do(t.Context())
	require.NoError(t, err)

	password, err := store.Get(t.Context(), ref)
	require.NoError(t, err)
	require.NotEmpty(t, password)
}

func Test_GeneratePgCredentialsJob_ReusesExistingPasswordOnRerun(t *testing.T) {
	store := newStaticSecretsStore()
	ref := pgSecretRef(testPgInstanceName)
	job := &generatePgCredentialsJob{secrets: store, secretRef: ref}

	err := job.Do(t.Context())
	require.NoError(t, err)

	first, err := store.Get(t.Context(), ref)
	require.NoError(t, err)

	err = job.Do(t.Context())
	require.NoError(t, err)

	second, err := store.Get(t.Context(), ref)
	require.NoError(t, err)
	require.Equal(t, first, second)
}

func Test_PgSecretRef_UsesPgdeployScopeAndKey(t *testing.T) {
	ref := pgSecretRef(testPgInstanceName)

	require.Equal(t, pgdeploy.SecretScope, ref.Scope)
	require.Equal(t, testPgInstanceName, ref.Owner)
	require.Equal(t, pgdeploy.SecretKey, ref.Key)
}
