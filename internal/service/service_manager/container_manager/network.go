package container_manager

import (
	"context"
	"errors"
	"strings"

	"github.com/docker/docker/errdefs"
	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

func (c *ContainerManager) ConnectToNetwork(ctx context.Context, req domain.Connection) error {
	runtime, err := c.runtimes.Runtime(ctx, req.Environment)
	if err != nil {
		return rerrors.Wrap(err, "error resolving environment")
	}

	connReq := container_runtime.ConnectToNetworkRequest{
		ContainerID: req.SmerdName,
		NetworkName: req.Network,
		Aliases:     req.Aliases,
	}

	err = runtime.ConnectToNetwork(ctx, connReq)
	if err == nil {
		return nil
	}

	var notFound errdefs.ErrNotFound

	if errors.As(err, &notFound) {
		return rerrors.Wrap(user_errors.ErrNetworkNotFound)
	}

	return rerrors.Wrap(err, "error connecting to network")
}

func (c *ContainerManager) DisconnectFromNetwork(ctx context.Context, req domain.Connection) error {
	runtime, err := c.runtimes.Runtime(ctx, req.Environment)
	if err != nil {
		return rerrors.Wrap(err, "error resolving environment")
	}

	err = runtime.DisconnectFromNetworks(ctx, req.SmerdName, []string{req.Network})
	if err == nil {
		return nil
	}

	if strings.Contains(err.Error(), "is not connected to the network") {
		return nil
	}

	return rerrors.Wrap(err, "error connecting to network")
}
