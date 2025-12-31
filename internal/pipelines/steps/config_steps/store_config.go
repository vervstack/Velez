package config_steps

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
)

type storeConfigStep struct {
	configApi cluster_clients.Configurator

	configName string
	content    []byte
	prefix     matreshka_api.ConfigTypePrefix
	format     matreshka_api.Format
}

func StoreConfig(
	clusterClients cluster_clients.ClusterClients,
	configName string,
	content []byte,
	prefix matreshka_api.ConfigTypePrefix,
	format matreshka_api.Format) steps.Step {
	return &storeConfigStep{
		clusterClients.Configurator(),
		configName,
		content,
		prefix,
		format,
	}
}

func (s *storeConfigStep) Do(ctx context.Context) error {
	storeConfigReq := &matreshka_api.StoreConfig_Request{
		Format:     s.format,
		ConfigName: s.configName,
		Config:     s.content,
	}

	_, err := s.configApi.StoreConfig(ctx, storeConfigReq)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}
