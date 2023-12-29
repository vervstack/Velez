package backservice

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
	cm service.ContainerManager

	duration time.Duration
}

func NewWatchTower(cfg config.Config, cm service.ContainerManager) *Watchtower {

	w := &Watchtower{
		cm: cm,
	}

	cfg.GetDuration()

	return w
}

func (b *Watchtower) Start() error {
	ctx := context.Background()

	command := "--interval 30"

	_, err := b.cm.LaunchSmerd(ctx, &velez_api.CreateSmerd_Request{
		Name:      watchTowerName,
		ImageName: watchTowerImage,
		Settings: &velez_api.Container_Settings{
			Volumes: []*velez_api.VolumeBindings{
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

}

func (b *Watchtower) Kill() error {

}
