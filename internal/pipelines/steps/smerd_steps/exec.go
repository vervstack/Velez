package smerd_steps

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
)

type execStep struct {
	docker      node_clients.Docker
	containerID *string
	command     *string
}

func Exec(nodeClients node_clients.NodeClients, containerID *string, command *string) *execStep {
	return &execStep{
		docker:      nodeClients.Docker(),
		containerID: containerID,
		command:     command,
	}
}

func (s *execStep) Do(ctx context.Context) error {
	ops := container.ExecOptions{
		Cmd:    strings.Split(toolbox.FromPtr(s.command), " "),
		Detach: false,
	}

	res, err := s.docker.Exec(ctx, toolbox.FromPtr(s.containerID), ops)
	if err != nil {
		return rerrors.Wrap(err)
	}

	_ = res

	return nil
}
