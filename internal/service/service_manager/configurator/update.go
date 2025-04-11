package configurator

import (
	"context"
	"fmt"

	"go.redsock.ru/evon"
	errors "go.redsock.ru/rerrors"
	rtb "go.redsock.ru/toolbox"
	"go.vervstack.ru/matreshka"
	"go.vervstack.ru/matreshka-be/pkg/matreshka_be_api"

	"github.com/godverv/Velez/internal/domain"
)

func (c *Configurator) UpdateConfig(ctx context.Context, meta domain.ConfigMeta, config matreshka.AppConfig) (err error) {
	patchRequest := &matreshka_be_api.PatchConfig_Request{
		ServiceName: meta.ServiceName,
		Version:     meta.CfgVersion,
	}

	envVars, err := evon.MarshalEnv(&config)
	if err != nil {
		return errors.Wrap(err, "error marshalling config")
	}

	nodesStorage := evon.NodesToStorage(envVars.InnerNodes)
	patchRequest.Changes = make([]*matreshka_be_api.Node, 0, len(envVars.InnerNodes))

	for _, node := range nodesStorage {
		if len(node.InnerNodes) == 0 && node.Value != nil {
			patchRequest.Changes = append(patchRequest.Changes,
				&matreshka_be_api.Node{
					Name:  node.Name,
					Value: rtb.ToPtr(fmt.Sprint(node.Value)),
				})
		}
	}
	_, err = c.MatreshkaBeAPIClient.PatchConfig(ctx, patchRequest)
	if err != nil {
		return errors.Wrap(err, "error patching config")
	}

	return nil
}
