package env

import (
	"context"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	errors "go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"

	"github.com/godverv/Velez/internal/clients/docker/dockerutils/list_request"
)

const (
	VervNetwork = "verv"
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

	contId := GetContainerId()
	if contId == nil {
		return nil
	}

	logrus.Debugf("Verv container at: %s", *contId)
	cont, err := dockerAPI.ContainerInspect(ctx, *contId)
	if err != nil {
		return errors.Wrap(err, "error getting Velez container info")
	}

	connection := &network.EndpointSettings{
		Aliases: []string{"velez"},
	}
	err = dockerAPI.NetworkConnect(ctx, VervNetwork, cont.Name, connection)
	if err != nil {
		return errors.Wrap(err, "error connecting this instance to network")
	}

	return nil
}

var instanceContainerId *string

// IsInContainer - function to determine weather
// this instance ran inside a container or as a standalone app
// returns container uuid if so
func IsInContainer() bool {
	return GetContainerId() != nil
}

func GetContainerId() *string {
	if instanceContainerId == nil {
		instanceContainerId = getContainerId()
		if instanceContainerId == nil {
			instanceContainerId = toolbox.ToPtr("")
		}
	}

	if *instanceContainerId == "" {
		return nil
	}

	return instanceContainerId
}

func getContainerId() *string {
	hm, err := os.ReadFile("/etc/hostname")
	if err != nil {
		return nil
	}
	return toolbox.ToPtr(strings.TrimRight(string(hm), "\n"))
}
