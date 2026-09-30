package jobs

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/network_owner"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	stepBindSidecars = "bind_sidecars"

	listAllContainersLimitNone = 0
)

type registeredGroupSetter interface {
	SetServiceName(v string)
	SetSidecarContainerNames(v []string)
	SetIsRootRegistered(v bool)
	SetServiceId(v int64)
}

type rootRegisteredAccessor interface {
	GetIsRootRegistered() bool
}

type registeredSidecarsAccessor interface {
	GetSidecarContainerNames() []string
}

// networkGroupPlan is what register_container does with a container and the
// containers sharing its network namespace.
type networkGroupPlan struct {
	// ServiceName is the service the root is already registered under; empty
	// when the root is not registered.
	ServiceName      string
	IsRootRegistered bool
	// SidecarNames are the sharing containers that are not registered yet.
	SidecarNames []string
}

// planNetworkGroup decides which part of a network group still has to be
// registered. boundServices maps a container name to the service its binding
// row points at.
func planNetworkGroup(
	rootLabels map[string]string,
	rootBoundService string,
	sidecars []container.Summary,
	boundServices map[string]string,
) networkGroupPlan {
	plan := networkGroupPlan{
		ServiceName:  rootLabels[labels.VervServiceLabel],
		SidecarNames: make([]string, 0, len(sidecars)),
	}

	if plan.ServiceName == "" {
		plan.ServiceName = rootBoundService
	}

	plan.IsRootRegistered = plan.ServiceName != ""

	for _, sidecar := range sidecars {
		name := summaryContainerName(sidecar)

		isRegistered := isSidecarRegistered(sidecar.Labels, boundServices[name])
		if isRegistered {
			continue
		}

		plan.SidecarNames = append(plan.SidecarNames, name)
	}

	return plan
}

func isSidecarRegistered(containerLabels map[string]string, boundService string) bool {
	_, isSidecarLabelled := containerLabels[labels.Sidecar]
	isLinked := containerLabels[labels.VervServiceLabel] != ""

	return (isSidecarLabelled && isLinked) || boundService != ""
}

func summaryContainerName(summary container.Summary) string {
	if len(summary.Names) == 0 {
		return ""
	}

	return strings.TrimPrefix(summary.Names[0], "/")
}

// boundServiceNames maps container name to the service name of its binding.
func boundServiceNames(
	ctx context.Context,
	bindings storage.ContainerBindingsStorage,
	environment string,
) (map[string]string, error) {
	bound := make(map[string]string)

	if bindings == nil {
		return bound, nil
	}

	list, err := bindings.ListByNode(ctx, domain.SelfNodeId, environment)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing container bindings")
	}

	for _, binding := range list {
		bound[binding.ContainerName] = binding.ServiceName
	}

	return bound, nil
}

// planRegisterGroup lists the runtime's containers and plans the group the
// inspected root belongs to.
func planRegisterGroup(
	ctx context.Context,
	runtime container_runtime.ContainerRuntime,
	rootInfo container.InspectResponse,
	boundServices map[string]string,
) (networkGroupPlan, error) {
	list, err := runtime.ListAllContainers(ctx, listAllContainersLimitNone)
	if err != nil {
		return networkGroupPlan{}, rerrors.Wrap(err, "error listing containers")
	}

	sidecars := network_owner.SidecarsOf(list, rootInfo.ID)
	rootName := strings.TrimPrefix(rootInfo.Name, "/")

	return planNetworkGroup(rootInfo.Config.Labels, boundServices[rootName], sidecars, boundServices), nil
}

// applyGroupPlan records the plan on the task payload. A group whose root is
// already registered adopts only its sidecars, under the root's own service.
func applyGroupPlan(
	ctx context.Context,
	dataStorage storage.Storage,
	plan networkGroupPlan,
	payload registeredGroupSetter,
) error {
	if plan.IsRootRegistered && len(plan.SidecarNames) == 0 {
		return rerrors.Wrap(user_errors.ErrContainerAlreadyLinked)
	}

	payload.SetSidecarContainerNames(plan.SidecarNames)

	if !plan.IsRootRegistered {
		return nil
	}

	svc, err := dataStorage.Services().GetByName(ctx, plan.ServiceName)
	if err != nil {
		return rerrors.Wrap(err, "error getting service of the registered root")
	}

	payload.SetServiceName(plan.ServiceName)
	payload.SetServiceId(svc.ID)
	payload.SetIsRootRegistered(true)

	return nil
}

// unlessRootRegistered skips a step that only concerns the root container
// when the root is already registered and only its sidecars are adopted.
type unlessRootRegistered struct {
	inner Job
	group rootRegisteredAccessor
}

func (j *unlessRootRegistered) Do(ctx context.Context) error {
	if j.group.GetIsRootRegistered() {
		return nil
	}

	err := j.inner.Do(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

func (j *unlessRootRegistered) Rollback(ctx context.Context) error {
	if j.group.GetIsRootRegistered() {
		return nil
	}

	rollbackable, ok := j.inner.(RollbackableJob)
	if !ok {
		return nil
	}

	err := rollbackable.Rollback(ctx)
	if err != nil {
		return rerrors.Wrap(err)
	}

	return nil
}

// bindSidecarsJob binds every sidecar of the group to the service in cluster
// mode, leaving the containers untouched.
type bindSidecarsJob struct {
	dataStorage storage.Storage

	req      registerRequestAccessor
	service  registeredServiceAccessor
	sidecars registeredSidecarsAccessor
}

func (j *bindSidecarsJob) Do(ctx context.Context) error {
	names := j.sidecars.GetSidecarContainerNames()
	if len(names) == 0 {
		return nil
	}

	bindings := j.dataStorage.ContainerBindings()
	if bindings == nil {
		return rerrors.Wrap(user_errors.ErrContainerBindingsUnavailable)
	}

	for _, name := range names {
		binding := domain.ContainerBinding{
			ServiceId:     j.service.GetServiceId(),
			ServiceName:   j.req.GetServiceName(),
			NodeId:        domain.SelfNodeId,
			Environment:   j.req.GetEnvironment(),
			ContainerName: name,
			IsSidecar:     true,
		}

		err := bindings.Upsert(ctx, binding)
		if err != nil {
			return rerrors.Wrap(err, "error upserting sidecar binding")
		}
	}

	return nil
}
