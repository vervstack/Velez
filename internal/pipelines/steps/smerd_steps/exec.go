package smerd_steps

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/clients"
)

type execStep struct {
	docker      clients.Docker
	containerId *string
	command     *string
}

func Exec(nodeClients clients.NodeClients, containerId *string, command *string) *execStep {
	return &execStep{
		docker:      nodeClients.Docker(),
		containerId: containerId,
		command:     command,
	}
}

func (s *execStep) Do(ctx context.Context) error {
	ops := container.ExecOptions{
		Cmd:    strings.Split(toolbox.FromPtr(s.command), " "),
		Detach: false,
	}

	res, err := s.docker.Exec(ctx, toolbox.FromPtr(s.containerId), ops)
	if err != nil {
		return rerrors.Wrap(err)
	}

	_ = res

	return nil
}
