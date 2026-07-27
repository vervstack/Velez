package headscale

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
)

var envs = []string{
	"HEADSCALE_LOG_FORMAT=text",
	"NO_COLOR=1",
}

const (
	listAPIKeys  = "headscale apikey list"
	issueAPIKey  = "headscale apikey create"
	expireAPIKey = "headscale apikey delete --prefix "
)

type keyIssuer struct {
	docker        node_clients.Docker
	containerName string
}

// TODO Make trade safe + safe key in security settings
// issueNewAPIKey - creates headscale api key
// because Velez mostly is a stateless app new keys issued at start time.
func issueNewAPIKey(ctx context.Context, docker node_clients.Docker, containerName string) (string, error) {
	s := keyIssuer{docker, containerName}

	// err := s.collectGarbage(ctx)
	// if err != nil {
	//	return "", rerrors.Wrap(err, "error collecting old keys garbage")
	// }

	newKey, err := s.issueNewKey(ctx)
	if err != nil {
		return "", rerrors.Wrap(err, "error issuing new key")
	}

	return newKey, nil
}

func (s *keyIssuer) issueNewKey(ctx context.Context) (string, error) {
	execIssueNewKey := container.ExecOptions{
		Cmd:          strings.Split(issueAPIKey, " "),
		Env:          envs,
		AttachStdout: true,
	}

	res, err := s.docker.Exec(ctx, s.containerName, execIssueNewKey)
	if err != nil {
		return "", rerrors.Wrap(err, "error calling exec on container")
	}

	if len(res) == 0 {
		return "", rerrors.New("can't parse output")
	}

	return string(res[1 : len(res)-1]), nil
}
