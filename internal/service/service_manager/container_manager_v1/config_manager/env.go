package config_manager

import (
	"context"
	"fmt"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/matreshka"
	"github.com/godverv/matreshka-be/pkg/matreshka_api"
)

func (c *Configurator) GetEnv(ctx context.Context, name string) ([]string, error) {
	raw, err := c.matreshkaClient.GetConfigRaw(ctx, &matreshka_api.GetConfigRaw_Request{
		ServiceName: name,
	})
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

	mEnv := matreshka.GenerateEnvironmentKeys(cfg.Environment)

	env := make([]string, len(mEnv))

	for i := range mEnv {
		env[i] = fmt.Sprintf("%s=%v", mEnv[i].Name, mEnv[i].Value)
	}

	return env, nil
}
