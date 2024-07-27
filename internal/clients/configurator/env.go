package configurator

import (
	"context"
	"fmt"

	"github.com/Red-Sock/evon"
	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/matreshka"
	"github.com/godverv/matreshka-be/pkg/matreshka_api"
)

func (c *Configurator) GetEnv(ctx context.Context, name string) ([]string, error) {
	getConfigReq := &matreshka_api.GetConfig_Request{
		ServiceName: name,
	}
	raw, err := c.matreshkaClient.GetConfig(ctx, getConfigReq)
	if err != nil {
		return nil, errors.Wrap(err, "error getting config from api")
	}

	if len(raw.Config) == 0 {
		return nil, nil
	}

	cfg := matreshka.NewEmptyConfig()
	err = cfg.Unmarshal(raw.Config)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling config")
	}

	evonEnv, err := evon.MarshalEnv(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "error marshalling to env")
	}
	ns := evon.NodesToStorage(evonEnv.InnerNodes)

	// TODO check if works correctly
	env := make([]string, 0, len(ns))
	for _, n := range ns {
		env = append(env, n.Name+"="+fmt.Sprint(n.Value))
	}

	return env, nil
}
