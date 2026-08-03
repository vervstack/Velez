package vpnconnect

import (
	"context"
	"strings"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/patterns"
)

// ConnectServiceToVpn launches the tailscale sidecar for req.ServiceName and
// registers it with service discovery. Moved verbatim from the deleted
// internal/pipelines/do_connect_service_to_vpn.go - see the package doc for
// why it is not a jobs-engine task.
func ConnectServiceToVpn(req domain.ConnectServiceToVcn,
	nc node_clients.NodeClients,
	vpnClient cluster_clients.VervClosedNetworkClient,
	sdClient cluster_clients.ServiceDiscovery,
) Runner[any] {
	// region Pipeline context
	launchContainer := patterns.TailScaleContainerSidecar(req.ServiceName)

	var (
		containerId string
		clientKey   string
		loginServer string
		namespaceId string
	)

	containerName := req.ServiceName + "-" + patterns.TailscaleSidecarSuffix
	hostname := strings.ReplaceAll(containerName, "_", "-")

	// endregion

	appendSidecarEnv := newSingleFunc(func(_ context.Context) error {
		launchContainer.Env = append(launchContainer.Env,
			"TS_HOSTNAME="+hostname,
			"TS_AUTHKEY="+clientKey,
			"TS_EXTRA_ARGS=--login-server="+loginServer,
		)

		return nil
	})

	return &runner[any]{
		steps: []step{
			newCheckSidecarExist(nc, containerName),
			newPrepareNamespace(vpnClient, &req.ServiceName, &namespaceId),
			newGetClientKey(vpnClient, &namespaceId, &clientKey),
			newGetLoginServerUrl(&loginServer),
			appendSidecarEnv,
			newPrepareImage(nc, launchContainer.Image),
			newCreateContainer(
				nc, &launchContainer,
				// VPN sidecar - node-level, no environment suffix.
				&containerName, "", &containerId),
			newStartContainer(nc, &containerId),
			newAddMakoshRecord(sdClient, req.ServiceName, hostname),
		},
	}
}
