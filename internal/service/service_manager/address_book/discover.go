package address_book

import (
	"net"
	"strings"

	"github.com/docker/docker/api/types/container"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/patterns"
)

type serviceAddresses []domain.ServiceAddress

type webUiDiscovery struct {
	host          string
	containerPort uint16
	publishedPort uint32
}

// discoverWebUis locates each web ui. A legacy web ui publishes its port on its
// own container; a sidecar shares its root's network namespace and cannot
// publish anything, so the root container carries it. A web ui without a
// published port is still reported (publishedPort 0): the VCN reaches it
// through the container port.
func discoverWebUis(containers []container.Summary, host string) map[string]webUiDiscovery {
	roots := serviceRoots(containers)
	discovered := make(map[string]webUiDiscovery)

	for _, cont := range containers {
		target, ok := WebUiTargetOf(cont.Labels)
		if !ok {
			continue
		}

		source := cont

		_, isSidecar := cont.Labels[labels.Sidecar]
		if isSidecar {
			root, hasRoot := roots[target.RootService]
			if hasRoot {
				source = root
			}
		}

		found := webUiDiscovery{
			host:          host,
			containerPort: target.ContainerPort,
			publishedPort: publishedPort(source, target.ContainerPort),
		}

		known, isKnown := discovered[target.RootService]
		if isKnown && known.publishedPort != 0 && found.publishedPort == 0 {
			continue
		}

		discovered[target.RootService] = found
	}

	return discovered
}

func vcnHostname(serviceName string) string {
	containerName := serviceName + "-" + patterns.TailscaleSidecarSuffix

	return strings.ReplaceAll(containerName, "_", "-")
}

func vcnIpOf(serviceName string, nodes []domain.VcnNode) (string, bool) {
	hostname := vcnHostname(serviceName)

	for _, node := range nodes {
		if node.Name != hostname {
			continue
		}

		for _, rawIp := range node.IpAddresses {
			ip := net.ParseIP(rawIp)
			if ip != nil && ip.To4() != nil {
				return ip.String(), true
			}
		}
	}

	return "", false
}

func buildAddresses(serviceName string, found webUiDiscovery, nodes []domain.VcnNode) serviceAddresses {
	var addresses serviceAddresses

	if found.publishedPort != 0 {
		address := domain.ServiceAddress{
			ServiceName: serviceName,
			Name:        domain.WebUiAddressName,
			Host:        found.host,
			Port:        found.publishedPort,
			Scope:       velez_api.AddressScope_ADDRESS_SCOPE_DOCKER,
		}

		addresses = append(addresses, address)
	}

	ip, hasIp := vcnIpOf(serviceName, nodes)
	if hasIp {
		address := domain.ServiceAddress{
			ServiceName: serviceName,
			Name:        domain.WebUiAddressName,
			Host:        ip,
			Port:        uint32(found.containerPort),
			Scope:       velez_api.AddressScope_ADDRESS_SCOPE_VCN,
		}

		addresses = append(addresses, address)
	}

	return addresses
}
