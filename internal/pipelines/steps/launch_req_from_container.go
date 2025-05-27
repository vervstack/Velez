package steps

import (
	"context"
	"strings"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/pkg/velez_api"
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

	*s.fromContId = cont.Uuid
	*s.result = domain.LaunchSmerd{
		CreateSmerd_Request: &velez_api.CreateSmerd_Request{
			Name:      cont.Name,
			ImageName: cont.ImageName,
			Settings: &velez_api.Container_Settings{
				Ports:   cont.Ports,
				Network: make([]*velez_api.NetworkBind, 0),
				Volumes: cont.Volumes,
			},
			Env:    cont.Env,
			Labels: cont.Labels,

			// TODO: when MVP of upgrade will work may think about this one
			Hardware:      nil,
			Command:       nil,
			Healthcheck:   nil,
			IgnoreConfig:  false,
			UseImagePorts: false,
			ConfigVersion: nil,
			AutoUpgrade:   false,
		},
	}

	for _, n := range cont.Networks {
		net := &velez_api.NetworkBind{
			NetworkName: n.NetworkName,
			Aliases:     nil,
		}
		for _, a := range n.Aliases {
			if !strings.HasPrefix(cont.Uuid, a) {
				net.Aliases = append(net.Aliases, a)
			}
		}

		if len(net.Aliases) != 0 {
			s.result.Settings.Network = append(s.result.Settings.Network, net)
		}
	}

	return nil
}

func (s *fromContainerToRequest) fromContainerEnv(env []string) map[string]string {
	out := make(map[string]string)
	for _, e := range env {
		nameVal := strings.Split(e, "=")
		if len(nameVal) < 2 {
			continue
		}

		out[nameVal[0]] = strings.Join(nameVal[1:], "")
	}

	return out
}
