package configurator

import (
	"context"
	"fmt"

	"github.com/Red-Sock/evon"
	rtb "github.com/Red-Sock/toolbox"
	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/matreshka"
	"github.com/godverv/matreshka-be/pkg/matreshka_be_api"
)

func (c *Configurator) UpdateConfig(ctx context.Context, serviceName string, config matreshka.AppConfig) (err error) {
	patchRequest := &matreshka_be_api.PatchConfig_Request{}
	patchRequest.ServiceName = serviceName

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
