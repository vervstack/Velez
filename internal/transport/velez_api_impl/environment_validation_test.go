package velez_api_impl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/service/service_manager/verv_services"
)

const (
	testSvcName  = "svc"
	testGhostEnv = "ghost"
)

// fakeEnvResolver is a hand-written service.VervServicesService fake (no
// mocking library is used in this repo) exposing only the environment
// resolution method the validation path calls.
type fakeEnvResolver struct {
	service.VervServicesService

	gotName string
	suffix  string
	err     error
}

func (f *fakeEnvResolver) ResolveEnvironmentSuffix(_ context.Context, name string) (string, error) {
	f.gotName = name

	return f.suffix, f.err
}

// failingJobsEngine makes any test that gets *past* validation fail loudly, so
// "validation rejected the request" can't be confused with "the RPC happened
// to succeed".
type panickingSmerdService struct {
	service.ContainerService
}

// recordingSmerdService stands in for the container service on paths that are
// expected to get PAST validation.
type recordingSmerdService struct {
	service.ContainerService

	called bool
}

func (r *recordingSmerdService) ListSmerds(
	_ context.Context, _ *velez_api.ListSmerds_Request,
) (*velez_api.ListSmerds_Response, error) {
	r.called = true

	return &velez_api.ListSmerds_Response{}, nil
}

// An omitted environment is NOT rejected: it's forwarded verbatim to the
// service layer, which resolves it to the default environment's suffix. The
// transport layer deliberately does no defaulting of its own - there's exactly
// one place that knows what "default" means.
func TestCreateSmerd_EmptyEnvironmentResolvesDefaultSuffix(t *testing.T) {
	resolver := &fakeEnvResolver{suffix: "prod-suffix"}
	impl := &Impl{vervServices: resolver}

	suffix, err := impl.resolveEnvironment(context.Background(), "")
	require.NoError(t, err)
	require.Equal(t, "prod-suffix", suffix)
	require.Empty(t, resolver.gotName)
}

// The full RPC path: an environment-less ListSmerds must reach the container
// service instead of failing validation.
func TestListSmerds_EmptyEnvironmentReachesService(t *testing.T) {
	resolver := &fakeEnvResolver{suffix: "prod-suffix"}
	lister := &recordingSmerdService{}
	impl := &Impl{vervServices: resolver, smerdService: lister}

	req := &velez_api.ListSmerds_Request{}

	_, err := impl.ListSmerds(context.Background(), req)
	require.NoError(t, err)
	require.True(t, lister.called, "ListSmerds must not be rejected for an empty environment")
	require.Empty(t, resolver.gotName)
}

func TestCreateSmerd_UnknownEnvironmentRejected(t *testing.T) {
	resolver := &fakeEnvResolver{err: rerrors.Wrap(verv_services.ErrEnvironmentNotFound)}
	impl := &Impl{vervServices: resolver}

	req := &velez_api.CreateSmerd_Request{Name: testSvcName, Environment: testGhostEnv}

	_, err := impl.CreateSmerd(context.Background(), req)
	require.ErrorIs(t, err, verv_services.ErrEnvironmentNotFound)
	require.Equal(t, testGhostEnv, resolver.gotName)
}

func TestListSmerds_UnknownEnvironmentRejected(t *testing.T) {
	resolver := &fakeEnvResolver{err: rerrors.Wrap(verv_services.ErrEnvironmentNotFound)}
	impl := &Impl{vervServices: resolver, smerdService: &panickingSmerdService{}}

	req := &velez_api.ListSmerds_Request{Environment: testGhostEnv}

	_, err := impl.ListSmerds(context.Background(), req)
	require.ErrorIs(t, err, verv_services.ErrEnvironmentNotFound)
}

func TestDropSmerd_UnknownEnvironmentRejected(t *testing.T) {
	resolver := &fakeEnvResolver{err: rerrors.Wrap(verv_services.ErrEnvironmentNotFound)}
	impl := &Impl{vervServices: resolver}

	req := &velez_api.DropSmerd_Request{Name: []string{testSvcName}, Environment: testGhostEnv}

	_, err := impl.DropSmerd(context.Background(), req)
	require.ErrorIs(t, err, verv_services.ErrEnvironmentNotFound)
}

func TestUpgradeSmerd_UnknownEnvironmentRejected(t *testing.T) {
	resolver := &fakeEnvResolver{err: rerrors.Wrap(verv_services.ErrEnvironmentNotFound)}
	impl := &Impl{vervServices: resolver}

	req := &velez_api.UpgradeSmerd_Request{Name: testSvcName, Environment: testGhostEnv}

	_, err := impl.UpgradeSmerd(context.Background(), req)
	require.ErrorIs(t, err, verv_services.ErrEnvironmentNotFound)
}

// A valid environment must pass validation through to the resolver unchanged.
func TestResolveEnvironment_ValidEnvironmentResolvesSuffix(t *testing.T) {
	resolver := &fakeEnvResolver{suffix: "stg"}
	impl := &Impl{vervServices: resolver}

	suffix, err := impl.resolveEnvironment(context.Background(), "STAGE")
	require.NoError(t, err)
	require.Equal(t, "stg", suffix)
	require.Equal(t, "STAGE", resolver.gotName)
}
