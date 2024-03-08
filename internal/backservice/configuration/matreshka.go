package configuration

import (
	"context"
	"time"

	errors "github.com/Red-Sock/trace-errors"

	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/pkg/velez_api"
)

const (
	containerName = "matreshka"
	image         = "godverv/matreshka-be"
	duration      = time.Second * 5
)

type Watchtower struct {
	cm service.ContainerManager

	duration time.Duration
}

func New(cfg config.Config, cm service.ContainerManager) *Watchtower {
	w := &Watchtower{
		cm:       cm,
		duration: duration,
	}

	return w
}

func (b *Watchtower) Start() error {
	ctx := context.Background()

	isAlive, err := b.IsAlive()
	if err != nil {
		return err
	}
	if isAlive {
		return nil
	}

	_, err = b.cm.LaunchSmerd(ctx, &velez_api.CreateSmerd_Request{
		Name:      containerName,
		ImageName: image,
	})
	if err != nil {
		return errors.Wrap(err, "error launching watchtower's smerd")
	}

	return nil
}

func (b *Watchtower) GetName() string {
	return containerName
}

func (b *Watchtower) GetDuration() time.Duration {
	return b.duration
}

func (b *Watchtower) IsAlive() (bool, error) {
	name := containerName

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

func (b *Watchtower) Kill() error {
	dropRes, err := b.cm.DropSmerds(context.Background(), &velez_api.DropSmerd_Request{
		Name: []string{containerName},
	})
	if err != nil {
		return errors.Wrap(err, "error dropping result")
	}

	if len(dropRes.Failed) != 0 {
		return errors.New(dropRes.Failed[0].Cause)
	}

	return nil
}
