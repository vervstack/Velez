package jobs

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	pginstancesstorage "go.vervstack.ru/Velez/internal/storage/pg_instances"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testDropPgServiceName = "pgaas_" + testPgInstanceName
)

func Test_DropPgInstanceHandler_ActionAndContext(t *testing.T) {
	h := NewDropPgInstanceHandler(nil, nil, nil)

	require.Equal(t, DropPgInstanceAction, h.Action())

	_, ok := h.NewContext().(*velez_api.DropPgInstanceTaskPayload)
	require.True(t, ok)
}

func Test_DropPgInstanceHandler_BuildJobsOrdersRowThenServiceThenSecret(t *testing.T) {
	stg := storage.NewStorageContainer(&fakeClusterStorage{})
	h := NewDropPgInstanceHandler(stg, nil, nil)

	namedJobs := h.BuildJobs(&velez_api.DropPgInstanceTaskPayload{Name: testDropPgServiceName})

	names := make([]string, 0, len(namedJobs))
	for _, namedJob := range namedJobs {
		names = append(names, namedJob.Name)
	}

	require.Equal(t,
		[]string{stepDeletePgInstanceRow, stepRemovePgInstanceService, stepDeletePgInstanceSecret},
		names,
	)
}

func Test_DropPgInstanceHandler_SecretStepTargetsBareNameRef(t *testing.T) {
	stg := storage.NewStorageContainer(&fakeClusterStorage{})
	h := NewDropPgInstanceHandler(stg, nil, nil)

	namedJobs := h.BuildJobs(&velez_api.DropPgInstanceTaskPayload{Name: testDropPgServiceName})

	secretJob, ok := namedJobs[2].Job.(*deletePgInstanceSecretJob)
	require.True(t, ok)
	require.Equal(t, pgSecretRef(testPgInstanceName), secretJob.secretRef)
}

func Test_DeletePgInstanceRowJob_RunTwiceRemovesRowOnce(t *testing.T) {
	services := newFakeServicesStorage()
	pgInstances := pginstancesstorage.NewStatic()

	err := services.UpsertService(t.Context(), testDropPgServiceName, testDropPgServiceName)
	require.NoError(t, err)

	svc, err := services.GetByName(t.Context(), testDropPgServiceName)
	require.NoError(t, err)

	upsertReq := domain.UpsertPgInstanceReq{ServiceId: svc.ID, DbName: "my_pg"}

	_, err = pgInstances.UpsertPgInstance(t.Context(), upsertReq)
	require.NoError(t, err)

	job := &deletePgInstanceRowJob{services: services, pgInstances: pgInstances, name: testDropPgServiceName}

	err = job.Do(t.Context())
	require.NoError(t, err)

	err = job.Do(t.Context())
	require.NoError(t, err)

	_, err = pgInstances.GetPgInstanceByServiceID(t.Context(), svc.ID)
	require.ErrorIs(t, err, user_errors.ErrStorageNotFound)
}

func Test_DeletePgInstanceSecretJob_RunTwiceDeletesSecretOnce(t *testing.T) {
	store := newStaticSecretsStore()
	ref := pgSecretRef(testPgInstanceName)

	err := store.Put(t.Context(), ref, "password")
	require.NoError(t, err)

	job := &deletePgInstanceSecretJob{secrets: store, secretRef: ref}

	err = job.Do(t.Context())
	require.NoError(t, err)

	err = job.Do(t.Context())
	require.NoError(t, err)

	_, err = store.Get(t.Context(), ref)
	require.ErrorIs(t, err, user_errors.ErrSecretNotFound)
}
