package jobs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/registries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testDropRegistryName = "cr_my-registry"
	testDropEnvironment  = "prod"
)

type dropTestServices struct {
	storage.ServicesStorage

	svc     domain.Service
	isFound bool
}

func (s *dropTestServices) GetByName(_ context.Context, _ string) (domain.Service, error) {
	if !s.isFound {
		return domain.Service{}, rerrors.Wrap(user_errors.ErrStorageNotFound)
	}

	return s.svc, nil
}

type dropTestVervServices struct {
	service.VervServicesService

	removed []domain.RemoveServiceReq
}

func (v *dropTestVervServices) Remove(_ context.Context, req domain.RemoveServiceReq) error {
	v.removed = append(v.removed, req)

	return nil
}

type dropTestRegistryInstances struct {
	storage.RegistryInstancesStorage

	deletedIds []int64
}

func (r *dropTestRegistryInstances) DeleteRegistryInstance(_ context.Context, serviceId int64) error {
	r.deletedIds = append(r.deletedIds, serviceId)

	return nil
}

func newDropRegistryPayload() *velez_api.DropRegistryInstanceTaskPayload {
	return &velez_api.DropRegistryInstanceTaskPayload{Name: testDropRegistryName}
}

func Test_DropRegistryInstanceHandler_Action(t *testing.T) {
	h := NewDropRegistryInstanceHandler(nil, nil, nil, nil, nil)

	require.Equal(t, DropRegistryInstanceAction, h.Action())
}

func Test_DropRegistryInstanceHandler_NewContextIsPayload(t *testing.T) {
	h := NewDropRegistryInstanceHandler(nil, nil, nil, nil, nil)

	_, ok := h.NewContext().(*velez_api.DropRegistryInstanceTaskPayload)
	require.True(t, ok)
}

func Test_DropRegistryInstanceHandler_BuildJobsNamesAndOrder(t *testing.T) {
	stg := storage.NewStorageContainer(&fakeClusterStorage{})
	h := NewDropRegistryInstanceHandler(stg, nil, nil, nil, nil)

	namedJobs := h.BuildJobs(newDropRegistryPayload())

	names := make([]string, 0, len(namedJobs))
	for _, namedJob := range namedJobs {
		names = append(names, namedJob.Name)
	}

	require.Equal(t,
		[]string{
			stepRemoveRegistryS3Access, stepDeleteRegistryInstanceRow, stepRemoveRegistryUiService,
			stepRemoveRegistryService, stepDeleteBuiltinRegistryRow, stepDeleteRegistrySecret,
		},
		names,
	)
}

func Test_RemoveRegistryServiceJob_RemovesUiSidecarInServiceEnvironment(t *testing.T) {
	services := &dropTestServices{isFound: true}

	services.svc.Env = testDropEnvironment

	verv := &dropTestVervServices{}
	job := &removeRegistryServiceJob{
		services:     services,
		vervServices: verv,
		req:          newDropRegistryPayload(),
		isUiSidecar:  true,
	}

	err := job.Do(t.Context())
	require.NoError(t, err)

	require.Len(t, verv.removed, 1)
	require.Equal(t, registryaasUiServiceName(testDropRegistryName), verv.removed[0].Name)
	require.Equal(t, testDropEnvironment, verv.removed[0].Environment)
	require.True(t, verv.removed[0].DropRunningInstances)
}

func Test_RemoveRegistryServiceJob_IsNoopWhenServiceAlreadyGone(t *testing.T) {
	verv := &dropTestVervServices{}
	job := &removeRegistryServiceJob{
		services:     &dropTestServices{},
		vervServices: verv,
		req:          newDropRegistryPayload(),
	}

	err := job.Do(t.Context())
	require.NoError(t, err)

	err = job.Do(t.Context())
	require.NoError(t, err)

	require.Empty(t, verv.removed)
}

func Test_DeleteRegistryInstanceRowJob_RerunsWithoutError(t *testing.T) {
	services := &dropTestServices{isFound: true}

	services.svc.ID = 7

	instances := &dropTestRegistryInstances{}
	job := &deleteRegistryInstanceRowJob{services: services, registryInstances: instances, req: newDropRegistryPayload()}

	err := job.Do(t.Context())
	require.NoError(t, err)

	err = job.Do(t.Context())
	require.NoError(t, err)

	require.Equal(t, []int64{7, 7}, instances.deletedIds)
}

func Test_DeleteRegistryInstanceRowJob_IsNoopWhenServiceAlreadyGone(t *testing.T) {
	instances := &dropTestRegistryInstances{}
	job := &deleteRegistryInstanceRowJob{
		services:          &dropTestServices{},
		registryInstances: instances,
		req:               newDropRegistryPayload(),
	}

	err := job.Do(t.Context())
	require.NoError(t, err)

	require.Empty(t, instances.deletedIds)
}

func Test_RemoveRegistryS3AccessJob_IsNoopWhenServiceAlreadyGone(t *testing.T) {
	job := &removeRegistryS3AccessJob{services: &dropTestServices{}, req: newDropRegistryPayload()}

	err := job.Do(t.Context())
	require.NoError(t, err)
}

func Test_DeleteBuiltinRegistryRowJob_DeletesRowAndRerunsWithoutError(t *testing.T) {
	stg := registries.NewStatic()
	upserter, ok := stg.(registries.BuiltinRegistryUpserter)
	require.True(t, ok)

	upsertReq := domain.CreateRegistryReq{Name: testDropRegistryName, Type: domain.RegistryTypeGenericV2}

	_, err := upserter.UpsertBuiltinRegistry(t.Context(), upsertReq)
	require.NoError(t, err)

	job := &deleteBuiltinRegistryRowJob{registries: stg, req: newDropRegistryPayload()}

	err = job.Do(t.Context())
	require.NoError(t, err)

	err = job.Do(t.Context())
	require.NoError(t, err)

	all, err := stg.ListRegistries(t.Context())
	require.NoError(t, err)

	for _, registry := range all {
		require.NotEqual(t, testDropRegistryName, registry.Name)
	}
}

func Test_DeleteRegistryPasswordSecretJob_DeletesSecretAndRerunsWithoutError(t *testing.T) {
	store := newStaticSecretsStore()
	ref := registryInstanceSecretRef(testDropRegistryName)

	err := store.Put(t.Context(), ref, "password")
	require.NoError(t, err)

	job := &deleteRegistryPasswordSecretJob{secrets: store, req: newDropRegistryPayload()}

	err = job.Do(t.Context())
	require.NoError(t, err)

	_, err = store.Get(t.Context(), ref)
	require.ErrorIs(t, err, user_errors.ErrSecretNotFound)

	err = job.Do(t.Context())
	require.NoError(t, err)
}
