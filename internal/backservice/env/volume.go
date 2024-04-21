package env

import (
	"context"
	"os"
	"path"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

const (
	VervVolumeName = "verv"
)

func StartVolumes(dockerAPI client.CommonAPIClient) error {
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

	if vervVolume == nil {
		vervVolume, err = createVervVolume(ctx, dockerAPI)
		if err != nil {
			return errors.Wrap(err, "error creating verv volume")
		}
	}

	vervVolumePath = vervVolume.Options["device"]
	if vervVolumePath == "" {
		vervVolumePath = vervVolume.Mountpoint
	}

	return nil
}

var vervVolumePath string

func GetVervVolumePath() (string, error) {
	if vervVolumePath != "" {
		return vervVolumePath, nil
	}

	return "", errors.New("no verv volume found")
}

func createVervVolume(ctx context.Context, dockerAPI client.CommonAPIClient) (*volume.Volume, error) {
	createOptions := volume.CreateOptions{
		Name: VervVolumeName,
	}

	isInContainer, err := IsInContainer(dockerAPI)
	if err != nil {
		return nil, errors.Wrap(err, "error detecting if velez ran in container or as standalone app")
	}

	if !isInContainer {
		vervVolumePath, err = os.UserCacheDir()
		if err != nil {
			return nil, errors.Wrap(err, "error getting cache dir")
		}

		vervVolumePath = path.Join(vervVolumePath, VervVolumeName)

		createOptions.DriverOpts = map[string]string{
			"type":   "none",
			"device": vervVolumePath,
			"o":      "bind",
		}

		err = os.MkdirAll(vervVolumePath, 0755)
		if err != nil {
			return nil, errors.Wrap(err, "error creating verv volume dir")
		}
	}

	v, err := dockerAPI.VolumeCreate(ctx, createOptions)
	if err != nil {
		return nil, errors.Wrap(err, "error creating verv volume")
	}

	return &v, nil
}
