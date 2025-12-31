package headscale

import (
	"context"

	"go.vervstack.ru/Velez/internal/domain"
)

func (s *Client) RegisterNode(ctx context.Context, req domain.RegisterVcnNodeReq) error {
	//execOpts := container.ExecOptions{
	//	Cmd: []string{
	//		"headscale", "nodes", "register",
	//		"-k", req.Key,
	//		"-u", req.Username,
	//	},
	//}
	//_, err := s.docker.Exec(ctx, s.containerName, execOpts)
	//if err != nil {
	//	return rerrors.Wrap(err)
	//}

	return nil
}
