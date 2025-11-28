package pipelines

import (
	"context"
	"strings"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/patterns"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/container_steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/network_steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/smerd_steps"
)

func (p *pipeliner) ConnectServiceToVpn(req domain.ConnectServiceToVpn) Runner[any] {
	// region Pipeline context
	launchContainer := patterns.TailScaleContainerSidecar(req.ServiceName)

	var containerId string
	var clientKey string
	var sidecarCommand string
	var loginServer string

	containerName := req.ServiceName + "-ts-sidecar"

	//endregion

	return &runner[any]{
		Steps: []steps.Step{
			network_steps.PreCheck(p.services, containerName),
			network_steps.IssueClientKey(p.clusterClients, req.NamespaceId, &clientKey),
			network_steps.GetLoginServerUrl(&loginServer),
			steps.SingleFunc(func(_ context.Context) error {
				hostname := strings.ReplaceAll(req.ServiceName+"-ts-sidecar", "_", "-")
				sidecarCommand =
					"tailscale up --authkey=" + clientKey +
						" --hostname=" + hostname +
						" --accept-routes --advertise-exit-node" +
						" --login-server=" + loginServer
				return nil
			}),
			steps.PrepareImage(p.nodeClients, launchContainer.Image, nil),
			container_steps.Create(p.nodeClients, &launchContainer,
				&containerName, &containerId),
			smerd_steps.Start(p.nodeClients, &containerId),
			smerd_steps.Exec(p.nodeClients, &containerName, &sidecarCommand),
		},
	}
}
