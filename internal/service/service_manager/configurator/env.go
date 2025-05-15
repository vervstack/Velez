package configurator

import (
	"context"

	"go.redsock.ru/evon"
	errors "go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"
	"go.vervstack.ru/matreshka/pkg/matreshka_be_api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/godverv/Velez/internal/domain"
)

func (c *Configurator) GetEnvFromApi(ctx context.Context, meta domain.ConfigMeta) ([]*evon.Node, error) {
	req := &matreshka_be_api.GetConfigNode_Request{
		ConfigName: meta.ServiceName,
		//TODO replace master down below onto constant from matreshka
		Version: toolbox.Coalesce(toolbox.FromPtr(meta.CfgVersion), "master"),
	}

	cfgNodes, err := c.MatreshkaBeAPIClient.GetConfigNodes(ctx, req)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
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
