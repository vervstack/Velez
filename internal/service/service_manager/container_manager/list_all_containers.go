package container_manager

import (
	"context"

	errors "go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/network_owner"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/parser"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListContainers lists every container on the daemon backing
// req.Environment, registered with Velez or not - unlike ListSmerds, there
// is no CreatedWithVelezLabel/Sidecar filtering, matching
// ContainerRuntime.ListAllContainers (see its doc comment for why that's
// daemon-wide by design). req.GetFilters() are applied AND-combined before
// a container is added to the response.
func (c *ContainerManager) ListContainers(
	ctx context.Context,
	req *velez_api.ListContainers_Request,
) (*velez_api.ListContainers_Response, error) {
	runtime, err := c.runtimes.Runtime(ctx, req.GetEnvironment())
	if err != nil {
		return nil, errors.Wrap(err, "error resolving environment")
	}

	cl, err := runtime.ListAllContainers(ctx, req.GetLimit())
	if err != nil {
		return nil, errors.Wrap(err, "error listing containers")
	}

	boundServices, err := c.boundServiceNames(ctx, req.GetEnvironment())
	if err != nil {
		return nil, errors.Wrap(err, "error resolving container bindings")
	}

	resp := &velez_api.ListContainers_Response{
		Containers: make([]*velez_api.DockerContainer, 0, len(cl)),
	}

	owners := network_owner.OwnerIds(cl)
	replaced := replacedContainerIds(cl)

	for _, cont := range cl {
		dc := &velez_api.DockerContainer{
			Id:           cont.ID,
			ImageName:    cont.Image,
			Status:       velez_api.Smerd_Status(velez_api.Smerd_Status_value[cont.State]),
			CreatedAt:    &timestamppb.Timestamp{Seconds: cont.Created},
			Labels:       cont.Labels,
			Ports:        parser.ToPortsSlice(cont.Ports),
			Mounts:       toMounts(cont.Mounts),
			IsRegistered: cont.Labels[labels.CreatedWithVelezLabel] == labelTrue,

			SuggestedPattern: suggestedPattern(cont.Image),
		}

		if domain.IsVelezImage(cont.Image) {
			dc.IsRegistered = true
		}

		if len(cont.Names) != 0 {
			dc.Name = cont.Names[0][1:]
		}

		if cont.NetworkSettings != nil {
			dc.Networks = toNetworkBinds(cont.ID, cont.NetworkSettings.Networks)
		}

		ownerId, isShared := owners[cont.ID]
		if isShared {
			dc.NetworkOwnerContainerId = &ownerId
		}

		replacedBy, isReplaced := replaced[cont.ID]
		if isReplaced {
			dc.ReplacedByContainerId = &replacedBy
		}

		svc, isLabelled := cont.Labels[labels.VervServiceLabel]
		if isLabelled {
			dc.LinkedServiceName = &svc
		}

		boundSvc, isBound := boundServices[dc.GetName()]
		if !isLabelled && isBound {
			dc.LinkedServiceName = &boundSvc
			dc.IsRegistered = true
		}

		if !matchesContainerFilters(dc, req.GetFilters()) {
			continue
		}

		resp.Containers = append(resp.Containers, dc)
	}

	return resp, nil
}
