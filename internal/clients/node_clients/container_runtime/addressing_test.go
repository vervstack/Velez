package container_runtime

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/user_errors"
)

const (
	testGatewayIp  = "172.18.0.1"
	testLocalhost  = "localhost"
	testGarageIpA  = "10.0.0.2"
	testGarageName = "/garage"
	testUnixHost   = "unix:///var/run/docker.sock"
)

func Test_ParseDialHost_Cases(t *testing.T) {
	cases := []struct {
		name          string
		daemonHost    string
		isInContainer bool
		gatewayIp     string
		want          string
	}{
		{"tcp host", "tcp://sc:2375", false, "", "sc"},
		{"tcp host in container ignores gateway", "tcp://10.0.0.5:2375", true, testGatewayIp, "10.0.0.5"},
		{"https host", "https://sc:2376", false, "", "sc"},
		{"unix socket on bare binary", testUnixHost, false, "", testLocalhost},
		{"unix socket in container", testUnixHost, true, testGatewayIp, testGatewayIp},
		{"unix socket in container without gateway", testUnixHost, true, "", ""},
		{"empty host on bare binary", "", false, "", testLocalhost},
		{"npipe on bare binary", "npipe:////./pipe/docker_engine", false, "", testLocalhost},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, parseDialHost(tc.daemonHost, tc.isInContainer, tc.gatewayIp))
		})
	}
}

func newTestInspect(name string, networks map[string]string, ports nat.PortMap) container.InspectResponse {
	endpoints := make(map[string]*network.EndpointSettings, len(networks))
	for networkName, ip := range networks {
		endpoints[networkName] = &network.EndpointSettings{IPAddress: ip}
	}

	settings := &container.NetworkSettings{Networks: endpoints}

	settings.Ports = ports

	return container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{Name: name},
		NetworkSettings:   settings,
	}
}

func newTestAddressing(dialHost string, ownNetworks ...string) Addressing {
	own := make(map[string]struct{}, len(ownNetworks))
	for _, name := range ownNetworks {
		own[name] = struct{}{}
	}

	return Addressing{ownNetworks: own, dialHost: dialHost}
}

func Test_Addressing_Address_Cases(t *testing.T) {
	published := nat.PortMap{"3903/tcp": {{HostIP: "0.0.0.0", HostPort: "53903"}}}

	cases := []struct {
		name       string
		addressing Addressing
		info       container.InspectResponse
		port       int
		want       string
		wantErr    error
	}{
		{
			"shared verv network wins over other shared networks",
			newTestAddressing("sc", "a", vervNetworkName),
			newTestInspect(testGarageName, map[string]string{"a": testGarageIpA, vervNetworkName: "10.0.1.2"}, published),
			3903, "garage:3903", nil,
		},
		{
			"first shared network in sorted order without verv",
			newTestAddressing("sc", "b", "a"),
			newTestInspect(testGarageName, map[string]string{"b": "10.0.0.3", "a": testGarageIpA, "c": "10.9.9.9"}, published),
			3903, "garage:3903", nil,
		},
		{
			"ip on the default bridge which has no name resolution",
			newTestAddressing("sc", defaultBridgeNetwork),
			newTestInspect(testGarageName, map[string]string{defaultBridgeNetwork: testGarageIpA}, published),
			3903, testGarageIpA + ":3903", nil,
		},
		{
			"published port through daemon host when no network is shared",
			newTestAddressing("sc", vervNetworkName),
			newTestInspect(testGarageName, map[string]string{"other": testGarageIpA}, published),
			3903, "sc:53903", nil,
		},
		{
			"published port without own networks",
			newTestAddressing(testLocalhost),
			newTestInspect(testGarageName, map[string]string{vervNetworkName: "10.0.1.2"}, published),
			3903, "localhost:53903", nil,
		},
		{
			"unpublished port",
			newTestAddressing(testLocalhost),
			newTestInspect(testGarageName, nil, published),
			3900, "", user_errors.ErrContainerPortNotPublished,
		},
		{
			"binding with empty host port is not published",
			newTestAddressing(testLocalhost),
			newTestInspect(testGarageName, nil, nat.PortMap{"3903/tcp": {{HostIP: "0.0.0.0"}}}),
			3903, "", user_errors.ErrContainerPortNotPublished,
		},
		{
			"unknown daemon host",
			newTestAddressing(""),
			newTestInspect(testGarageName, nil, published),
			3903, "", errDaemonHostUnknown,
		},
		{
			"missing network settings",
			newTestAddressing(testLocalhost),
			container.InspectResponse{},
			3903, "", user_errors.ErrContainerPortNotPublished,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.addressing.Address(tc.info, tc.port)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
