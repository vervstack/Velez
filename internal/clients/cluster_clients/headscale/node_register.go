package headscale

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
)

func (s *Client) RegisterNode(ctx context.Context, req domain.RegisterVcnNodeReq) error {
	//region Dto
	type reqBody struct {
		Key      string
		Username string `json:"user"`
	}

	//endregion

	execOpts := container.ExecOptions{
		Cmd: []string{
			"headscale", "nodes", "register",
			"--key", req.Key,
			"--user", req.Username,
		},
		Detach: true,
	}
	_, err := s.docker.Exec(ctx, s.containerName, execOpts)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}
