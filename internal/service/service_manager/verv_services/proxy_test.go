package verv_services

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/sqlc-dev/pqtype"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/proxyenv"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/deployments_queries"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testProxyServiceName = "gitlab_runner_Artel"
	testProxyServiceUrl  = "socks5://192.168.1.44:1080"
	testProxyOtherKey    = "FOO"
	testProxyOtherValue  = "bar"
)

// specRecordingDeployments keeps the specification payloads UpgradeDeploy writes.
type specRecordingDeployments struct {
	*testDeploymentsStorage

	payloads [][]byte
}

func (r *specRecordingDeployments) CreateSpecification(
	_ context.Context, params deployments_queries.CreateSpecificationParams,
) (int64, error) {
	r.payloads = append(r.payloads, params.VervPayload.RawMessage)

	return 0, nil
}

func (r *specRecordingDeployments) WithTx(_ *sql.Tx) deployments_queries.Querier {
	return r
}

func newProxyService(t *testing.T, specEnv map[string]string) (*VervService, *specRecordingDeployments) {
	t.Helper()

	spec := &velez_api.CreateSmerd_Request{
		Name:      testProxyServiceName,
		ImageName: "gitlab/gitlab-runner:latest",
		Env:       specEnv,
	}

	payload, err := json.Marshal(spec)
	require.NoError(t, err)

	recorder := &specRecordingDeployments{testDeploymentsStorage: &testDeploymentsStorage{
		listFunc: func(_ context.Context, _ domain.ListDeploymentsReq) ([]domain.Deployment, error) {
			return []domain.Deployment{
				{Id: 1, SpecId: 1, NodeId: 1, Status: deployments_queries.VelezDeploymentStatusRUNNING},
			}, nil
		},
		getSpecificationByIdFunc: func(_ context.Context, _ int64) (deployments_queries.GetSpecificationByIdRow, error) {
			return deployments_queries.GetSpecificationByIdRow{
				VervPayload: pqtype.NullRawMessage{RawMessage: payload, Valid: true},
			}, nil
		},
	}}

	service := &VervService{
		dataStorage: &testStorage{
			services:    &testServicesStorageWithGetByName{},
			deployments: recorder,
			txManager:   fakeTransactor{},
		},
	}

	return service, recorder
}

func recordedSpec(t *testing.T, recorder *specRecordingDeployments) *velez_api.CreateSmerd_Request {
	t.Helper()

	require.Len(t, recorder.payloads, 1)

	spec := &velez_api.CreateSmerd_Request{}

	err := json.Unmarshal(recorder.payloads[0], spec)
	require.NoError(t, err)

	return spec
}

func Test_SetServiceProxy_SetsProxyEnvOnTheSameSpec(t *testing.T) {
	service, recorder := newProxyService(t, map[string]string{testProxyOtherKey: testProxyOtherValue})

	req := domain.SetServiceProxyReq{
		ServiceName:      testProxyServiceName,
		ProxyUrl:         testProxyServiceUrl,
		ProxyBypassHosts: []string{"gitlab.internal"},
	}

	err := service.SetServiceProxy(context.Background(), req)
	require.NoError(t, err)

	spec := recordedSpec(t, recorder)

	require.Equal(t, "gitlab/gitlab-runner:latest", spec.GetImageName())
	require.Equal(t, testProxyOtherValue, spec.GetEnv()[testProxyOtherKey])

	proxyUrl, bypassHosts := proxyenv.Parse(spec.GetEnv())

	require.Equal(t, testProxyServiceUrl, proxyUrl)
	require.Equal(t, []string{"gitlab.internal"}, bypassHosts)
	require.Equal(t, testProxyServiceUrl, spec.GetEnv()["http_proxy"])
}

func Test_SetServiceProxy_EmptyProxyUrl_RemovesProxyEnv(t *testing.T) {
	env := proxyenv.Env(testProxyServiceUrl, nil)

	env[testProxyOtherKey] = testProxyOtherValue

	service, recorder := newProxyService(t, env)

	err := service.SetServiceProxy(context.Background(), domain.SetServiceProxyReq{ServiceName: testProxyServiceName})
	require.NoError(t, err)

	spec := recordedSpec(t, recorder)

	require.Equal(t, map[string]string{testProxyOtherKey: testProxyOtherValue}, spec.GetEnv())
}

func Test_SetServiceProxy_InvalidRequest_ReturnsError(t *testing.T) {
	cases := []struct {
		name    string
		req     domain.SetServiceProxyReq
		wantErr error
	}{
		{
			"no service name",
			domain.SetServiceProxyReq{ProxyUrl: testProxyServiceUrl},
			user_errors.ErrServiceNameRequiredToFind,
		},
		{
			"proxy url without scheme",
			domain.SetServiceProxyReq{ServiceName: testProxyServiceName, ProxyUrl: "192.168.1.44:1080"},
			user_errors.ErrServiceProxyUrlInvalid,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			service, recorder := newProxyService(t, nil)

			err := service.SetServiceProxy(context.Background(), tc.req)

			require.ErrorIs(t, err, tc.wantErr)
			require.Empty(t, recorder.payloads)
		})
	}
}

func Test_SetServiceProxy_NoRunningDeployment_ReturnsError(t *testing.T) {
	service, recorder := newProxyService(t, nil)

	recorder.listFunc = func(_ context.Context, _ domain.ListDeploymentsReq) ([]domain.Deployment, error) {
		return nil, nil
	}

	req := domain.SetServiceProxyReq{ServiceName: testProxyServiceName, ProxyUrl: testProxyServiceUrl}

	err := service.SetServiceProxy(context.Background(), req)

	require.ErrorIs(t, err, user_errors.ErrNoRunningDeployment)
}
