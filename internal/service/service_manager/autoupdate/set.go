package autoupdate

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

// SetService recreates the service through upgrade_smerd with the new labels:
// Docker can neither change labels of a running container nor remove them, so
// disabling overrides the enabled label with "false".
func (s *AutoUpdateServiceImpl) SetService(
	ctx context.Context, req *velez_api.SetServiceAutoUpdate_Request,
) (*velez_api.SetServiceAutoUpdate_Response, error) {
	name := req.GetName()
	if name == "" {
		return nil, rerrors.Wrap(user_errors.ErrAutoUpdateServiceNameRequired)
	}

	conts, err := s.listAppContainers(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing app services")
	}

	image := ""
	isFound := false

	for _, cont := range conts {
		if containerName(cont) == name {
			image = cont.Image
			isFound = true

			break
		}
	}

	if !isFound {
		return nil, rerrors.Wrap(user_errors.ErrAutoUpdateServiceNotFound)
	}

	upgradeReq := &velez_api.UpgradeSmerd_Request{
		Name:  name,
		Image: image,
	}

	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: upgradeReq,
		ExtraLabels:    autoUpdateLabels(req),
	}

	_, err = s.jobsEngine.Enqueue(ctx, name, jobs.UpgradeSmerdAction, payload)
	if err != nil {
		return nil, rerrors.Wrap(err, "error enqueuing upgrade_smerd task")
	}

	resp := &velez_api.SetServiceAutoUpdate_Response{
		EntityId: name,
		Action:   jobs.UpgradeSmerdAction,
	}

	return resp, nil
}

func autoUpdateLabels(req *velez_api.SetServiceAutoUpdate_Request) map[string]string {
	if !req.GetIsEnabled() {
		return map[string]string{
			labels.AutoUpgrade: "false",
		}
	}

	return map[string]string{
		labels.AutoUpgrade:          labelTrue,
		labels.AutoUpgradeBumpLevel: domain.BumpLevelLabelValue(req.GetBumpLevel()),
	}
}
