package configurator

import (
	"context"

	"go.redsock.ru/evon"
	errors "go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka"
	"go.vervstack.ru/matreshka/pkg/matreshka_be_api"

	"github.com/godverv/Velez/internal/domain"
)

func (c *Configurator) UpdateConfig(ctx context.Context, meta domain.ConfigMeta, config matreshka.AppConfig) (err error) {
	patchRequest := &matreshka_be_api.PatchConfig_Request{
		ConfigName: meta.ServiceName,
		Version:    meta.CfgVersion,
	}

	envVars, err := evon.MarshalEnv(&config)
	if err != nil {
		return errors.Wrap(err, "error marshalling config")
	}

	nodesStorage := evon.NodesToStorage(envVars.InnerNodes)
	patchRequest.Patches = make([]*matreshka_be_api.PatchConfig_Patch, 0, len(envVars.InnerNodes))

	for _, node := range nodesStorage {
		if len(node.InnerNodes) == 0 && node.Value != nil {
			//TODO implement
			patchRequest.Patches = append(patchRequest.Patches)
		}
	}
	_, err = c.MatreshkaBeAPIClient.PatchConfig(ctx, patchRequest)
	if err != nil {
		return errors.Wrap(err, "error patching config")
	}

	return nil
}
