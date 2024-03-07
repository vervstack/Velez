package config_manager

import (
	"context"
	"fmt"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/matreshka"
	"github.com/godverv/matreshka-be/pkg/matreshka_api"
	"github.com/sirupsen/logrus"
)

type Configurator struct {
	matreshkaClient matreshka_api.MatreshkaBeAPIClient
}

func New(api matreshka_api.MatreshkaBeAPIClient) *Configurator {
	return &Configurator{
		matreshkaClient: api,
	}
}

func (c *Configurator) GetConfigEnvs(ctx context.Context, name string) ([]string, error) {
	confRaw, err := c.matreshkaClient.GetConfigRaw(ctx,
		&matreshka_api.GetConfigRaw_Request{
			ServiceName: name,
		},
	)
	if err != nil {
		logrus.Warnf("error getting config for service \"%s\". Error: %s", name, err)
		return nil, nil
	}

	if confRaw == nil || confRaw.Config == "" {
		logrus.Warnf("no config returned for service \"%s\"", name)
		return nil, nil
	}

	conf, err := matreshka.ParseConfig([]byte(confRaw.Config))
	if err != nil {
		return nil, errors.Wrap(err, "error parsing config")
	}

	envKeys, err := matreshka.GenerateKeys(conf)
	if err != nil {
		return nil, errors.Wrap(err, "error generating keys")
	}

	keys := make([]string, 0, len(envKeys))

	for _, item := range envKeys {
		keys = append(keys, item.Name+"="+fmt.Sprintf("%v", item.Value))
	}

	return keys, nil
}
