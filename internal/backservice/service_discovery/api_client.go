package service_discovery

import (
	"github.com/godverv/makosh/pkg/makosh_be"
)

type ApiClient struct {
	cl makosh_be.MakoshBeAPIClient
}

func newApiClient(addr string, token string) *ApiClient {
	return &ApiClient{}
}
