package dockerutils

import (
	"context"

	"github.com/docker/docker/client"
	"go.redsock.ru/rerrors"
)

func HasRuntime(ctx context.Context, cli client.APIClient, name string) (bool, error) {
	info, err := cli.Info(ctx)
	if err != nil {
		return false, rerrors.Wrap(err, "error getting docker info")
	}

	_, isRegistered := info.Runtimes[name]

	return isRegistered, nil
}
