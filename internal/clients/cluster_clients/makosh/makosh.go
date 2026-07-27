package makosh

import (
	"os"

	errors "go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"
	"go.vervstack.ru/makosh/pkg/makosh_be"
	vervResolver "go.vervstack.ru/makosh/pkg/resolver"
	"go.vervstack.ru/makosh/pkg/resolver/makosh_resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/middleware"
)

const (
	AuthHeader = "Makosh-Auth"

	ServiceName = "makosh"
)

func NewClient(token string, opts ...grpc.DialOption) (makosh_be.MakoshBeAPIClient, error) {
	opts = append(opts,
		grpc.WithUnaryInterceptor(middleware.HeaderOutgoingInterceptor(AuthHeader, token)))

	dial, err := grpc.NewClient("verv://"+ServiceName, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "error dialing")
	}

	return makosh_be.NewMakoshBeAPIClient(dial), nil
}

type ServiceDiscovery struct {
	makosh_be.MakoshBeAPIClient

	Sd *vervResolver.ServiceDiscovery
}

func NewServiceDiscovery(cfg config.Config) (*ServiceDiscovery, error) {
	url := toolbox.Coalesce(cfg.Overrides.MakoshUrl, cfg.Environment.MakoshURL)
	token := toolbox.Coalesce(cfg.Overrides.MakoshToken, cfg.Environment.MakoshKey)

	sd := &ServiceDiscovery{}

	var err error

	_ = os.Setenv(makosh_resolver.MakoshURL, url)
	_ = os.Setenv(makosh_resolver.MakoshSecret, token)

	sd.Sd, err = vervResolver.Init()
	if err != nil {
		return sd, errors.Wrap(err, "error initializing verv-service-discovery")
	}

	sd.Sd.SetOverrides(cfg.Overrides.Overrides)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	sd.MakoshBeAPIClient, err = NewClient(token, opts...)
	if err != nil {
		return sd, errors.Wrap(err, "error creating makosh api client")
	}

	return sd, nil
}

func (s *ServiceDiscovery) GetAddrs(vervName string) []string {
	matreshkaResolverPtr, err := s.Sd.GetResolver(vervName)
	if err != nil {
		return nil
	}

	return matreshkaResolverPtr.Load().GetAddrs()
}
