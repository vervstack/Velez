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
	containerVervMountPoint = "/opt/velez/smerds"
)

var vervVolumeName = "verv"

func GetVervVolumeName() string {
	return vervVolumeName
}

func StartVolumes(dockerAPI client.CommonAPIClient) error {
	ctx := context.Background()

	isInContainer, err := IsInContainer(dockerAPI)
	if err != nil {
		return errors.Wrap(err, "error checking if velez is deployed in container")
	}

	if !isInContainer {
		vervVolumeName += "_host"
	}

	f := filters.NewArgs(filters.KeyValuePair{
		Key:   "name",
		Value: vervVolumeName,
	})

	volumes, err := dockerAPI.VolumeList(ctx, volume.ListOptions{Filters: f})
	if err != nil {
		return errors.Wrap(err, "error listing volumes")
	}

	var vervVolume *volume.Volume

	for _, v := range volumes.Volumes {
		if v.Name == vervVolumeName {
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
		vervVolumePath = containerVervMountPoint
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
		Name: vervVolumeName,
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

		vervVolumePath = path.Join(vervVolumePath, vervVolumeName)

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
