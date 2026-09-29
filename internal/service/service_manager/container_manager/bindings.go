package container_manager

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

// boundServiceNames maps container name to the service it is bound to on this
// node. Empty when no bindings backend is live (single mode).
func (c *ContainerManager) boundServiceNames(ctx context.Context, environment string) (map[string]string, error) {
	names := make(map[string]string)

	if c.bindings == nil {
		return names, nil
	}

	bindingsStorage := c.bindings.ContainerBindings()
	if bindingsStorage == nil {
		return names, nil
	}

	list, err := bindingsStorage.ListByNode(ctx, domain.SelfNodeId, environment)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing container bindings")
	}

	for _, binding := range list {
		names[binding.ContainerName] = binding.ServiceName
	}

	return names, nil
}
