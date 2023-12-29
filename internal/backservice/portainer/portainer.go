package portainer

import (
	"context"
	"time"

	errors "github.com/Red-Sock/trace-errors"

	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/pkg/velez_api"
)

const (
	portainerName     = "portainer-ce"
	portainerImage    = "portainer/" + portainerName
	portainerDuration = time.Second * 5
)

type Portainer struct {
	cm service.ContainerManager

	duration time.Duration
}

func NewPortainer(cm service.ContainerManager) *Portainer {
	w := &Portainer{
		cm:       cm,
		duration: portainerDuration,
	}

	return w
}

func (b *Portainer) Start() error {
	ctx := context.Background()

	command := "--interval 30"

	_, err := b.cm.LaunchSmerd(ctx, &velez_api.CreateSmerd_Request{
		Name:      portainerName,
		ImageName: portainerImage,
		Settings: &velez_api.Container_Settings{
			Ports: []*velez_api.PortBindings{},
			Volumes: []*velez_api.VolumeBindings{
				{
					Host:      "/var/run/docker.sock",
					Container: "/var/run/docker.sock",
				},
				{
					Host:      "portainer_data",
					Container: "/data",
				},
			},
		},
		Command: &command,
	})
	if err != nil {
		return errors.Wrap(err, "error launching watchtower's smerd")
	}

	return nil
}

func (b *Portainer) GetName() string {
	return portainerName
}

func (b *Portainer) GetDuration() time.Duration {
	return b.duration
}

func (b *Portainer) IsAlive() (bool, error) {
	name := portainerName

	smerds, err := b.cm.ListSmerds(context.Background(), &velez_api.ListSmerds_Request{Name: &name})
	if err != nil {
		return false, errors.Wrap(err, "error listing smerds with name "+name)
	}

	for _, smerd := range smerds.Smerds {
		if smerd.Name == name && smerd.Status == velez_api.Smerd_running {
			return true, nil
		}
	}

	return false, nil
}

func (b *Portainer) Kill() error {
	dropRes, err := b.cm.DropSmerds(context.Background(), &velez_api.DropSmerd_Request{
		Name: []string{portainerName},
	})
	if err != nil {
		return errors.Wrap(err, "error dropping result")
	}

	if len(dropRes.Failed) != 0 {
		return errors.New(dropRes.Failed[0].Cause)
	}

	return nil
}
