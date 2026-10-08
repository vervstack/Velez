package jobs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/storage"
)

const (
	testDropDindName = "my-dind"
)

type dropTestDindInstances struct {
	storage.DindInstancesStorage

	deletedIds []int64
}

func (d *dropTestDindInstances) DeleteDindInstance(_ context.Context, serviceId int64) error {
	d.deletedIds = append(d.deletedIds, serviceId)

	return nil
}

func newDropDindPayload() *velez_api.DropDindTaskPayload {
	return &velez_api.DropDindTaskPayload{Name: testDropDindName}
}

func Test_DropDindHandler_Action(t *testing.T) {
	h := NewDropDindHandler(nil, nil, nil)

	require.Equal(t, DropDindAction, h.Action())
}

func Test_DropDindHandler_NewContextIsPayload(t *testing.T) {
	h := NewDropDindHandler(nil, nil, nil)

	_, ok := h.NewContext().(*velez_api.DropDindTaskPayload)
	require.True(t, ok)
}

func Test_DropDindHandler_BuildJobsNamesAndOrder(t *testing.T) {
	stg := storage.NewStorageContainer(&fakeClusterStorage{})
	h := NewDropDindHandler(stg, nil, nil)

	namedJobs := h.BuildJobs(newDropDindPayload())

	names := make([]string, 0, len(namedJobs))
	for _, namedJob := range namedJobs {
		names = append(names, namedJob.Name)
	}

	require.Equal(t,
		[]string{stepDeleteDindRow, stepRemoveDindService, stepRemoveDindNetwork, stepRemoveDindVolume},
		names,
	)
}

func Test_DeleteDindRowJob_RerunsWithoutError(t *testing.T) {
	services := &dropTestServices{isFound: true}

	services.svc.ID = 3

	instances := &dropTestDindInstances{}
	job := &deleteDindRowJob{services: services, dindInstances: instances, req: newDropDindPayload()}

	err := job.Do(t.Context())
	require.NoError(t, err)

	err = job.Do(t.Context())
	require.NoError(t, err)

	require.Equal(t, []int64{3, 3}, instances.deletedIds)
}

func Test_DeleteDindRowJob_IsNoopWhenServiceAlreadyGone(t *testing.T) {
	instances := &dropTestDindInstances{}
	job := &deleteDindRowJob{services: &dropTestServices{}, dindInstances: instances, req: newDropDindPayload()}

	err := job.Do(t.Context())
	require.NoError(t, err)

	require.Empty(t, instances.deletedIds)
}

func Test_RemoveDindServiceJob_RemovesServiceWithRunningInstances(t *testing.T) {
	verv := &dropTestVervServices{}
	job := &removeDindServiceJob{vervServices: verv, req: newDropDindPayload()}

	err := job.Do(t.Context())
	require.NoError(t, err)

	require.Len(t, verv.removed, 1)
	require.Equal(t, testDropDindName, verv.removed[0].Name)
	require.True(t, verv.removed[0].DropRunningInstances)
}
