package jobs

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
	"go.vervstack.ru/Velez/internal/clients/node_clients/runtime_policy"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage"
)

type whitelistAction int

const (
	whitelistActionNone whitelistAction = iota
	whitelistActionRefuse
	whitelistActionAddRepository
)

// decideWhitelistAction picks what registering a container needs from the image whitelist: nothing when the
// container is not elevated or already allowed, a refusal when the user did not allow elevated access, and a
// repository row in cluster mode (single-node mode keeps the allowance as a container label instead).
func decideWhitelistAction(isElevatedAllowed, isAccessRequested, isCluster bool) whitelistAction {
	if isElevatedAllowed {
		return whitelistActionNone
	}

	if !isAccessRequested {
		return whitelistActionRefuse
	}

	if isCluster {
		return whitelistActionAddRepository
	}

	return whitelistActionNone
}

// stampElevatedAccessAllowance marks the recreated container as allowed to keep elevated access. Cluster mode
// records the allowance as a whitelist row instead.
func stampElevatedAccessAllowance(extraLabels map[string]string, isAccessRequested, isCluster bool) {
	if !isAccessRequested || isCluster {
		return
	}

	extraLabels[labels.ElevatedAccessLabel] = labels.ElevatedAccessAllowedValue
}

// whitelistImageJob refuses a register whose container needs elevated access nobody allowed, before anything on
// the node is changed, and whitelists the image repository in cluster mode when the user allowed it.
type whitelistImageJob struct {
	dataStorage storage.Storage
	runtimes    container_runtime.RuntimeResolver

	req       registerRequestAccessor
	root      registerRootAccessor
	container registeredContainerAccessor

	addedRepository string
	isAdded         bool
}

func (j *whitelistImageJob) Do(ctx context.Context) error {
	runtime, err := j.runtimes.Runtime(ctx, j.req.GetEnvironment())
	if err != nil {
		return rerrors.Wrap(err, "error resolving container runtime")
	}

	isCluster := j.dataStorage.IsStatefull()

	targetId := j.req.GetContainerId()
	if !isCluster && j.root.GetRootContainerId() != "" {
		targetId = j.root.GetRootContainerId()
	}

	info, found, err := runtime.InspectAny(ctx, targetId)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container")
	}

	if !found || info.Config == nil {
		return rerrors.Wrap(user_errors.ErrRegisterContainerNotFound)
	}

	image := info.Config.Image

	if targetId == j.req.GetContainerId() {
		image, err = j.serviceImage()
		if err != nil {
			return rerrors.Wrap(err)
		}
	}

	settings, err := j.dataStorage.Settings().GetSettings(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error getting settings")
	}

	isAllowed := runtime_policy.IsElevatedAllowed(info.HostConfig, image, settings, info.Config.Labels)

	action := decideWhitelistAction(isAllowed, j.req.GetIsElevatedAccessAllowed(), isCluster)

	switch action {
	case whitelistActionNone:
		return nil
	case whitelistActionRefuse:
		return rerrors.Wrap(user_errors.ErrElevatedAccessImageNotWhitelisted, "error checking image "+image)
	case whitelistActionAddRepository:
		return j.addRepository(ctx, image)
	}

	return nil
}

func (j *whitelistImageJob) Rollback(ctx context.Context) error {
	if !j.isAdded {
		return nil
	}

	err := j.dataStorage.Whitelist().RemoveByRepository(ctx, j.addedRepository)
	if err != nil {
		return rerrors.Wrap(err, "error removing whitelisted image repository")
	}

	return nil
}

func (j *whitelistImageJob) serviceImage() (string, error) {
	image := j.container.GetImageName()

	if j.req.GetImageTag() == "" {
		return image, nil
	}

	tagged, err := imageWithTag(image, j.req.GetImageTag())
	if err != nil {
		return "", rerrors.Wrap(err)
	}

	return tagged, nil
}

func (j *whitelistImageJob) addRepository(ctx context.Context, image string) error {
	repository := runtime_policy.RepositoryOf(image)

	_, isAdded, err := j.dataStorage.Whitelist().Add(ctx, repository)
	if err != nil {
		return rerrors.Wrap(err, "error whitelisting image repository")
	}

	j.isAdded = isAdded
	j.addedRepository = repository

	return nil
}
