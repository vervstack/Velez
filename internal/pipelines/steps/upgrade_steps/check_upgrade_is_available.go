package upgrade_steps

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"
	"google.golang.org/grpc/codes"

	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
	"go.vervstack.ru/Velez/internal/service"
)

var ErrSelfUpgradeIsForbidden = rerrors.NewUserError("Can't perform self upgrade", codes.FailedPrecondition)

type checkUpgradeIsAvailableStep struct {
	smerdService service.ContainerService

	smerdName *string
}

func CheckUpgradeIsAvailable(
	services service.Services,
	smerdName *string,
) steps.Step {
	return &checkUpgradeIsAvailableStep{
		smerdService: services.SmerdManager(),
		smerdName:    smerdName,
	}
}

func (s *checkUpgradeIsAvailableStep) Do(ctx context.Context) error {
	id := env.GetContainerId()
	if id != nil {
		smerd, err := s.smerdService.InspectSmerd(ctx, toolbox.FromPtr(s.smerdName))
		if err != nil {
			return rerrors.Wrap(err)
		}
		if smerd.Uuid == *id {
			return ErrSelfUpgradeIsForbidden
		}
	}

	return nil
}
