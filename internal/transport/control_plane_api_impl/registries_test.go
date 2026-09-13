package control_plane_api_impl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func Test_ListRegistries_NeverExposesSecret(t *testing.T) {
	svc := &fakeVervServices{
		regListResp: []domain.Registry{
			{Id: 1, Name: "hub", Type: domain.RegistryTypeDockerHub, Secret: "super-secret"},
		},
	}
	impl := &Impl{vervServices: svc}

	resp, err := impl.ListRegistries(context.Background(), &pb.ListRegistries_Request{})
	require.NoError(t, err)
	require.Len(t, resp.GetRegistries(), 1)
	require.Equal(t, "hub", resp.GetRegistries()[0].GetName())
	require.Equal(t, pb.RegistryType_REGISTRY_TYPE_DOCKERHUB, resp.GetRegistries()[0].GetType())
}

func Test_ListRegistries_Error(t *testing.T) {
	impl := &Impl{vervServices: &fakeVervServices{regListErr: rerrors.New("boom")}}

	_, err := impl.ListRegistries(context.Background(), &pb.ListRegistries_Request{})
	require.Error(t, err)
}

func Test_CreateRegistry_ForwardsFields(t *testing.T) {
	svc := &fakeVervServices{
		regCreateResp: domain.Registry{Id: 2, Name: "generic", Type: domain.RegistryTypeGenericV2},
	}
	impl := &Impl{vervServices: svc}

	url := "https://registry.example.com"
	username := "bob"
	secret := "hunter2"
	isDefault := true

	req := &pb.CreateRegistry_Request{
		Name:      "generic",
		Type:      pb.RegistryType_REGISTRY_TYPE_GENERIC_V2,
		Url:       &url,
		Username:  &username,
		Secret:    &secret,
		IsDefault: &isDefault,
	}

	resp, err := impl.CreateRegistry(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "generic", svc.regCreateReq.Name)
	require.Equal(t, domain.RegistryTypeGenericV2, svc.regCreateReq.Type)
	require.Equal(t, url, svc.regCreateReq.Url)
	require.Equal(t, username, svc.regCreateReq.Username)
	require.Equal(t, secret, svc.regCreateReq.Secret)
	require.True(t, svc.regCreateReq.IsDefault)
	require.Equal(t, int64(2), resp.GetRegistry().GetId())
}

func Test_CreateRegistry_Error(t *testing.T) {
	impl := &Impl{vervServices: &fakeVervServices{regCreateErr: rerrors.New("duplicate")}}

	req := &pb.CreateRegistry_Request{Name: "dup", Type: pb.RegistryType_REGISTRY_TYPE_DOCKERHUB}

	_, err := impl.CreateRegistry(context.Background(), req)
	require.Error(t, err)
}

// An omitted field on UpdateRegistry must reach the service as nil, not a
// zeroed pointer - that's what makes "leave secret blank to keep it
// unchanged" work.
func Test_UpdateRegistry_OmittedFieldsStayNil(t *testing.T) {
	svc := &fakeVervServices{regUpdateResp: domain.Registry{Id: 5, Name: "renamed"}}
	impl := &Impl{vervServices: svc}

	name := "renamed"
	req := &pb.UpdateRegistry_Request{Id: 5, Name: &name}

	resp, err := impl.UpdateRegistry(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, int64(5), svc.regUpdateReq.Id)
	require.NotNil(t, svc.regUpdateReq.Name)
	require.Equal(t, "renamed", *svc.regUpdateReq.Name)
	require.Nil(t, svc.regUpdateReq.Secret, "omitted secret must stay nil, not be blanked out")
	require.Nil(t, svc.regUpdateReq.Type)
	require.Equal(t, "renamed", resp.GetRegistry().GetName())
}

func Test_UpdateRegistry_TypeForwarded(t *testing.T) {
	svc := &fakeVervServices{}
	impl := &Impl{vervServices: svc}

	regType := pb.RegistryType_REGISTRY_TYPE_GENERIC_V2
	req := &pb.UpdateRegistry_Request{Id: 5, Type: &regType}

	_, err := impl.UpdateRegistry(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, svc.regUpdateReq.Type)
	require.Equal(t, domain.RegistryTypeGenericV2, *svc.regUpdateReq.Type)
}

func Test_UpdateRegistry_Error(t *testing.T) {
	impl := &Impl{vervServices: &fakeVervServices{regUpdateErr: rerrors.New("nope")}}

	_, err := impl.UpdateRegistry(context.Background(), &pb.UpdateRegistry_Request{Id: 1})
	require.Error(t, err)
}

func Test_DeleteRegistry_ByName(t *testing.T) {
	svc := &fakeVervServices{}
	impl := &Impl{vervServices: svc}

	name := "hub"
	req := &pb.DeleteRegistry_Request{Name: &name}

	_, err := impl.DeleteRegistry(context.Background(), req)
	require.NoError(t, err)
	require.Nil(t, svc.regDeleteReq.Id)
	require.NotNil(t, svc.regDeleteReq.Name)
	require.Equal(t, "hub", *svc.regDeleteReq.Name)
}

func Test_DeleteRegistry_ById(t *testing.T) {
	svc := &fakeVervServices{}
	impl := &Impl{vervServices: svc}

	id := int64(9)
	req := &pb.DeleteRegistry_Request{Id: &id}

	_, err := impl.DeleteRegistry(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, svc.regDeleteReq.Id)
	require.Equal(t, int64(9), *svc.regDeleteReq.Id)
}

func Test_DeleteRegistry_Error(t *testing.T) {
	impl := &Impl{vervServices: &fakeVervServices{regDeleteErr: rerrors.New("not found")}}

	id := int64(1)

	_, err := impl.DeleteRegistry(context.Background(), &pb.DeleteRegistry_Request{Id: &id})
	require.Error(t, err)
}
