package dockerutils

import (
	"context"
	"strings"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

func CreateVolumeSoft(ctx context.Context, d client.CommonAPIClient, volumeName string) (volume.Volume, error) {
	vol, err := d.VolumeInspect(ctx, volumeName)
	if err != nil {
		if !strings.Contains(err.Error(), "no such volume") {
			return volume.Volume{}, errors.Wrap(err, "error inspecting network for service")
		}
	}

	if vol.Name != "" {
		return vol, nil
	}

	volCreateReq := volume.CreateOptions{
		Name: volumeName,
		DriverOpts: map[string]string{
			"type":   "none",
			"device": "~/verv",
			"o":      "bind",
		},
	}

	vol, err = d.VolumeCreate(ctx, volCreateReq)
	if err != nil {
		return vol, errors.Wrap(err, "error creating network for service")
	}

	return vol, nil
}
