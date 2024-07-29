package service_discovery_task

import (
	"context"
	"strings"
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
	image    = "godverv/makosh:v0.0.3"
	duration = time.Second * 5

	makoshContainerAuthTokenEnvVariable = "MAKOSH_ENVIRONMENT_AUTH-TOKEN"
)

var (
	ErrRequireMakoshPortExportToRunAsDaemon = errors.New("makosh port must be exported in order to run velez as daemon")
)

type ServiceDiscoveryTask struct {
	targetURL string

	authToken string

	dockerAPI      client.CommonAPIClient
	image          string
	portToExposeTo *string
	duration       time.Duration
}

func New(cfg config.Config, internalClients clients.InternalClients) (*ServiceDiscoveryTask, error) {
	envVar := cfg.GetEnvironment()

	serviceDiscoveryTask := &ServiceDiscoveryTask{
		dockerAPI: internalClients.Docker(),
		image:     rtb.Coalesce(envVar.MakoshImageName, image),
		duration:  duration,
	}

	var err error
	serviceDiscoveryTask.portToExposeTo, err = getPortToExposeTo(envVar, internalClients)
	if err != nil {
		return nil, errors.Wrap(err, "error getting port to expose to makosh")
	}

	serviceDiscoveryTask.targetURL, err = getTargetURL(envVar, internalClients, serviceDiscoveryTask.portToExposeTo)
	if err != nil {
		return nil, errors.Wrap(err, "error getting target URL")
	}

	serviceDiscoveryTask.authToken, err = generateAuthToken()
	if err != nil {
		return nil, errors.Wrap(err, "error generating auth token")
	}

	return serviceDiscoveryTask, nil
}

func (s *ServiceDiscoveryTask) Start() error {
	isAlive, err := s.IsAlive()
	if err != nil {
		return errors.Wrap(err)
	}
	if isAlive {
		logrus.Info("Makosh is already running")
		return err
	}

	ctx := context.Background()

	_, err = dockerutils.PullImage(ctx, s.dockerAPI, s.image, false)
	if err != nil {
		return errors.Wrap(err, "error pulling matreshka image")
	}

	hostConf := &container.HostConfig{}

	if s.portToExposeTo != nil {
		hostConf.PortBindings = nat.PortMap{
			"80/tcp": []nat.PortBinding{
				{
					HostPort: *s.portToExposeTo,
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
			Env: []string{makoshContainerAuthTokenEnvVariable + "=" + s.authToken},
		},
		hostConf,
		&network.NetworkingConfig{},
		&v1.Platform{},
		Name,
	)
	if err != nil {
		return errors.Wrap(err, "error creating makosh container")
	}

	err = s.dockerAPI.ContainerStart(ctx, cont.ID, container.StartOptions{})
	if err != nil {
		return errors.Wrap(err, "error starting makosh container")
	}

	err = s.dockerAPI.NetworkConnect(ctx,
		env.VervNetwork,
		cont.ID,
		&network.EndpointSettings{
			Aliases: []string{Name},
		})
	if err != nil {
		return errors.Wrap(err, "error connecting makosh container to verv network")
	}

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
	ctx := context.Background()

	cont, err := s.dockerAPI.ContainerInspect(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "No such container") {
			return false, nil
		}

		return false, errors.Wrap(err, "error getting makosh container")
	}

	if cont.State.Status != velez_api.Smerd_running.String() {
		return false, nil
	}

	if !rtb.Contains(cont.Config.Env, makoshContainerAuthTokenEnvVariable+"="+s.authToken) {
		return false, nil
	}

	return true, nil
}

func (s *ServiceDiscoveryTask) Kill() error {
	ctx := context.Background()
	rmOpts := container.RemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}

	err := s.dockerAPI.ContainerRemove(ctx, Name, rmOpts)
	if err != nil {
		if !strings.Contains(err.Error(), "No such container") {
			return errors.Wrap(err, "error dropping result")
		}
	}

	return nil
}
