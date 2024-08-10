package grpc

import (
	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/makosh/pkg/makosh_resolver"
	"google.golang.org/grpc/resolver"

	"github.com/godverv/Velez/internal/config"
)

// RegisterServiceDiscovery - requires environment variables
// MAKOSH_URL and MAKOSH_SECRET to be specified to work properly
func RegisterServiceDiscovery(cfg config.Config) error {
	overrides := make(map[string][]string)
	for _, sd := range cfg.GetServiceDiscovery().Overrides {
		overrides[sd.ServiceName] = sd.Urls
	}

	makoshOpts := makosh_resolver.WithOverrides(overrides)

	resolverBuilder, err := makosh_resolver.New(makoshOpts)
	if err != nil {
		return errors.Wrap(err, "error creating makosh resolver")
	}

	resolver.Register(resolverBuilder)

	return nil
}
