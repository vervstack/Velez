package configuration

import (
	"context"
	"strconv"
	"strings"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/godverv/matreshka-be/pkg/matreshka_be_api"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/backservice/env"
	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/service/service_manager/container_manager_v1"
	"github.com/godverv/Velez/pkg/velez_api"
)

const (
	Name  = "matreshka"
	image = "godverv/matreshka-be:v1.0.31"
)

type MatreshkaTask struct {
	Address string

	dockerAPI client.CommonAPIClient

	ApiClient *ApiClient
	port      *int
}

func newKeepAliveTask(cfg config.Config, nodeClients clients.NodeClients) (*MatreshkaTask, error) {
	task := &MatreshkaTask{
		dockerAPI: nodeClients.Docker(),
	}

	var err error
	task.port, err = getPort(cfg, nodeClients)
	if err != nil {
		return nil, errors.Wrap(err, "error getting port")
	}

	task.Address, err = getTargetURL(cfg, nodeClients, task.port)
	if err != nil {
		return nil, errors.Wrap(err, "error getting target URL")
	}

	return task, nil
}

func (t *MatreshkaTask) Start() error {
	ctx := context.Background()

	_, err := dockerutils.PullImage(ctx, t.dockerAPI, image, false)
	if err != nil {
		return errors.Wrap(err, "error pulling matreshka image")
	}

	hostConf := &container.HostConfig{}

	if t.port != nil {
		hostConf.PortBindings = nat.PortMap{
			"53891/tcp": []nat.PortBinding{
				{
					HostPort: strconv.Itoa(*t.port),
				},
			},
		}
	}

	cont, err := t.dockerAPI.ContainerCreate(ctx,
		&container.Config{
			Hostname: Name,
			Image:    image,
			Labels: map[string]string{
				container_manager_v1.CreatedWithVelezLabel: "true",
			},
		},
		hostConf,
		&network.NetworkingConfig{},
		&v1.Platform{},
		Name,
	)
	if err != nil {
		return errors.Wrap(err, "error creating matreshka container")
	}

	err = t.dockerAPI.ContainerStart(ctx, cont.ID, container.StartOptions{})
	if err != nil {
		return errors.Wrap(err, "error starting matreshka container")
	}

	err = t.dockerAPI.NetworkConnect(
		ctx,
		env.VervNetwork,
		cont.ID,
		&network.EndpointSettings{Aliases: []string{Name}},
	)
	if err != nil {
		return errors.Wrap(err, "error connecting matreshka container to verv network")
	}

	return nil
}

func (t *MatreshkaTask) GetName() string {
	return Name
}

func (t *MatreshkaTask) IsAlive() bool {
	name := Name
	ctx := context.Background()

	cont, err := t.dockerAPI.ContainerInspect(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "No such container") {
			return false
		}
		logrus.Error(errors.Wrap(err, "error getting matreshka container"))
		return false
	}

	if cont.State.Status != velez_api.Smerd_running.String() {
		return false
	}

	t.ApiClient, err = newApiClient(t.Address)
	if err != nil {
		logrus.Error(errors.Wrap(err, "error creating api client"))
		return false
	}

	resp, err := t.ApiClient.MatreshkaBeAPIClient.ApiVersion(ctx, &matreshka_be_api.ApiVersion_Request{})
	if err != nil {
		return false
	}
	if resp == nil {
		return false
	}

	return false
}

func (t *MatreshkaTask) Kill() error {
	err := t.dockerAPI.ContainerRemove(context.Background(), Name, container.RemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	})
	if err != nil {
		if !strings.Contains(err.Error(), "No such container") {
			return errors.Wrap(err, "error dropping result")
		}
	}

	return nil
}
