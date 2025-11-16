package vpn_manager

import (
	"bytes"
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/backservice/headscale"
	"go.vervstack.ru/Velez/internal/domain"
)

const listNameSpacesCommand = `headscale user list`

func (s *Service) ListNamespaces(ctx context.Context) ([]domain.VpnNamespace, error) {
	exec := container.ExecOptions{
		Cmd: strings.Split(listNameSpacesCommand, " "),
		Env: []string{
			"HEADSCALE_LOG_FORMAT=text",
			"NO_COLOR=1",
		},
	}

	res, err := s.docker.Exec(ctx, headscale.Name, exec)
	if err != nil {
		return nil, rerrors.Wrap(err, "")
	}

	return unmarshalListNamespaces(res), nil
}

func unmarshalListNamespaces(in []byte) []domain.VpnNamespace {
	lineSkip := bytes.IndexByte(in, '\n')
	if lineSkip == -1 {
		return nil
	}

	return nil
}
