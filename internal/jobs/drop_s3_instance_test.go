package jobs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

const (
	testDropS3InstanceName = "my-s3"
)

func newDropS3InstancePayload() *velez_api.DropS3InstanceTaskPayload {
	return &velez_api.DropS3InstanceTaskPayload{Name: testDropS3InstanceName}
}

func Test_DropS3InstanceHandler_Action(t *testing.T) {
	h := NewDropS3InstanceHandler(nil, nil, nil, nil, nil, nil)

	require.Equal(t, DropS3InstanceAction, h.Action())
}

func Test_DropS3InstanceHandler_NewContextIsPayload(t *testing.T) {
	h := NewDropS3InstanceHandler(nil, nil, nil, nil, nil, nil)

	_, ok := h.NewContext().(*velez_api.DropS3InstanceTaskPayload)
	require.True(t, ok)
}

func Test_DropS3InstanceHandler_BuildJobsStepOrder(t *testing.T) {
	h := NewDropS3InstanceHandler(nil, nil, nil, nil, nil, nil)

	namedJobs := h.BuildJobs(newDropS3InstancePayload())

	names := make([]string, 0, len(namedJobs))
	for _, namedJob := range namedJobs {
		names = append(names, namedJob.Name)
	}

	want := []string{
		"unbind_bucket_owners",
		"remove_web_ui",
		"remove_service",
		"remove_docker_resources",
		"delete_secrets",
		"delete_config_entries",
	}

	require.Equal(t, want, names)
}

func Test_DropS3InstanceHandler_BuildJobsPanicsOnMismatchedContext(t *testing.T) {
	h := NewDropS3InstanceHandler(nil, nil, nil, nil, nil, nil)

	require.Panics(t, func() { h.BuildJobs(&velez_api.CreateS3InstanceTaskPayload{}) })
}

func Test_DeleteS3SecretsJob_RemovesOnlyInstanceSecretsAndIsRepeatable(t *testing.T) {
	ctx := context.Background()
	store := newStaticSecretsStore()

	own := []domain.SecretRef{
		domain.S3AdminTokenSecretRef(testDropS3InstanceName),
		domain.S3RpcSecretRef(testDropS3InstanceName),
		domain.S3WebUiPasswordSecretRef(testDropS3InstanceName),
	}
	other := domain.S3AdminTokenSecretRef("other-s3")

	for _, ref := range append(own, other) {
		require.NoError(t, store.Put(ctx, ref, "value"))
	}

	job := &deleteS3SecretsJob{secrets: store, name: testDropS3InstanceName}

	require.NoError(t, job.Do(ctx))
	require.NoError(t, job.Do(ctx))

	refs, err := store.ListRefs(ctx, other.Scope, testDropS3InstanceName)
	require.NoError(t, err)
	require.Empty(t, refs)

	_, err = store.Get(ctx, other)
	require.NoError(t, err)
}
