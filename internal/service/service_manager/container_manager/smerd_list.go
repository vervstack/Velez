package container_manager

import (
	"context"

	errors "go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients/docker/dockerutils/parser"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ListSmerds filters on req.GetName() exactly as given - the real Docker
// container name (and VervServiceLabel) preserves the case a service was
// created with, and Docker's name filter is case-sensitive. Lowercasing the
// filter here used to break every lookup for a name with an uppercase
// letter, including List's enrichment and Drop's smerd lookup, which
// silently no-oped (0 smerds found) instead of dropping the container. See
// CLAUDE.md's "Known pitfalls" entry - matching by the case-preserving
// VervServiceLabel, not this Docker Name filter, avoided the same trap in
// services.go's GetByName.
func (c *ContainerManager) ListSmerds(
	ctx context.Context,
	req *velez_api.ListSmerds_Request,
) (*velez_api.ListSmerds_Response, error) {
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
