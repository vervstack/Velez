package service_discovery

import (
	"github.com/docker/docker/client"
	makosh "github.com/godverv/makosh/pkg/makosh_be"
	"google.golang.org/grpc/metadata"

	"github.com/godverv/Velez/internal/config"
)

const (
	MakoshAuthHeader = "Makosh-Auth"
)

type ServiceDiscovery struct {
	authToken string
	cl        makosh.MakoshBeAPIClient
	md        metadata.MD

	dockerAPI    client.CommonAPIClient
	image        string
	portToExpose *string
}

func New(
	token string,
	cfg config.Config,
	cl makosh.MakoshBeAPIClient,
	dockerAPI client.CommonAPIClient,
) *ServiceDiscovery {

	env := cfg.GetEnvironment()

	var portToExpose *string
	if env.MakoshExposePort {
		portToExpose = env.MakoshPortToExpose
	}
	return &ServiceDiscovery{
		authToken: token,
		cl:        cl,
		md: metadata.New(map[string]string{
			MakoshAuthHeader: token,
		}),

		dockerAPI:    dockerAPI,
		image:        env.MakoshImage,
		portToExpose: portToExpose,
	}
}

func (s *ServiceDiscovery) GetToken() string {
	return s.authToken
}
