package velez_api_impl

import (
	"context"
	"strings"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

// registerRoles are the network root and the service container of a register request. RootId is empty when the
// root is the service. PreviousRootId is the current network owner, set only when another member becomes the root.
type registerRoles struct {
	RootId         string
	ServiceId      string
	PreviousRootId string
}

func (r registerRoles) isRootSwitched() bool {
	return r.PreviousRootId != ""
}

// resolveRegisterRoles resolves the root and the service container of a register request: members of the requested
// container's network group when root_container_id or service_container_id name one.
func (impl *Impl) resolveRegisterRoles(
	ctx context.Context, req *velez_api.RegisterContainer_Request,
) (registerRoles, error) {
	roles := registerRoles{ServiceId: req.GetContainerId()}

	if req.GetServiceContainerId() == "" && req.GetRootContainerId() == "" {
		return roles, nil
	}

	listReq := &velez_api.ListContainers_Request{Environment: req.GetEnvironment()}

	list, err := impl.smerdService.ListContainers(ctx, listReq)
	if err != nil {
		return registerRoles{}, rerrors.Wrap(err, "error listing containers")
	}

	refs := registerRoleRefs{
		Requested: req.GetContainerId(),
		Root:      req.GetRootContainerId(),
		Service:   req.GetServiceContainerId(),
	}

	roles, err = selectRegisterRoles(list.GetContainers(), refs)
	if err != nil {
		return registerRoles{}, rerrors.Wrap(err)
	}

	if roles.isRootSwitched() {
		err = validateRecreateFreeRequest(req, user_errors.ErrRegisterRootSwitchUnsupported)
		if err != nil {
			return registerRoles{}, rerrors.Wrap(err)
		}

		return roles, nil
	}

	if roles.RootId == "" {
		return roles, nil
	}

	err = validateRecreateFreeRequest(req, user_errors.ErrRegisterServiceContainerUnsupported)
	if err != nil {
		return registerRoles{}, rerrors.Wrap(err)
	}

	return roles, nil
}

// registerRoleRefs are the container references (id prefixes) of a register request. Requested is the current
// network owner; empty Root and Service keep it as both.
type registerRoleRefs struct {
	Requested string
	Root      string
	Service   string
}

// selectRegisterRoles matches the references against the listed containers by id prefix. A requested container
// unknown to the list is reported later by the task, so the roles stay unchanged.
func selectRegisterRoles(containers []*velez_api.DockerContainer, refs registerRoleRefs) (registerRoles, error) {
	roles := registerRoles{ServiceId: refs.Requested}

	requested := findContainerByIdPrefix(containers, refs.Requested)
	if requested == nil {
		return roles, nil
	}

	root := requested

	if refs.Root != "" {
		root = findContainerByIdPrefix(containers, refs.Root)
		if root == nil || !isGroupMember(requested, root) {
			return registerRoles{}, rerrors.Wrap(user_errors.ErrRegisterRootNotInGroup, refs.Root)
		}
	}

	service := root

	if refs.Service != "" {
		service = findContainerByIdPrefix(containers, refs.Service)
		if service == nil || !isGroupMember(requested, service) {
			return registerRoles{}, rerrors.Wrap(user_errors.ErrRegisterServiceNotInGroup, refs.Service)
		}
	}

	roles = registerRoles{ServiceId: service.GetId()}

	if root.GetId() != service.GetId() {
		roles.RootId = root.GetId()
	}

	if root.GetId() != requested.GetId() {
		roles.PreviousRootId = requested.GetId()
	}

	return roles, nil
}

// isGroupMember reports whether member is the owner of the network group or shares its namespace.
func isGroupMember(owner, member *velez_api.DockerContainer) bool {
	if owner.GetId() == member.GetId() {
		return true
	}

	return member.NetworkOwnerContainerId != nil && member.GetNetworkOwnerContainerId() == owner.GetId()
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

// validateRecreateFreeRequest refuses what a request that only relabels and rejoins containers can't take: the
// patterns and recreate settings apply to a container recreated in place.
func validateRecreateFreeRequest(req *velez_api.RegisterContainer_Request, unsupportedErr error) error {
	isGeneric := req.GetPg() == nil && req.GetRegistry() == nil && req.GetRunner() == nil
	hasRecreateSettings := len(req.GetBindMountLinks()) != 0 ||
		len(req.GetPorts()) != 0 ||
		req.ImageTag != nil ||
		req.GetKeepPortMapping()

	if !isGeneric || hasRecreateSettings {
		return rerrors.Wrap(unsupportedErr)
	}

	return nil
}
