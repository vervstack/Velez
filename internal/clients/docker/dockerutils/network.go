package dockerutils

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/godverv/Velez/internal/clients/docker/dockerutils/list_request"
)

func CreateNetworkSoft(ctx context.Context, d client.CommonAPIClient, networkName string) error {
	f := list_request.New()
	f.Name(networkName)
	net, err := d.NetworkList(ctx, types.NetworkListOptions{
		Filters: f.Args(),
	})
	if err != nil {
		return errors.Wrap(err, "error inspecting network for service")
	}

	for _, item := range net {
		if item.Name == networkName {
			return nil
		}
	}

	_, err = d.NetworkCreate(ctx,
		networkName,
		types.NetworkCreate{})
	if err != nil {
		return errors.Wrap(err, "error creating network for service")
	}

	return nil
}
