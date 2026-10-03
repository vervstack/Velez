package container_manager

import (
	"context"

	errors "go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/user_errors"
)

func (c *ContainerManager) FinishOnboarding(
	ctx context.Context,
	req *velez_api.FinishOnboarding_Request,
) (*velez_api.FinishOnboarding_Response, error) {
	runtime, err := c.runtimes.Runtime(ctx, req.GetEnvironment())
	if err != nil {
		return nil, errors.Wrap(err, "error resolving environment")
	}

	info, found, err := runtime.InspectAny(ctx, req.GetContainerId())
	if err != nil {
		return nil, errors.Wrap(err, "error inspecting container")
	}

	if !found {
		return nil, errors.Wrap(user_errors.ErrRegisterContainerNotFound)
	}

	list, err := runtime.ListAllContainers(ctx, 0)
	if err != nil {
		return nil, errors.Wrap(err, "error listing containers")
	}

	_, isReplaced := replacedContainerIds(list)[info.ID]
	if !isReplaced {
		return nil, errors.Wrap(user_errors.ErrContainerNotOnboardingLeftover)
	}

	err = runtime.Remove(ctx, info.ID)
	if err != nil {
		return nil, errors.Wrap(err, "error removing replaced container")
	}

	resp := &velez_api.FinishOnboarding_Response{}

	return resp, nil
}
