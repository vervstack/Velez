package service_discovery

import (
	"context"
	"strconv"
	"time"

	rtb "github.com/Red-Sock/toolbox"
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
	Name     = "makosh"
	duration = time.Second * 5
)

type ServiceDiscoveryTask struct {
	authToken string

	dockerAPI    client.CommonAPIClient
	image        string
	portToExpose *uint32
	duration     time.Duration
}

func New(cfg config.Config, cls clients.Clients) (*ServiceDiscoveryTask, error) {
	envVar := cfg.GetEnvironment()
	var portToExposeTo *uint32
	var err error
	if envVar.MakoshExposePort {
		p := uint32(envVar.MakoshPort)

		if p == 0 {
			p, err = cls.PortManager().GetPort()
			if err != nil {
				return nil, errors.Wrap(err, "error obtaining port from pool")
			}
		}

		portToExposeTo = &p
	}

	key := envVar.MakoshKey
	if rtb.IsEmpty(key) {
		var keyBytes []byte
		keyBytes, err = rtb.Random(256)
		if err != nil {
			return nil, errors.Wrap(err, "error generating random makosh key")
		}

		key = string(keyBytes)
	}

	return &ServiceDiscoveryTask{
		authToken: string(key),

		image:        envVar.MakoshImageName,
		portToExpose: portToExposeTo,
		dockerAPI:    cls.Docker(),
		duration:     duration,
	}, nil
}

func (s *ServiceDiscoveryTask) Start() error {
	isAlive, err := s.IsAlive()
	if err != nil {
		return err
	}
	if isAlive {
		logrus.Info("Matreshka is already running")
		return err
	}

	ctx := context.Background()

	_, err = dockerutils.PullImage(ctx, s.dockerAPI, s.image, false)
	if err != nil {
		return errors.Wrap(err, "error pulling matreshka image")
	}

	hostConf := &container.HostConfig{}

	if s.portToExpose != nil {
		hostConf.PortBindings = nat.PortMap{
			"53891/tcp": []nat.PortBinding{
				{
					HostPort: strconv.Itoa(int(*s.portToExpose)),
				},
			},
		}
	}

	cont, err := s.dockerAPI.ContainerCreate(ctx,
		&container.Config{
			Hostname: Name,
			Image:    s.image,
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

	err = s.dockerAPI.ContainerStart(ctx, cont.ID, container.StartOptions{})
	if err != nil {
		return errors.Wrap(err, "error starting matreshka container")
	}

	err = s.dockerAPI.NetworkConnect(ctx, env.VervNetwork, cont.ID, &network.EndpointSettings{
		Aliases: []string{Name},
	})
	if err != nil {
		return errors.Wrap(err, "error connecting matreshka container to verv network")
	}

	logrus.Info("Matreshka successfully started")

	return nil
}

func (s *ServiceDiscoveryTask) GetName() string {
	return Name
}

func (s *ServiceDiscoveryTask) GetDuration() time.Duration {
	return s.duration
}

func (s *ServiceDiscoveryTask) IsAlive() (bool, error) {
	name := Name

	containers, err := dockerutils.ListContainers(
		context.Background(),
		s.dockerAPI, &velez_api.ListSmerds_Request{
			Name: &name,
		})
	if err != nil {
		return false, errors.Wrap(err, "error listing smerds with name "+name)
	}

	for _, cont := range containers {
		hasName := false
		for _, contName := range cont.Names {
			if name == contName[1:] {
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

func (s *ServiceDiscoveryTask) Kill() error {
	ctx := context.Background()
	rmOpts := container.RemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}

	err := s.dockerAPI.ContainerRemove(ctx, Name, rmOpts)
	if err != nil {
		return errors.Wrap(err, "error dropping result")
	}

	return nil
}
