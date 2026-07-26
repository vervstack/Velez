package steps

import (
	"context"
	"strings"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service"
)

type fromContainerToRequest struct {
	containerService service.ContainerService

	containerName string

	result     *domain.LaunchSmerd
	fromContId *string
}

func FromContainerToRequest(
	containerName string,
	containerService service.ContainerService,

	result *domain.LaunchSmerd,
	fromContId *string,
) *fromContainerToRequest {
	return &fromContainerToRequest{
		containerName:    containerName,
		containerService: containerService,

		result:     result,
		fromContId: fromContId,
	}
}

func (s *fromContainerToRequest) Do(ctx context.Context) error {
	cont, err := s.containerService.InspectSmerd(ctx, s.containerName)
	if err != nil {
		return rerrors.Wrap(err, "error inspecting container")
	}

	// base
	*s.fromContId = cont.GetUuid()
	*s.result = domain.LaunchSmerd{
		CreateSmerd_Request: &velez_api.CreateSmerd_Request{
			Name:      cont.GetName(),
			ImageName: cont.GetImageName(),
			Settings: &velez_api.Container_Settings{
				Ports:   cont.GetPorts(),
				Network: s.fromContainerNetwork(cont),
				Volumes: cont.GetVolumes(),
			},
			Env:    cont.GetEnv(),
			Labels: cont.GetLabels(),

			// TODO: when MVP of upgrade will work may think about this one
			Hardware:      nil,
			Command:       nil,
			Healthcheck:   nil,
			IgnoreConfig:  false,
			UseImagePorts: false,
			AutoUpgrade:   false,
		},
	}

	return nil
}

func (s *fromContainerToRequest) fromContainerNetwork(cont *velez_api.Smerd) []*velez_api.NetworkBind {
	out := make([]*velez_api.NetworkBind, 0, len(cont.GetNetworks()))

	for _, n := range cont.GetNetworks() {
		net := &velez_api.NetworkBind{
			NetworkName: n.GetNetworkName(),
			Aliases:     nil,
		}
		for _, a := range n.GetAliases() {
			if !strings.HasPrefix(cont.GetUuid(), a) {
				net.Aliases = append(net.Aliases, a)
			}
		}

		if len(net.GetAliases()) != 0 {
			out = append(out, net)
		}
	}

	return out
}
