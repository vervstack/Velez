package env

import (
	"context"
	"os"
	"path"
	"sync"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	errors "go.redsock.ru/rerrors"
)

const (
	containerVervMountPoint = "/opt/velez/smerds"

	vervVolumeBaseName = "verv"
)

// volumeMu guards vervVolumeName and vervVolumePath: StartVolumes can be
// called concurrently by multiple in-process environments (e.g. parallel
// e2e TestEnvironments), and without this lock that's a data race caught by
// `go test -race`.
var (
	volumeMu       sync.RWMutex
	vervVolumeName = vervVolumeBaseName
	vervVolumePath string
)

func GetVervVolumeName() string {
	volumeMu.RLock()
	defer volumeMu.RUnlock()

	return vervVolumeName
}

func StartVolumes(dockerAPI client.APIClient) error {
	volumeMu.Lock()
	defer volumeMu.Unlock()

	ctx := context.Background()

	isInContainer := IsInContainer()

	vervVolumeName = vervVolumeBaseName
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

func GetVervVolumePath() (string, error) {
	volumeMu.RLock()
	defer volumeMu.RUnlock()

	if vervVolumePath != "" {
		return vervVolumePath, nil
	}

	return "", errors.New("no verv volume found")
}

// createVervVolume is only ever called by StartVolumes, with volumeMu
// already held — it must not lock volumeMu itself.
func createVervVolume(ctx context.Context, dockerAPI client.APIClient) (*volume.Volume, error) {
	createOptions := volume.CreateOptions{
		Name: vervVolumeName,
	}

	isInContainer := IsInContainer()

	if !isInContainer {
		var err error

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
		//nolint
		err = os.MkdirAll(vervVolumePath, 0o755)
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
