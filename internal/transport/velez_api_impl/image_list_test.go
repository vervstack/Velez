package velez_api_impl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testRegistryHost = "myreg.io"
	testRegistryUrl  = "https://" + testRegistryHost
)

// fakeRegistrySvc is a hand-written service.VervServicesService fake exposing
// only the three methods resolveSearchRegistry touches.
type fakeRegistrySvc struct {
	service.VervServicesService

	statefull  bool
	registries []domain.Registry
	byId       map[int64]domain.Registry
	getErr     error
	listErr    error
}

func (f *fakeRegistrySvc) IsStatefull() bool {
	return f.statefull
}

func (f *fakeRegistrySvc) ListRegistries(_ context.Context) ([]domain.Registry, error) {
	return f.registries, f.listErr
}

func (f *fakeRegistrySvc) GetRegistry(_ context.Context, id int64) (domain.Registry, error) {
	if f.getErr != nil {
		return domain.Registry{}, f.getErr
	}

	return f.byId[id], nil
}

func searchReq(name string, registryId int64) *velez_api.SearchImages_Request {
	req := &velez_api.SearchImages_Request{Name: name}
	if registryId != 0 {
		req.RegistryId = &registryId
	}

	return req
}

func TestRegistryDomainFromRef(t *testing.T) {
	cases := []struct {
		name       string
		ref        string
		wantDomain string
		wantTerm   string
	}{
		{"no domain", "nginx", "", ""},
		{"localhost with port", "localhost:5000/x", "localhost:5000", "x"},
		{"ghcr", "ghcr.io/o/i", "ghcr.io", "o/i"},
		{"explicit docker hub", "docker.io/library/x", "", ""},
		{"bare official name", "postgres", "", ""},
		{"user namespace, no domain", "team/api", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotDomain, gotTerm := registryDomainFromRef(tc.ref)
			require.Equal(t, tc.wantDomain, gotDomain)
			require.Equal(t, tc.wantTerm, gotTerm)
		})
	}
}

func TestRegistryUrlHost(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"with scheme", testRegistryUrl, testRegistryHost},
		{"no scheme", testRegistryHost, testRegistryHost},
		{"no scheme with port", testRegistryHost + ":5000", testRegistryHost + ":5000"},
		{"empty", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, registryUrlHost(tc.raw))
		})
	}
}

func TestResolveSearchRegistry_RegistryIdRequiresStatefullMode(t *testing.T) {
	svc := &fakeRegistrySvc{statefull: false}
	impl := &Impl{vervServices: svc}

	_, _, err := impl.resolveSearchRegistry(context.Background(), searchReq("postgres", 7))
	require.ErrorIs(t, err, user_errors.ErrRequiresStatefullMode)
}

func TestResolveSearchRegistry_RegistryIdInStatefullMode(t *testing.T) {
	want := domain.Registry{Id: 7, Type: domain.RegistryTypeGenericV2, Url: testRegistryUrl}
	svc := &fakeRegistrySvc{statefull: true, byId: map[int64]domain.Registry{7: want}}
	impl := &Impl{vervServices: svc}

	reg, term, err := impl.resolveSearchRegistry(context.Background(), searchReq("myreg.io/team/api", 7))
	require.NoError(t, err)
	require.Equal(t, want, reg)
	require.Equal(t, "myreg.io/team/api", term, "an explicit id keeps the term verbatim")
}

func TestResolveSearchRegistry_DomainMatchesConfiguredRegistry(t *testing.T) {
	configured := domain.Registry{Id: 3, Type: domain.RegistryTypeGenericV2, Url: testRegistryUrl, Username: "u"}
	svc := &fakeRegistrySvc{registries: []domain.Registry{configured}}
	impl := &Impl{vervServices: svc}

	reg, term, err := impl.resolveSearchRegistry(context.Background(), searchReq("myreg.io/team/api", 0))
	require.NoError(t, err)
	require.Equal(t, configured, reg)
	require.Equal(t, "team/api", term)
}

func TestResolveSearchRegistry_DomainWithoutMatchIsAdHoc(t *testing.T) {
	svc := &fakeRegistrySvc{}
	impl := &Impl{vervServices: svc}

	reg, term, err := impl.resolveSearchRegistry(context.Background(), searchReq("ghcr.io/o/i", 0))
	require.NoError(t, err)
	require.Equal(t, domain.RegistryTypeGenericV2, reg.Type)
	require.Equal(t, "https://ghcr.io", reg.Url)
	require.Empty(t, reg.Username)
	require.Equal(t, "o/i", term)
}

func TestResolveSearchRegistry_DomainPathWorksInSingleNodeMode(t *testing.T) {
	svc := &fakeRegistrySvc{statefull: false}
	impl := &Impl{vervServices: svc}

	reg, term, err := impl.resolveSearchRegistry(context.Background(), searchReq("ghcr.io/o/i", 0))
	require.NoError(t, err)
	require.Equal(t, "https://ghcr.io", reg.Url)
	require.Equal(t, "o/i", term)
}

func TestResolveSearchRegistry_DefaultRow(t *testing.T) {
	def := domain.Registry{Id: 1, Type: domain.RegistryTypeGenericV2, Url: "https://default.io", IsDefault: true}
	svc := &fakeRegistrySvc{registries: []domain.Registry{{Id: 2}, def}}
	impl := &Impl{vervServices: svc}

	reg, term, err := impl.resolveSearchRegistry(context.Background(), searchReq("postgres", 0))
	require.NoError(t, err)
	require.Equal(t, def, reg)
	require.Equal(t, "postgres", term)
}

func TestResolveSearchRegistry_FallsBackToDockerHub(t *testing.T) {
	svc := &fakeRegistrySvc{}
	impl := &Impl{vervServices: svc}

	reg, term, err := impl.resolveSearchRegistry(context.Background(), searchReq("docker.io/library/postgres", 0))
	require.NoError(t, err)
	require.Equal(t, domain.RegistryTypeDockerHub, reg.Type)
	require.Equal(t, "docker.io/library/postgres", term)
}

func TestSearchImages_RegistryIdRejectedInSingleNodeMode(t *testing.T) {
	svc := &fakeRegistrySvc{statefull: false}
	impl := &Impl{vervServices: svc}

	_, err := impl.SearchImages(context.Background(), searchReq("postgres", 7))
	require.ErrorIs(t, err, user_errors.ErrRequiresStatefullMode)
}

// GetRegistry errors must propagate rather than being swallowed into a
// dockerhub fallback.
func TestResolveSearchRegistry_GetRegistryErrorPropagates(t *testing.T) {
	sentinel := rerrors.New("boom")
	svc := &fakeRegistrySvc{statefull: true, getErr: sentinel}
	impl := &Impl{vervServices: svc}

	_, _, err := impl.resolveSearchRegistry(context.Background(), searchReq("postgres", 7))
	require.ErrorIs(t, err, sentinel)
}
