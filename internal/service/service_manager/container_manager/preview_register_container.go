package container_manager

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/network_owner"
	"go.vervstack.ru/Velez/internal/domain/register_preview"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

func (c *ContainerServiceImpl) PreviewRegisterContainer(
	ctx context.Context,
	req *velez_api.PreviewRegisterContainer_Request,
) (*velez_api.PreviewRegisterContainer_Response, error) {
	runtime, err := c.runtimes.Runtime(ctx, req.GetEnvironment())
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving environment")
	}

	list, err := runtime.ListAllContainers(ctx, 0)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	requested, isFound := findSummaryByPrefix(list, req.GetContainerId())
	if !isFound {
		return nil, rerrors.Wrap(user_errors.ErrRegisterContainerNotFound)
	}

	groupSummaries := append([]container.Summary{requested}, network_owner.SidecarsOf(list, requested.ID)...)

	members := make([]container.InspectResponse, 0, len(groupSummaries))

	for _, summary := range groupSummaries {
		info, isInspected, inspectErr := runtime.InspectAny(ctx, summary.ID)
		if inspectErr != nil {
			return nil, rerrors.Wrap(inspectErr, "error inspecting container")
		}

		if !isInspected {
			return nil, rerrors.Wrap(user_errors.ErrRegisterContainerNotFound)
		}

		members = append(members, info)
	}

	rootId, err := resolveMemberId(members, req.GetRootContainerId(), requested.ID)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	serviceId, err := resolveMemberId(members, req.GetServiceContainerId(), rootId)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	previousRootId := ""
	if rootId != requested.ID {
		previousRootId = requested.ID
	}

	input := register_preview.RegisterPreviewInput{
		Members:     members,
		ServiceName: req.GetServiceName(),
		RootId:      rootId,
		ServiceId:   serviceId,

		PreviousRootId: previousRootId,
		Pattern:        previewPattern(req),
	}

	previews, err := register_preview.BuildRegisterPreview(input)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	resp := &velez_api.PreviewRegisterContainer_Response{Containers: previews}

	return resp, nil
}

func findSummaryByPrefix(list []container.Summary, id string) (container.Summary, bool) {
	if id == "" {
		return container.Summary{}, false
	}

	for _, summary := range list {
		if strings.HasPrefix(summary.ID, id) {
			return summary, true
		}
	}

	return container.Summary{}, false
}

func resolveMemberId(members []container.InspectResponse, requested, fallback string) (string, error) {
	if requested == "" {
		return fallback, nil
	}

	for _, member := range members {
		if strings.HasPrefix(member.ID, requested) {
			return member.ID, nil
		}
	}

	return "", rerrors.Wrap(user_errors.ErrPreviewRoleNotInGroup)
}

func previewPattern(req *velez_api.PreviewRegisterContainer_Request) velez_api.ServicePattern {
	switch {
	case req.GetPg() != nil:
		return velez_api.ServicePattern_SERVICE_PATTERN_POSTGRES
	case req.GetRegistry() != nil:
		return velez_api.ServicePattern_SERVICE_PATTERN_REGISTRY
	case req.GetRunner() != nil:
		if req.GetRunner().GetProvider() == velez_api.RunnerProvider_GITLAB {
			return velez_api.ServicePattern_SERVICE_PATTERN_GITLAB_RUNNER
		}

		return velez_api.ServicePattern_SERVICE_PATTERN_GITHUB_RUNNER
	default:
		return velez_api.ServicePattern_SERVICE_PATTERN_UNSPECIFIED
	}
}
