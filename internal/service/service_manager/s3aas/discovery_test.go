package s3aas

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	refTestInstance = "main"
	refTestWebUi    = 3909
)

func newRootSummary(publicPort uint16) container.Summary {
	port := container.Port{PrivatePort: domain.S3WebUiContainerPort, PublicPort: publicPort, Type: tcpProtocol}

	return container.Summary{Ports: []container.Port{port}}
}

func newSidecarWebUiSummary() container.Summary {
	return container.Summary{Labels: map[string]string{
		labels.S3WebUiLabel: refTestInstance,
		labels.Sidecar:      "true",
	}}
}

func newLegacyWebUiSummary(publicPort uint16) container.Summary {
	summary := newRootSummary(publicPort)

	summary.Labels = map[string]string{labels.S3WebUiLabel: refTestInstance}

	return summary
}

func Test_InstanceRef_SidecarWebUi(t *testing.T) {
	sidecar := newSidecarWebUiSummary()
	ref := instanceRef{name: refTestInstance, instance: newRootSummary(32001), webUi: &sidecar}

	require.True(t, ref.isWebUiSidecar())
	require.Equal(t, uint32(32001), hostPort(ref.webUiPortSource(), refTestWebUi))
	require.Equal(t, domain.S3ServiceName(refTestInstance), ref.webUiHostName())
}

func Test_InstanceRef_LegacyWebUi(t *testing.T) {
	legacy := newLegacyWebUiSummary(32002)
	ref := instanceRef{name: refTestInstance, instance: newRootSummary(32001), webUi: &legacy}

	require.False(t, ref.isWebUiSidecar())
	require.Equal(t, uint32(32002), hostPort(ref.webUiPortSource(), refTestWebUi))
	require.Equal(t, domain.S3WebUiServiceName(refTestInstance), ref.webUiHostName())
}

func Test_InstanceRef_NoWebUi(t *testing.T) {
	ref := instanceRef{name: refTestInstance, instance: newRootSummary(32001)}

	require.False(t, ref.isWebUiSidecar())
}

type endpointCase struct {
	name       string
	remoteHost string
	port       uint32
	want       string
}

func Test_S3Endpoint_ResolvesHost(t *testing.T) {
	cases := []endpointCase{
		{name: "remote host", remoteHost: "10.0.0.7", port: 3900, want: "http://10.0.0.7:3900"},
		{name: "remote hostname", remoteHost: "node-2.lan", port: 3901, want: "http://node-2.lan:3901"},
	}

	if !env.IsInContainer() {
		cases = append(cases, endpointCase{name: "local bare binary", port: 3900, want: "http://localhost:3900"})
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := s3Endpoint("s3-main", tc.remoteHost, tc.port)
			require.Equal(t, tc.want, got)
		})
	}
}

func Test_WebUiEndpoint_ResolvesHost(t *testing.T) {
	cases := []endpointCase{
		{name: "remote host", remoteHost: "10.0.0.7", port: 3902, want: "http://10.0.0.7:3902"},
		{name: "remote hostname", remoteHost: "node-2.lan", port: 3903, want: "http://node-2.lan:3903"},
	}

	if !env.IsInContainer() {
		cases = append(cases, endpointCase{name: "local bare binary", port: 3902, want: "http://localhost:3902"})
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := webUiEndpoint("s3-main-webui", tc.remoteHost, tc.port)
			require.Equal(t, tc.want, got)
		})
	}
}
