package velez_api_impl

import (
	"context"
	"strings"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

// registerRoles are the network root and the service container of a register request. RootId is empty when the
// root is the service.
type registerRoles struct {
	RootId    string
	ServiceId string
}

// resolveRegisterRoles resolves the service container of a register request, which is another member of the
// root's network group when service_container_id names one.
func (impl *Impl) resolveRegisterRoles(
	ctx context.Context, req *velez_api.RegisterContainer_Request,
) (registerRoles, error) {
	roles := registerRoles{ServiceId: req.GetContainerId()}

	if req.GetServiceContainerId() == "" {
		return roles, nil
	}

	listReq := &velez_api.ListContainers_Request{Environment: req.GetEnvironment()}

	list, err := impl.smerdService.ListContainers(ctx, listReq)
	if err != nil {
		return registerRoles{}, rerrors.Wrap(err, "error listing containers")
	}

	roles, err = selectRegisterRoles(list.GetContainers(), req.GetContainerId(), req.GetServiceContainerId())
	if err != nil {
		return registerRoles{}, rerrors.Wrap(err)
	}

	if roles.RootId == "" {
		return roles, nil
	}

	err = validateServiceContainerRequest(req)
	if err != nil {
		return registerRoles{}, rerrors.Wrap(err)
	}

	return roles, nil
}

// selectRegisterRoles matches the requested container and the service container by id prefix. A service container
// that is the root itself, or a root unknown to the list (reported later by the task), keeps the roles unchanged.
func selectRegisterRoles(
	containers []*velez_api.DockerContainer, rootRef, serviceRef string,
) (registerRoles, error) {
	roles := registerRoles{ServiceId: rootRef}

	root := findContainerByIdPrefix(containers, rootRef)
	if root == nil {
		return roles, nil
	}

	service := findContainerByIdPrefix(containers, serviceRef)
	if service == nil {
		return registerRoles{}, rerrors.Wrap(user_errors.ErrRegisterServiceNotInGroup, serviceRef)
	}

	if service.GetId() == root.GetId() {
		return roles, nil
	}

	if service.NetworkOwnerContainerId == nil || service.GetNetworkOwnerContainerId() != root.GetId() {
		return registerRoles{}, rerrors.Wrap(user_errors.ErrRegisterServiceNotInGroup, serviceRef)
	}

	return registerRoles{RootId: root.GetId(), ServiceId: service.GetId()}, nil
}

func findContainerByIdPrefix(containers []*velez_api.DockerContainer, idPrefix string) *velez_api.DockerContainer {
	if idPrefix == "" {
		return nil
	}

	for _, cont := range containers {
		if strings.HasPrefix(cont.GetId(), idPrefix) {
			return cont
		}
	}

	return nil
}

// validateServiceContainerRequest refuses what a service container other than the network root can't take: it is
// only labelled and bound, never recreated, so the patterns and recreate settings don't apply.
func validateServiceContainerRequest(req *velez_api.RegisterContainer_Request) error {
	isGeneric := req.GetPg() == nil && req.GetRegistry() == nil && req.GetRunner() == nil
	hasRecreateSettings := len(req.GetBindMountLinks()) != 0 ||
		len(req.GetPorts()) != 0 ||
		req.ImageTag != nil ||
		req.GetKeepPortMapping()

	if !isGeneric || hasRecreateSettings {
		return rerrors.Wrap(user_errors.ErrRegisterServiceContainerUnsupported)
	}

	return nil
}
