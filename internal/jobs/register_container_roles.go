package jobs

import (
	"context"
	"slices"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

// registerRootAccessor is set when the service container is a member of a
// network group other than its root. container_name/container_id of the
// payload then name the service, and the root is recreated as a sidecar.
type registerRootAccessor interface {
	GetRootContainerId() string
	GetRootContainerName() string
}

// registerSidecarNames are the group members the task adopts besides the
// service. serviceName is empty when the root is the service. Single-node mode
// recreates every unregistered sidecar, the service among them; cluster mode
// only binds, so the root joins the bound members and the service leaves them.
func registerSidecarNames(
	plan networkGroupPlan, rootName, serviceName string, isStatefull bool,
) ([]string, error) {
	isRootService := serviceName == "" || serviceName == rootName
	if isRootService {
		return plan.SidecarNames, nil
	}

	isServiceAdoptable := !plan.IsRootRegistered && slices.Contains(plan.SidecarNames, serviceName)
	if !isServiceAdoptable {
		return nil, rerrors.Wrap(user_errors.ErrContainerAlreadyLinked, serviceName)
	}

	if !isStatefull {
		return plan.SidecarNames, nil
	}

	names := make([]string, 0, len(plan.SidecarNames))

	names = append(names, rootName)

	for _, name := range plan.SidecarNames {
		if name != serviceName {
			names = append(names, name)
		}
	}

	return names, nil
}

// registerSidecarGroup is the group recreate_sidecars works on for a register
// task: the network root is the recreated root, not the service container.
type registerSidecarGroup struct {
	payload *velez_api.RegisterContainerTaskPayload
}

func (g registerSidecarGroup) GetEnvironment() string {
	return g.payload.GetEnvironment()
}

func (g registerSidecarGroup) GetServiceName() string {
	return g.payload.GetServiceName()
}

func (g registerSidecarGroup) GetContainerName() string {
	if g.payload.GetRootContainerName() != "" {
		return g.payload.GetRootContainerName()
	}

	return g.payload.GetContainerName()
}

func (g registerSidecarGroup) GetSidecarContainerNames() []string {
	return g.payload.GetSidecarContainerNames()
}

func (g registerSidecarGroup) GetServiceMemberName() string {
	if g.payload.GetRootContainerName() == "" {
		return ""
	}

	return g.payload.GetContainerName()
}

// rootUpgradeRequest is the register request for a network root that is not
// the service: its published ports and volumes stay as they are.
type rootUpgradeRequest struct {
	serviceName string
}

func (r rootUpgradeRequest) GetServiceName() string {
	return r.serviceName
}

func (r rootUpgradeRequest) GetKeepPortMapping() bool {
	return true
}

func (r rootUpgradeRequest) GetPorts() []*velez_api.Port {
	return nil
}

func (r rootUpgradeRequest) GetBindMountLinks() []*bindMountLink {
	return nil
}

// recreateRootAsSidecar recreates the network root as a labelled sidecar of
// the service, which is another member of its group. The service container is
// recreated later by recreate_sidecars, joined to the new root.
func (j *recreateWithLabelsJob) recreateRootAsSidecar(ctx context.Context) error {
	runtime, err := j.runtimes.Runtime(ctx, j.req.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	info, found, err := runtime.InspectAny(ctx, j.root.GetRootContainerId())
	if err != nil {
		return rerrors.Wrap(err, "error inspecting network root")
	}

	if !found || info.Config == nil {
		return rerrors.Wrap(user_errors.ErrRegisterContainerNotFound)
	}

	rootName := j.root.GetRootContainerName()

	upgradeReq := &velez_api.UpgradeSmerd_Request{
		Name:        rootName,
		Image:       info.Config.Image,
		Environment: j.req.GetEnvironment(),
	}

	extraLabels := sidecarLabels(j.req.GetServiceName())

	extraLabels[labels.OnboardedFromLabel] = info.ID

	payload := &velez_api.UpgradeSmerdTaskPayload{
		UpgradeRequest: upgradeReq,
		ExtraLabels:    extraLabels,
		RemovedLabels:  []string{labels.DisplayNameLabel},

		IsSidecarsSkipped:  true,
		IsOldContainerKept: true,
	}

	err = applyRegisterOverrides(payload, info, rootUpgradeRequest{serviceName: j.req.GetServiceName()})
	if err != nil {
		return rerrors.Wrap(err)
	}

	err = j.upgradeAndWait(ctx, rootName, payload)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}
