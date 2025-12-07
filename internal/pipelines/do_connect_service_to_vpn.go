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
	return ConnectServiceToVpn(req, p.nodeClients,
		p.clusterClients.Vpn(), p.clusterClients.ServiceDiscovery())
}

func ConnectServiceToVpn(req domain.ConnectServiceToVpn,
	nc node_clients.NodeClients,
	vpnClient cluster_clients.VervPrivateNetworkClient,
	sdClient cluster_clients.ServiceDiscovery,
) Runner[any] {
	// region Pipeline context
	launchContainer := patterns.TailScaleContainerSidecar(req.ServiceName)

	var containerId string
	var clientKey string
	var loginServer string
	var namespaceId string

	containerName := req.ServiceName + "-ts-sidecar"
	hostname := strings.ReplaceAll(req.ServiceName+"-ts-sidecar", "_", "-")

	//endregion

	return &runner[any]{
		Steps: []steps.Step{
			network_steps.CheckSidecarExist(nc, containerName),
			network_steps.PrepareNamespace(vpnClient, &req.ServiceName, &namespaceId),
			network_steps.IssueClientKey(vpnClient, &namespaceId, &clientKey),
			network_steps.GetLoginServerUrl(&loginServer),
			steps.SingleFunc(func(_ context.Context) error {
				//TODO Change onto ENV variables
				launchContainer.Env = append(launchContainer.Env,
					"TS_HOSTNAME="+hostname,
					"TS_AUTHKEY="+clientKey,
					"TS_EXTRA_ARGS=--login-server="+loginServer,
				)
				return nil
			}),
			steps.PrepareImage(nc, launchContainer.Image, nil),
			container_steps.Create(
				nc, &launchContainer,
				&containerName, &containerId),
			smerd_steps.Start(nc, &containerId),
			network_steps.AddMakoshRecord(sdClient, req.ServiceName, hostname),
		},
	}
}
