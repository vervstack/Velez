package address_book

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	testService   = "svc"
	testInstance  = "main"
	testLegacyId  = "legacy"
	testHost      = "host"
	testSidecar   = "true"
	testUiPort    = 3909
	testVcnIpV4   = "100.64.0.7"
	testVcnIpV6   = "fd7a:115c:a1e0::7"
	testOtherNode = "other-ts-sidecar"
)

func Test_WebUiTargetOf_Labels(t *testing.T) {
	cases := []struct {
		name   string
		labels map[string]string
		want   WebUiTarget
		wantOk bool
	}{
		{
			name: "generic labels",
			labels: map[string]string{
				labels.WebUiForLabel:  testService,
				labels.WebUiPortLabel: "8080",
			},
			want:   WebUiTarget{RootService: testService, ContainerPort: 8080},
			wantOk: true,
		},
		{
			name: "generic label with bad port",
			labels: map[string]string{
				labels.WebUiForLabel:  testService,
				labels.WebUiPortLabel: "not-a-port",
			},
		},
		{
			name:   "legacy s3 label",
			labels: map[string]string{labels.S3WebUiLabel: testInstance},
			want: WebUiTarget{
				RootService:   domain.S3ServiceName(testInstance),
				ContainerPort: domain.S3WebUiContainerPort,
			},
			wantOk: true,
		},
		{
			name:   "no labels",
			labels: map[string]string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := WebUiTargetOf(tc.labels)

			require.Equal(t, tc.wantOk, ok)
			require.Equal(t, tc.want, got)
		})
	}
}

func Test_PublishedPort_Lookup(t *testing.T) {
	cases := []struct {
		name  string
		ports []container.Port
		want  uint32
	}{
		{
			name:  "matching tcp port",
			ports: []container.Port{{PrivatePort: 3909, PublicPort: 32768, Type: tcpProtocol}},
			want:  32768,
		},
		{
			name:  "no public port",
			ports: []container.Port{{PrivatePort: 3909, Type: tcpProtocol}},
		},
		{
			name:  "wrong protocol",
			ports: []container.Port{{PrivatePort: 3909, PublicPort: 32768, Type: "udp"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			summary := container.Summary{Ports: tc.ports}

			require.Equal(t, tc.want, publishedPort(summary, 3909))
		})
	}
}

func newRootSummary(serviceName string, publicPort uint16) container.Summary {
	port := container.Port{PrivatePort: testUiPort, PublicPort: publicPort, Type: tcpProtocol}

	return container.Summary{
		ID:     "root-" + serviceName,
		Labels: map[string]string{labels.VervServiceLabel: serviceName},
		Ports:  []container.Port{port},
	}
}

func newSidecarSummary(rootService string) container.Summary {
	return container.Summary{
		ID: "sidecar",
		Labels: map[string]string{
			labels.VervServiceLabel: rootService,
			labels.Sidecar:          testSidecar,
			labels.S3WebUiLabel:     testInstance,
			labels.WebUiForLabel:    rootService,
			labels.WebUiPortLabel:   "3909",
		},
	}
}

func Test_DiscoverWebUis_Layouts(t *testing.T) {
	rootService := domain.S3ServiceName(testInstance)
	sidecar := newSidecarSummary(rootService)

	legacyPort := container.Port{PrivatePort: testUiPort, PublicPort: 40000, Type: tcpProtocol}
	legacy := container.Summary{
		ID:     testLegacyId,
		Labels: map[string]string{labels.S3WebUiLabel: testInstance},
		Ports:  []container.Port{legacyPort},
	}

	published := func(port uint32) map[string]webUiDiscovery {
		found := webUiDiscovery{host: testHost, containerPort: testUiPort, publishedPort: port}

		return map[string]webUiDiscovery{rootService: found}
	}

	cases := []struct {
		name       string
		containers []container.Summary
		want       map[string]webUiDiscovery
	}{
		{
			name:       "sidecar reads port from root container",
			containers: []container.Summary{newRootSummary(rootService, 32001), sidecar},
			want:       published(32001),
		},
		{
			name:       "sidecar listed before root",
			containers: []container.Summary{sidecar, newRootSummary(rootService, 32001)},
			want:       published(32001),
		},
		{
			name:       "sidecar without root is reported unpublished",
			containers: []container.Summary{sidecar},
			want:       published(0),
		},
		{
			name:       "sidecar ignores root without published port",
			containers: []container.Summary{newRootSummary(rootService, 0), sidecar},
			want:       published(0),
		},
		{
			name:       "legacy reads port from its own container",
			containers: []container.Summary{newRootSummary(rootService, 0), legacy},
			want:       published(40000),
		},
		{
			name:       "unpublished web ui does not hide a published one",
			containers: []container.Summary{newRootSummary(rootService, 0), legacy, sidecar},
			want:       published(40000),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, discoverWebUis(tc.containers, testHost))
		})
	}
}

func Test_VcnHostname_ReplacesUnderscores(t *testing.T) {
	require.Equal(t, "garage-main-ts-sidecar", vcnHostname("garage_main"))
}

func Test_VcnIpOf_PicksIpv4(t *testing.T) {
	nodes := []domain.VcnNode{
		{Name: testOtherNode, IpAddresses: []string{"100.64.0.9"}},
		{Name: vcnHostname(testService), IpAddresses: []string{testVcnIpV6, testVcnIpV4}},
	}

	ip, ok := vcnIpOf(testService, nodes)

	require.True(t, ok)
	require.Equal(t, testVcnIpV4, ip)
}

func Test_VcnIpOf_UnderscoreServiceMatchesHyphenatedNode(t *testing.T) {
	nodes := []domain.VcnNode{{Name: "garage-main-ts-sidecar", IpAddresses: []string{testVcnIpV4}}}

	ip, ok := vcnIpOf("garage_main", nodes)

	require.True(t, ok)
	require.Equal(t, testVcnIpV4, ip)
}

func Test_VcnIpOf_NoMatchingNode(t *testing.T) {
	nodes := []domain.VcnNode{{Name: testOtherNode, IpAddresses: []string{testVcnIpV4}}}

	_, ok := vcnIpOf(testService, nodes)

	require.False(t, ok)
}

func Test_VcnIpOf_OnlyIpv6(t *testing.T) {
	nodes := []domain.VcnNode{{Name: vcnHostname(testService), IpAddresses: []string{testVcnIpV6}}}

	_, ok := vcnIpOf(testService, nodes)

	require.False(t, ok)
}

func Test_BuildAddresses_DockerAndVcn(t *testing.T) {
	found := webUiDiscovery{host: testHost, containerPort: testUiPort, publishedPort: 32001}
	nodes := []domain.VcnNode{{Name: vcnHostname(testService), IpAddresses: []string{testVcnIpV4}}}

	got := buildAddresses(testService, found, nodes)

	want := serviceAddresses{
		{
			ServiceName: testService,
			Name:        domain.WebUiAddressName,
			Host:        testHost,
			Port:        32001,
			Scope:       velez_api.AddressScope_ADDRESS_SCOPE_DOCKER,
		},
		{
			ServiceName: testService,
			Name:        domain.WebUiAddressName,
			Host:        testVcnIpV4,
			Port:        testUiPort,
			Scope:       velez_api.AddressScope_ADDRESS_SCOPE_VCN,
		},
	}

	require.Equal(t, want, got)
}

func Test_BuildAddresses_UnpublishedHasOnlyVcn(t *testing.T) {
	found := webUiDiscovery{host: testHost, containerPort: testUiPort}
	nodes := []domain.VcnNode{{Name: vcnHostname(testService), IpAddresses: []string{testVcnIpV4}}}

	got := buildAddresses(testService, found, nodes)

	require.Len(t, got, 1)
	require.Equal(t, velez_api.AddressScope_ADDRESS_SCOPE_VCN, got[0].Scope)
}

func Test_BuildAddresses_UnpublishedWithoutNodeHasNothing(t *testing.T) {
	found := webUiDiscovery{host: testHost, containerPort: testUiPort}

	require.Empty(t, buildAddresses(testService, found, nil))
}
