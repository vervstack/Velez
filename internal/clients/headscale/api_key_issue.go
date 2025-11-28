package headscale

import (
	"bytes"
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients"
)

var envs = []string{
	"HEADSCALE_LOG_FORMAT=text",
	"NO_COLOR=1",
}

const (
	listApiKeys  = "headscale apikey list"
	issueApiKey  = "headscale apikey create"
	expireApiKey = "headscale apikey delete --prefix "
)

type keyIssuer struct {
	docker        clients.Docker
	containerName string
}

// TODO Make trade safe + safe key in security settings
// issueNewApiKey - creates headscale api key
// because Velez mostly is a stateless app new keys issued at start time
func issueNewApiKey(ctx context.Context, docker clients.Docker, containerName string) (string, error) {
	s := keyIssuer{docker, containerName}

	//err := s.collectGarbage(ctx)
	//if err != nil {
	//	return "", rerrors.Wrap(err, "error collecting old keys garbage")
	//}

	newKey, err := s.issueNewKey(ctx)
	if err != nil {
		return "", rerrors.Wrap(err, "error issuing new key")
	}

	return newKey, nil
}

func (s *keyIssuer) collectGarbage(ctx context.Context) error {
	apiKeys, err := s.listApiKeys(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error listing keys")
	}

	err = s.deleteOldApiKeys(ctx, apiKeys)
	if err != nil {
		return rerrors.Wrap(err, "error expiring keys")
	}

	return nil
}

// listApiKeys returns new  api prefixes
func (s *keyIssuer) listApiKeys(ctx context.Context) ([]string, error) {
	execListApiKeys := container.ExecOptions{
		Cmd: strings.Split(listApiKeys, " "),
		Env: envs,
	}

	res, err := s.docker.Exec(ctx, s.containerName, execListApiKeys)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}
	rows := bytes.Split(res, []byte{'\n'})[1:]

	prefixes := make([]string, 0, len(rows))

	for _, r := range rows {
		if len(r) == 0 {
			continue
		}

		startIdx, endIdx := -1, -1

		for idx, b := range r {
			if startIdx != -1 && endIdx != -1 {
				break
			}

			switch b {
			case '|':
				if startIdx == -1 {
					startIdx = idx + 2
					continue
				}

				endIdx = idx - 1
				break
			case '\n':
				break
			default:
				continue
			}
		}
		if startIdx == -1 && endIdx == -1 {
			return nil, rerrors.New("can't parse output")
		}

		prefixes = append(prefixes, string(r[startIdx:endIdx]))
	}

	return prefixes, nil
}

func (s *keyIssuer) deleteOldApiKeys(ctx context.Context, keyPrefixes []string) error {
	for _, pref := range keyPrefixes {
		command := expireApiKey + pref

		execIssueNewKey := container.ExecOptions{
			Cmd: strings.Split(command, " "),
			Env: envs,
		}

		res, err := s.docker.Exec(ctx, s.containerName, execIssueNewKey)
		if err != nil {
			return rerrors.Wrap(err)
		}
		_ = res

		return nil
	}

	return nil
}

func (s *keyIssuer) issueNewKey(ctx context.Context) (string, error) {
	execIssueNewKey := container.ExecOptions{
		Cmd: strings.Split(issueApiKey, " "),
		Env: envs,
	}

	res, err := s.docker.Exec(ctx, s.containerName, execIssueNewKey)
	if err != nil {
		return "", rerrors.Wrap(err)
	}

	return string(res[1 : len(res)-1]), nil
}
