package env

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	errors "go.redsock.ru/rerrors"

	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
	"github.com/godverv/Velez/internal/clients/docker/dockerutils/list_request"
	"github.com/godverv/Velez/pkg/velez_api"
)

const (
	VervNetwork    = "verv"
	velezImageName = "godverv/velez"
)

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
		logrus.Debug("Verv network not found. Creating...")
		createNetReq := network.CreateOptions{
			Internal:   true,
			Attachable: true,
		}

		_, err = dockerAPI.NetworkCreate(ctx, VervNetwork, createNetReq)
		if err != nil {
			return errors.Wrap(err, "error creating network")
		}
		logrus.Debug("Verv network created")
	}

	contId := GetContainerId(dockerAPI)
	if contId == nil {
		return nil
	}

	connection := &network.EndpointSettings{
		Aliases: []string{"velez"},
	}
	err = dockerAPI.NetworkConnect(ctx, VervNetwork, *contId, connection)
	if err != nil {
		return errors.Wrap(err, "error connecting this instance to network")
	}

	return nil
}

var instanceContainerId *string

// IsInContainer - function to determine weather
// this instance ran inside a container or as a standalone app
// returns container uuid if so
func IsInContainer(dockerAPI client.CommonAPIClient) bool {
	return GetContainerId(dockerAPI) != nil
}

func GetContainerId(dockerAPI client.CommonAPIClient) *string {
	var err error

	if instanceContainerId == nil {
		instanceContainerId, err = getContainerId(dockerAPI)
		if err != nil {
			return nil
		}
		if instanceContainerId == nil {
			empty := ""
			instanceContainerId = &empty
		}
	}

	if *instanceContainerId == "" {
		return nil
	}

	return instanceContainerId
}

func getContainerId(dockerAPI client.CommonAPIClient) (*string, error) {
	ctx := context.Background()
	containers, err := dockerutils.ListContainers(ctx, dockerAPI, &velez_api.ListSmerds_Request{})
	if err != nil {
		return nil, errors.Wrap(err, "error listing containers")
	}

	for _, container := range containers {
		if strings.Contains(container.Image, "velez:") {
			return &container.ID, nil
		}
	}

	return nil, nil
}
