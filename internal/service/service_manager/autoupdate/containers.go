package autoupdate

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	labelTrue = "true"
)

func (s *Service) listAppContainers(ctx context.Context) ([]container.Summary, error) {
	listReq := &velez_api.ListSmerds_Request{
		Label: map[string]string{
			labels.CreatedWithVelezLabel: labelTrue,
		},
	}

	conts, err := dockerutils.ListContainers(ctx, s.dockerAPI, listReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	apps := make([]container.Summary, 0, len(conts))

	for _, cont := range conts {
		if isAppContainer(cont) {
			apps = append(apps, cont)
		}
	}

	return apps, nil
}

// isAppContainer keeps plain app services of the default environment - the
// only ones the auto-update worker handles.
func isAppContainer(cont container.Summary) bool {
	if len(cont.Names) == 0 {
		return false
	}

	if cont.Labels[labels.Sidecar] == labelTrue {
		return false
	}

	if cont.Labels[labels.SuffixLabel] != "" {
		return false
	}

	classes := domain.ClassifyService(containerName(cont), "", false)

	return len(classes) == 1 && classes[0] == domain.LabelServiceApp
}

func containerName(cont container.Summary) string {
	return strings.TrimPrefix(cont.Names[0], "/")
}
