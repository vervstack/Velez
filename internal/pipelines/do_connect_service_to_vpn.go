package pipelines

import (
	"context"
	"strings"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/patterns"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/container_steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/network_steps"
	"go.vervstack.ru/Velez/internal/pipelines/steps/smerd_steps"
)

func (p *pipeliner) ConnectServiceToVpn(req domain.ConnectServiceToVpn) Runner[any] {
	return ConnectServiceToVpn(req, p.nodeClients, p.clusterClients.Vpn())
}

func ConnectServiceToVpn(req domain.ConnectServiceToVpn,
	nc node_clients.NodeClients,
	vpnClient cluster_clients.VervPrivateNetworkClient) Runner[any] {
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
			network_steps.PreCheck(nc, containerName),
			network_steps.IssueClientKey(vpnClient, req.NamespaceId, &clientKey),
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
			steps.PrepareImage(nc, launchContainer.Image, nil),
			container_steps.Create(nc,
				&launchContainer,
				&containerName, &containerId),
			smerd_steps.Start(nc, &containerId),
			smerd_steps.Exec(nc, &containerName, &sidecarCommand),
		},
	}
}
