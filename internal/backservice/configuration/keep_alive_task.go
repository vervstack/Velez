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

	port *int
}

func newTask(cfg config.Config, cls clients.NodeClients) (*MatreshkaTask, error) {
	w := &MatreshkaTask{
		dockerAPI: cls.Docker(),
	}
	var err error
	w.port, err = getPort(cfg, cls)
	if err != nil {
		return nil, errors.Wrap(err, "error getting port")
	}

	w.Address, err = getTargetURL(cfg, cls, w.port)
	if err != nil {
		return nil, errors.Wrap(err, "error getting target URL")
	}

	return w, nil
}

func (b *MatreshkaTask) Start() error {
	isAlive, err := b.IsAlive()
	if err != nil {
		return err
	}
	if isAlive {
		logrus.Info("Matreshka is already running")
		return err
	}

	ctx := context.Background()

	_, err = dockerutils.PullImage(ctx, b.dockerAPI, image, false)
	if err != nil {
		return errors.Wrap(err, "error pulling matreshka image")
	}

	hostConf := &container.HostConfig{}

	if b.port != nil {
		hostConf.PortBindings = nat.PortMap{
			"53891/tcp": []nat.PortBinding{
				{
					HostPort: strconv.Itoa(*b.port),
				},
			},
		}
	}

	cont, err := b.dockerAPI.ContainerCreate(ctx,
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

	err = b.dockerAPI.ContainerStart(ctx, cont.ID, container.StartOptions{})
	if err != nil {
		return errors.Wrap(err, "error starting matreshka container")
	}

	err = b.dockerAPI.NetworkConnect(ctx, env.VervNetwork, cont.ID, &network.EndpointSettings{
		Aliases: []string{Name},
	})
	if err != nil {
		return errors.Wrap(err, "error connecting matreshka container to verv network")
	}

	return nil
}

func (b *MatreshkaTask) GetName() string {
	return Name
}

func (b *MatreshkaTask) IsAlive() (bool, error) {
	name := Name

	containers, err := dockerutils.ListContainers(
		context.Background(),
		b.dockerAPI, &velez_api.ListSmerds_Request{
			Name: &name,
		})
	if err != nil {
		return false, errors.Wrap(err, "error listing smerds with name "+name)
	}

	for _, cont := range containers {
		hasName := false
		for _, cNname := range cont.Names {
			if name == cNname[1:] {
				hasName = true
				break
			}
		}

		if hasName && cont.State == velez_api.Smerd_running.String() {
			return true, nil
		}
	}

	return false, nil
}

func (b *MatreshkaTask) Kill() error {
	err := b.dockerAPI.ContainerRemove(context.Background(), Name, container.RemoveOptions{
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
