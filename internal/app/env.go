package app

import (
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"
	"gopkg.in/yaml.v3"

	"go.vervstack.ru/Velez/internal/backservice/env"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
)

func setupVervNodeEnvironment(clients node_clients.NodeClients) (err error) {
	// Verv network for communication inside node
	err = env.StartNetwork(clients.Docker().Client())
	if err != nil {
		return rerrors.Wrap(err, "error creating network")
	}

	// Verv volumes for persistence inside node
	err = env.StartVolumes(clients.Docker().Client())
	if err != nil {
		return rerrors.Wrap(err, "error creating volumes")
	}

	return nil
}

func (c *Custom) saveSelfConfig(a *App) (err error) {
	storeConfig := &matreshka_api.StoreConfig_Request{
		Format:     matreshka_api.Format_yaml,
		ConfigName: env.VelezName,
		Config:     nil,
	}

	storeConfig.Config, err = yaml.Marshal(a.Cfg.MatreshkaConfig)
	if err != nil {
		return rerrors.Wrap(err, "error marshalling original config before saving it to matreshka")
	}

	_, err = c.ClusterClients.Configurator().StoreConfig(a.Ctx, storeConfig)
	if err != nil {
		return rerrors.Wrap(err, "error storing config in matreshka")
	}

	return nil
}
