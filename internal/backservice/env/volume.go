package env

import (
	"context"
	"os"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"

	"github.com/godverv/Velez/internal/config"
)

const (
	DefaultSmerdsVolumesPath = "/opt/velez/smerds/"
	VervVolumeName           = "verv"
)

func StartVolumes(cfg config.Config, dockerAPI client.CommonAPIClient) error {
	volumePath := cfg.GetString(config.SmerdVolumePath)
	if volumePath == "" {
		volumePath = DefaultSmerdsVolumesPath
	}

	err := os.MkdirAll(volumePath, 0777)
	if err != nil {
		return errors.Wrap(err, "error creating velez volume folder")
	}
	ctx := context.Background()

	f := filters.NewArgs(filters.KeyValuePair{
		Key:   "name",
		Value: VervVolumeName,
	})

	volumes, err := dockerAPI.VolumeList(ctx, volume.ListOptions{Filters: f})
	if err != nil {
		return errors.Wrap(err, "error listing volumes")
	}

	var vervVolume *volume.Volume

	for _, v := range volumes.Volumes {
		if v.Name == VervVolumeName {
			vervVolume = v
			break
		}
	}

	if vervVolume != nil {
		return nil
	}

	_, err = dockerAPI.VolumeCreate(ctx, volume.CreateOptions{
		Name: VervVolumeName,

		DriverOpts: map[string]string{
			"type":   "none",
			"device": volumePath,
			"o":      "bind",
		},
	})
	if err != nil {
		return errors.Wrap(err, "error creating verv volume")
	}

	return nil
}
