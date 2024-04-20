package env

import (
	"context"
	"strings"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"

	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
	"github.com/godverv/Velez/internal/clients/docker/dockerutils/list_request"
	"github.com/godverv/Velez/pkg/velez_api"
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
	if len(networks) == 0 {
		_, err = dockerAPI.NetworkCreate(ctx, VervNetwork, types.NetworkCreate{
			CheckDuplicate: true,
			Internal:       true,
			Attachable:     true,
		})
		if err != nil {
			return errors.Wrap(err, "error creating network")
		}
	}

	contId, err := IsInContainer(dockerAPI)
	if err != nil {
		return errors.Wrap(err, "error getting information about this instance")
	}
	if contId == nil {
		return nil
	}

	err = dockerAPI.NetworkConnect(ctx, VervNetwork, *contId, &network.EndpointSettings{
		Aliases: []string{"velez"},
	})
	if err != nil {
		return errors.Wrap(err, "error connecting this instance to network")
	}

	return nil
}

var instanceContainerId *string

// IsInContainer - function to determine weather
// this instance ran inside a container or as a standalone app
// returns container uuid if so
func IsInContainer(dockerAPI client.CommonAPIClient) (_ *string, err error) {
	if instanceContainerId == nil {
		instanceContainerId, err = isRunningInContainer(dockerAPI)
		if err != nil {
			return nil, errors.Wrap(err, "error checking if instance is running")
		}
		if instanceContainerId == nil {
			empty := ""
			instanceContainerId = &empty
		}
	}

	if *instanceContainerId == "" {
		return nil, nil
	}

	return instanceContainerId, nil
}

func isRunningInContainer(dockerAPI client.CommonAPIClient) (*string, error) {
	ctx := context.Background()
	containers, err := dockerutils.ListContainers(ctx, dockerAPI, &velez_api.ListSmerds_Request{})
	if err != nil {
		return nil, errors.Wrap(err, "error listing containers")
	}

	for _, container := range containers {
		if strings.HasPrefix(container.Image, "velez") {
			return &container.ID, nil
		}
	}

	return nil, nil
}
