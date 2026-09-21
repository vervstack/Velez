package container_manager

import (
	"context"
	"strings"

	errors "go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/parser"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (c *ContainerManager) ListSmerds(
	ctx context.Context,
	req *velez_api.ListSmerds_Request,
) (*velez_api.ListSmerds_Response, error) {
	if req.GetName() != "" {
		lowered := strings.ToLower(req.GetName())

		// Copies into a fresh request rather than lowercasing *req.Name in
		// place - callers (verv_services.List/Get) pass a pointer straight
		// into their own domain.Service.Name field, and mutating through it
		// silently rewrote the caller's service name to lowercase, breaking
		// every later GetByName lookup keyed on that name.
		req = &velez_api.ListSmerds_Request{
			Limit:       req.Limit,
			Name:        &lowered,
			Id:          req.Id,
			Label:       req.GetLabel(),
			Environment: req.GetEnvironment(),
		}
	}

	runtime, err := c.runtimes.Runtime(ctx, req.GetEnvironment())
	if err != nil {
		return nil, errors.Wrap(err, "error resolving environment")
	}

	cl, err := runtime.ListContainers(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "error listing containers")
	}

	resp := &velez_api.ListSmerds_Response{
		Smerds: make([]*velez_api.Smerd, 0, len(cl)),
	}

	for _, container := range cl {
		if container.Labels[labels.CreatedWithVelezLabel] != labelTrue {
			continue
		}

		if container.Labels[labels.Sidecar] == labelTrue {
			continue
		}

		repo := container.Labels[labels.RepoLabel]

		smerd := &velez_api.Smerd{
			Uuid:      container.ID,
			ImageName: container.Image,

			Status: velez_api.Smerd_Status(velez_api.Smerd_Status_value[container.State]),
			CreatedAt: &timestamppb.Timestamp{
				Seconds: container.Created,
			},

			Labels: container.Labels,
			Repo:   &repo,

			Ports: parser.ToPortsSlice(container.Ports),
		}

		if len(container.Names) != 0 {
			smerd.Name = container.Names[0][1:]
		}

		resp.Smerds = append(resp.Smerds, smerd)
	}

	return resp, nil
}
