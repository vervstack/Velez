package watchtower

import (
	"context"
	"time"

	errors "github.com/Red-Sock/trace-errors"

	"github.com/godverv/Velez/internal/config"
	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/pkg/velez_api"
)

const (
	watchTowerName     = "watchtower"
	watchTowerImage    = "containrrr/" + watchTowerName
	watchTowerDuration = time.Second * 5
)

type Watchtower struct {
	cm service.Services

	duration time.Duration
}

func New(cfg config.Config, cm service.Services) *Watchtower {
	w := &Watchtower{
		cm: cm,
	}

	w.duration = cfg.GetEnvironment().WatchTowerInterval
	if w.duration == 0 {
		w.duration = watchTowerDuration
	}

	return w
}

func (b *Watchtower) Start() error {
	ctx := context.Background()

	command := "--interval 30"
	name := watchTowerName
	_, err := b.cm.LaunchSmerd(ctx, &velez_api.CreateSmerd_Request{
		Name:      name,
		ImageName: watchTowerImage,
		Settings: &velez_api.Container_Settings{
			Mounts: []*velez_api.MountBindings{
				{
					Host:      "/var/run/docker.sock",
					Container: "/var/run/docker.sock",
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

func (b *Watchtower) GetName() string {
	return watchTowerName
}

func (b *Watchtower) GetDuration() time.Duration {
	return b.duration
}

func (b *Watchtower) IsAlive() (bool, error) {
	name := watchTowerName

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
		Name: []string{watchTowerName},
	})
	if err != nil {
		return errors.Wrap(err, "error dropping result")
	}

	if len(dropRes.Failed) != 0 {
		return errors.New(dropRes.Failed[0].Cause)
	}

	return nil
}
