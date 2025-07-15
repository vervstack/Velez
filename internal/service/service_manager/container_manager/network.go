package container_manager

import (
	"context"
	"errors"
	"strings"

	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/errdefs"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

func (c *ContainerManager) ConnectToNetwork(ctx context.Context, req domain.Connection) error {
	es := network.EndpointSettings{
		Aliases: req.Aliases,
	}

	err := c.docker.NetworkConnect(ctx, req.Network, req.SmerdName, &es)
	if err == nil {
		return nil
	}

	var notFound errdefs.ErrNotFound
	if errors.As(err, &notFound) {
		return rerrors.NewUserError("network not found")
	}

	return rerrors.Wrap(err, "error connecting to network")
}

func (c *ContainerManager) DisconnectFromNetwork(ctx context.Context, req domain.Connection) error {
	err := c.docker.NetworkDisconnect(ctx, req.Network, req.SmerdName, true)
	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), "is not connected to the network") {
		return nil
	}

	return rerrors.Wrap(err, "error connecting to network")
}
