package control_plane_api_impl

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service"
)

const (
	testEnvStage     = "STAGE"
	testEnvSuffixStg = "stg"
)

// fakeVervServices is a hand-written service.VervServicesService fake (no
// mocking library is used in this package - see enable_plugin_test.go). Only
// the environment methods are exercised; the rest satisfy the interface.
type fakeVervServices struct {
	service.VervServicesService

	listResp []domain.Environment
	listErr  error

	createReq  domain.CreateEnvironmentReq
	createResp domain.Environment
	createErr  error

	updateReq  domain.UpdateEnvironmentReq
	updateResp domain.Environment
	updateErr  error

	deleteReq domain.DeleteEnvironmentReq
	deleteErr error

	regListResp []domain.Registry
	regListErr  error

	regCreateReq  domain.CreateRegistryReq
	regCreateResp domain.Registry
	regCreateErr  error

	regUpdateReq  domain.UpdateRegistryReq
	regUpdateResp domain.Registry
	regUpdateErr  error

	regDeleteReq domain.DeleteRegistryReq
	regDeleteErr error
}

func (f *fakeVervServices) ListRegistries(_ context.Context) ([]domain.Registry, error) {
	return f.regListResp, f.regListErr
}

func (f *fakeVervServices) CreateRegistry(
	_ context.Context, req domain.CreateRegistryReq,
) (domain.Registry, error) {
	f.regCreateReq = req

	return f.regCreateResp, f.regCreateErr
}

func (f *fakeVervServices) UpdateRegistry(
	_ context.Context, req domain.UpdateRegistryReq,
) (domain.Registry, error) {
	f.regUpdateReq = req

	return f.regUpdateResp, f.regUpdateErr
}

func (f *fakeVervServices) DeleteRegistry(_ context.Context, req domain.DeleteRegistryReq) error {
	f.regDeleteReq = req

	return f.regDeleteErr
}

func (f *fakeVervServices) ListEnvironments(_ context.Context) ([]domain.Environment, error) {
	return f.listResp, f.listErr
}

func (f *fakeVervServices) CreateEnvironment(
	_ context.Context, req domain.CreateEnvironmentReq,
) (domain.Environment, error) {
	f.createReq = req

	return f.createResp, f.createErr
}

func (f *fakeVervServices) UpdateEnvironment(
	_ context.Context, req domain.UpdateEnvironmentReq,
) (domain.Environment, error) {
	f.updateReq = req

	return f.updateResp, f.updateErr
}

func (f *fakeVervServices) DeleteEnvironment(_ context.Context, req domain.DeleteEnvironmentReq) error {
	f.deleteReq = req

	return f.deleteErr
}

// ListEnvironments used to wrap bare names into &pb.Environment{Name: name}
// with no id/suffix/timestamps just to keep the build green. It must now
// surface the full DB-backed row.
func Test_ListEnvironments_ReturnsFullEnvironments(t *testing.T) {
	created := time.Date(2026, 8, 2, 10, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 8, 2, 11, 0, 0, 0, time.UTC)

	svc := &fakeVervServices{
		listResp: []domain.Environment{
			{ID: 1, Name: "PROD", Suffix: "", CreatedAt: created, UpdatedAt: updated},
			{ID: 2, Name: testEnvStage, Suffix: testEnvSuffixStg, CreatedAt: created, UpdatedAt: updated},
		},
	}

	impl := &Impl{vervServices: svc}

	resp, err := impl.ListEnvironments(context.Background(), &pb.ListEnvironments_Request{})
	require.NoError(t, err)
	require.Len(t, resp.GetEnvironments(), 2)

	first := resp.GetEnvironments()[0]
	require.Equal(t, int64(1), first.GetId())
	require.Equal(t, "PROD", first.GetName())
	require.Equal(t, created, first.GetCreatedAt().AsTime())
	require.Equal(t, updated, first.GetUpdatedAt().AsTime())

	require.Equal(t, testEnvSuffixStg, resp.GetEnvironments()[1].GetSuffix())
}

func Test_ListEnvironments_Error(t *testing.T) {
	impl := &Impl{vervServices: &fakeVervServices{listErr: rerrors.New("boom")}}

	_, err := impl.ListEnvironments(context.Background(), &pb.ListEnvironments_Request{})
	require.Error(t, err)
}

func Test_CreateEnvironment_PassesSuffixThrough(t *testing.T) {
	svc := &fakeVervServices{
		createResp: domain.Environment{ID: 3, Name: testEnvStage, Suffix: testEnvSuffixStg},
	}
	impl := &Impl{vervServices: svc}

	suffix := testEnvSuffixStg
	req := &pb.CreateEnvironment_Request{Name: testEnvStage, Suffix: &suffix}

	resp, err := impl.CreateEnvironment(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, testEnvStage, svc.createReq.Name)
	require.Equal(t, testEnvSuffixStg, svc.createReq.Suffix)
	require.Equal(t, int64(3), resp.GetEnvironment().GetId())
}

// An omitted suffix reaches the service as "" - the service layer is what
// defaults it to the name, so every caller gets the same rule.
func Test_CreateEnvironment_OmittedSuffixIsEmptyAtTransport(t *testing.T) {
	svc := &fakeVervServices{createResp: domain.Environment{ID: 4, Name: testEnvStage, Suffix: testEnvStage}}
	impl := &Impl{vervServices: svc}

	req := &pb.CreateEnvironment_Request{Name: testEnvStage}

	resp, err := impl.CreateEnvironment(context.Background(), req)
	require.NoError(t, err)
	require.Empty(t, svc.createReq.Suffix)
	require.Equal(t, testEnvStage, resp.GetEnvironment().GetSuffix())
}

func Test_CreateEnvironment_Error(t *testing.T) {
	impl := &Impl{vervServices: &fakeVervServices{createErr: rerrors.New("duplicate")}}

	_, err := impl.CreateEnvironment(context.Background(), &pb.CreateEnvironment_Request{Name: testEnvStage})
	require.Error(t, err)
}

func Test_UpdateEnvironment_ForwardsOptionalFields(t *testing.T) {
	svc := &fakeVervServices{updateResp: domain.Environment{ID: 5, Name: "STAGING", Suffix: testEnvSuffixStg}}
	impl := &Impl{vervServices: svc}

	name := "STAGING"
	req := &pb.UpdateEnvironment_Request{Id: 5, Name: &name}

	resp, err := impl.UpdateEnvironment(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, int64(5), svc.updateReq.ID)
	require.NotNil(t, svc.updateReq.Name)
	require.Equal(t, "STAGING", *svc.updateReq.Name)
	require.Nil(t, svc.updateReq.Suffix, "omitted suffix must stay nil, not be blanked out")
	require.Equal(t, "STAGING", resp.GetEnvironment().GetName())
}

func Test_UpdateEnvironment_Error(t *testing.T) {
	impl := &Impl{vervServices: &fakeVervServices{updateErr: rerrors.New("nope")}}

	_, err := impl.UpdateEnvironment(context.Background(), &pb.UpdateEnvironment_Request{Id: 1})
	require.Error(t, err)
}

func Test_DeleteEnvironment_ByName(t *testing.T) {
	svc := &fakeVervServices{}
	impl := &Impl{vervServices: svc}

	name := testEnvStage
	req := &pb.DeleteEnvironment_Request{Name: &name}

	_, err := impl.DeleteEnvironment(context.Background(), req)
	require.NoError(t, err)
	require.Nil(t, svc.deleteReq.ID)
	require.NotNil(t, svc.deleteReq.Name)
	require.Equal(t, testEnvStage, *svc.deleteReq.Name)
}

func Test_DeleteEnvironment_ById(t *testing.T) {
	svc := &fakeVervServices{}
	impl := &Impl{vervServices: svc}

	id := int64(7)
	req := &pb.DeleteEnvironment_Request{Id: &id}

	_, err := impl.DeleteEnvironment(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, svc.deleteReq.ID)
	require.Equal(t, int64(7), *svc.deleteReq.ID)
}

// A failed cascade must fail the RPC - the DB row is only dropped once every
// Docker resource is gone.
func Test_DeleteEnvironment_CascadeFailureFailsRPC(t *testing.T) {
	impl := &Impl{vervServices: &fakeVervServices{deleteErr: rerrors.New("container stuck")}}

	name := testEnvStage

	_, err := impl.DeleteEnvironment(context.Background(), &pb.DeleteEnvironment_Request{Name: &name})
	require.Error(t, err)
}
