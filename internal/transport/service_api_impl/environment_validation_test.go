package service_api_impl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testSvcName = "svc"
)

// fakeEnvResolver is a hand-written service.VervServicesService fake exposing
// only the environment resolution method the validation path calls.
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

// An omitted environment is NOT rejected - it's forwarded verbatim to the
// service layer, which resolves it onto the default environment's suffix.
func TestCreateService_EmptyEnvironmentResolvesDefaultSuffix(t *testing.T) {
	resolver := &fakeEnvResolver{suffix: "prod-suffix"}
	impl := &Impl{servicesService: resolver}

	suffix, err := impl.resolveEnvironment(context.Background(), "")
	require.NoError(t, err)
	require.Equal(t, "prod-suffix", suffix)
	require.Empty(t, resolver.gotName)
}

// The full RPC path: an environment-less CreateDeploy must get past validation
// and reach the service layer.
func TestCreateDeploy_EmptyEnvironmentReachesService(t *testing.T) {
	resolver := &fakeDeployService{}
	impl := &Impl{servicesService: resolver}

	newSpec := &pb.CreateSmerd_Request{Name: testSvcName}
	req := &pb.CreateDeploy_Request{
		ServiceName:   testSvcName,
		Specification: &pb.CreateDeploy_Request_New{New: newSpec},
	}

	_, err := impl.CreateDeploy(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, testSvcName, resolver.gotDeploy.ServiceName)
	require.Empty(t, resolver.gotDeploy.GetEnvironment(),
		"an unset environment stays unset on the persisted spec - it's resolved to the default at launch time")
}

func TestCreateService_UnknownEnvironmentRejected(t *testing.T) {
	resolver := &fakeEnvResolver{err: rerrors.Wrap(user_errors.ErrEnvironmentNotFound)}
	impl := &Impl{servicesService: resolver}

	req := &pb.CreateService_Request{Name: testSvcName, Environment: "ghost"}

	_, err := impl.CreateService(context.Background(), req)
	require.ErrorIs(t, err, user_errors.ErrEnvironmentNotFound)
	require.Equal(t, "ghost", resolver.gotName)
}

func TestCreateDeploy_UnknownEnvironmentRejected(t *testing.T) {
	resolver := &fakeEnvResolver{err: rerrors.Wrap(user_errors.ErrEnvironmentNotFound)}
	impl := &Impl{servicesService: resolver}

	req := &pb.CreateDeploy_Request{ServiceName: testSvcName, Environment: "ghost"}

	_, err := impl.CreateDeploy(context.Background(), req)
	require.ErrorIs(t, err, user_errors.ErrEnvironmentNotFound)
}

// The deploy's environment has to end up on the persisted CreateSmerd spec -
// deploy_watcher.go only ever sees that stored request and re-resolves the
// suffix from its environment name.
func TestCreateDeploy_PropagatesEnvironmentIntoSmerdSpec(t *testing.T) {
	resolver := &fakeDeployService{}
	impl := &Impl{servicesService: resolver}

	newSpec := &pb.CreateSmerd_Request{Name: testSvcName}
	req := &pb.CreateDeploy_Request{
		ServiceName:   testSvcName,
		Environment:   "STAGE",
		Specification: &pb.CreateDeploy_Request_New{New: newSpec},
	}

	_, err := impl.CreateDeploy(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, "STAGE", newSpec.GetEnvironment())
	require.Equal(t, "STAGE", resolver.gotDeploy.GetEnvironment())
}

func TestResolveEnvironment_ValidEnvironmentResolvesSuffix(t *testing.T) {
	resolver := &fakeEnvResolver{suffix: "stg"}
	impl := &Impl{servicesService: resolver}

	suffix, err := impl.resolveEnvironment(context.Background(), "STAGE")
	require.NoError(t, err)
	require.Equal(t, "stg", suffix)
}

// fakeDeployService resolves any environment successfully and records the
// deploy request it was handed, so the environment-propagation assertion can
// look at what would actually be persisted.
type fakeDeployService struct {
	service.VervServicesService

	gotDeploy domain.CreateDeployReq
}

func (f *fakeDeployService) ResolveEnvironmentSuffix(_ context.Context, _ string) (string, error) {
	return "stg", nil
}

func (f *fakeDeployService) CreateNewDeploy(_ context.Context, req domain.CreateDeployReq) error {
	f.gotDeploy = req

	return nil
}
