package configurator

import (
	"context"

	"go.redsock.ru/evon"
	errors "go.redsock.ru/rerrors"
	"go.verv.tech/matreshka-be/pkg/matreshka_be_api"
)

func (c *Configurator) GetEnvFromApi(ctx context.Context, serviceName string) ([]*evon.Node, error) {
	req := &matreshka_be_api.GetConfigNode_Request{
		ServiceName: serviceName,
	}
	cfgNodes, err := c.MatreshkaBeAPIClient.GetConfigNodes(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "error obtaining raw config")
	}

	if cfgNodes.Root == nil {
		return nil, nil
	}

	return fromApiNodes(cfgNodes.Root), nil
}

func fromApiNodes(root *matreshka_be_api.Node) []*evon.Node {
	out := make([]*evon.Node, 0)

	for _, node := range root.InnerNodes {
		if len(node.InnerNodes) != 0 {
			out = append(out, fromApiNodes(node)...)
		}
		if node.Value != nil {
			out = append(out, &evon.Node{
				Name:  node.Name,
				Value: *node.Value,
			})
		}
	}

	return out
}
