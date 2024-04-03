package env

import (
	"context"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/godverv/Velez/internal/clients/docker/dockerutils/list_request"
)

const VervNetwork = "verv"

func StartNetwork(dockerAPI client.CommonAPIClient) error {
	ctx := context.Background()
	args := list_request.New()
	args.Name(VervNetwork)

	networks, err := dockerAPI.NetworkList(ctx, types.NetworkListOptions{
		Filters: args.Args(),
	})
	if err != nil {
		return errors.Wrap(err, "error listing networks")
	}
	if len(networks) != 0 {
		return nil
	}

	_, err = dockerAPI.NetworkCreate(ctx, VervNetwork, types.NetworkCreate{
		CheckDuplicate: true,
		Internal:       true,
		Attachable:     true,
	})
	if err != nil {
		return errors.Wrap(err, "error creating network")
	}

	return err
}
