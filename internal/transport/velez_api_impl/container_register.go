package velez_api_impl

import (
	"context"

	"github.com/google/uuid"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/user_errors"
)

func (impl *Impl) RegisterContainer(
	ctx context.Context,
	req *velez_api.RegisterContainer_Request,
) (*velez_api.RegisterContainer_Response, error) {
	if req.GetContainerId() == "" {
		return nil, rerrors.Wrap(user_errors.ErrRegisterContainerIdRequired)
	}

	if req.GetServiceName() == "" {
		return nil, rerrors.Wrap(user_errors.ErrRegisterServiceNameRequired)
	}

	if req.GetKeepPortMapping() && len(req.GetPorts()) != 0 {
		return nil, rerrors.Wrap(user_errors.ErrRegisterPortsWithKeepMapping)
	}

	_, err := impl.resolveEnvironment(ctx, req.GetEnvironment())
	if err != nil {
		return nil, err
	}

	isGeneric := req.GetPattern() == nil || req.GetGeneric() != nil
	if !isGeneric {
		return nil, rerrors.Wrap(errPatternNotImplemented)
	}

	payload := &velez_api.RegisterContainerTaskPayload{
		ContainerId: req.GetContainerId(),
		Environment: req.GetEnvironment(),
		ServiceName: req.GetServiceName(),

		BindMountLinks:  req.GetBindMountLinks(),
		KeepPortMapping: req.GetKeepPortMapping(),
		Ports:           req.GetPorts(),
	}

	entityId := req.GetServiceName() + "/" + uuid.NewString()

	_, err = impl.jobsEngine.Enqueue(ctx, entityId, jobs.RegisterContainerAction, payload)
	if err != nil {
		return nil, rerrors.Wrap(err, "error enqueuing register_container task")
	}

	resp := &velez_api.RegisterContainer_Response{
		EntityId: entityId,
		Action:   jobs.RegisterContainerAction,
	}

	return resp, nil
}
