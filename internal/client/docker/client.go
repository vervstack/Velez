package docker

import (
	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/client"
)

func NewClient() (client.CommonAPIClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, errors.Wrap(err, "error getting docker client")
	}

	return cli, nil
}
