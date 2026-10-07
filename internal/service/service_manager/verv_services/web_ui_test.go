package verv_services

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	webUiTestService   = "svc"
	webUiTestInstance  = "main"
	webUiTestGenericId = "generic"
	webUiTestLegacyId  = "legacy"
	webUiTestHost      = "host"

	webUiTestSidecarValue = "true"
)

func Test_WebUiRootService_Mapping(t *testing.T) {
	cases := []struct {
		name        string
		resource    domain.BoundResource
		serviceName string
		want        string
		wantOk      bool
	}{
		{
			name:        "own web ui resolves to the page's service",
			resource:    domain.BoundResource{ResourceType: webUiResourceType, Name: webUiResourceType},
			serviceName: domain.S3ServiceName(webUiTestInstance),
			want:        domain.S3ServiceName(webUiTestInstance),
			wantOk:      true,
		},
		{
			name:        "non web ui resource on the same service resolves nothing",
			resource:    domain.BoundResource{ResourceType: "postgres", Name: "postgres"},
			serviceName: domain.S3ServiceName(webUiTestInstance),
		},
		{
			name:        "s3 bucket resolves to instance service",
			resource:    domain.BoundResource{ResourceType: domain.S3ResourceType, Name: "main/photos"},
			serviceName: webUiTestService,
			want:        domain.S3ServiceName(webUiTestInstance),
			wantOk:      true,
		},
		{
			name:     "s3 bucket without slash",
			resource: domain.BoundResource{ResourceType: domain.S3ResourceType, Name: webUiTestInstance},
		},
		{
			name:     "unknown resource type",
			resource: domain.BoundResource{ResourceType: "pg_database", Name: "main/db"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := webUiRootService(tc.resource, tc.serviceName)

			require.Equal(t, tc.wantOk, ok)
			require.Equal(t, tc.want, got)
		})
	}
}

func Test_WebUiTargetOf_Labels(t *testing.T) {
	cases := []struct {
		name   string
		labels map[string]string
		want   webUiTarget
		wantOk bool
	}{
		{
			name: "generic labels",
			labels: map[string]string{
				labels.WebUiForLabel:  webUiTestService,
				labels.WebUiPortLabel: "8080",
			},
			want:   webUiTarget{rootService: webUiTestService, containerPort: 8080},
			wantOk: true,
		},
		{
			name: "generic label with bad port",
			labels: map[string]string{
				labels.WebUiForLabel:  webUiTestService,
				labels.WebUiPortLabel: "not-a-port",
			},
		},
		{
			name:   "legacy s3 label",
			labels: map[string]string{labels.S3WebUiLabel: webUiTestInstance},
			want: webUiTarget{
				rootService:   domain.S3ServiceName(webUiTestInstance),
				containerPort: domain.S3WebUiContainerPort,
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
			got, ok := webUiTargetOf(tc.labels)

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

func Test_WebUiSidecars_Selection(t *testing.T) {
	genericUi := container.Summary{
		ID:    webUiTestGenericId,
		Names: []string{"/generic-ui"},
		Labels: map[string]string{
			labels.WebUiForLabel:  webUiTestService,
			labels.WebUiPortLabel: "8080",
		},
	}
	legacyUi := container.Summary{
		ID:     webUiTestLegacyId,
		Names:  []string{"/legacy-ui"},
		Labels: map[string]string{labels.S3WebUiLabel: webUiTestInstance},
	}
	otherUi := container.Summary{
		ID: "other",
		Labels: map[string]string{
			labels.WebUiForLabel:  "another",
			labels.WebUiPortLabel: "8080",
		},
	}
	all := []container.Summary{genericUi, legacyUi, otherUi}

	cases := []struct {
		name        string
		serviceName string
		known       []domain.ServiceSidecar
		wantIds     []string
	}{
		{
			name:        "matches by generic label",
			serviceName: webUiTestService,
			wantIds:     []string{webUiTestGenericId},
		},
		{
			name:        "matches by legacy label",
			serviceName: domain.S3ServiceName(webUiTestInstance),
			wantIds:     []string{webUiTestLegacyId},
		},
		{
			name:        "dedupes known sidecars",
			serviceName: webUiTestService,
			known:       []domain.ServiceSidecar{{ContainerId: webUiTestGenericId}},
			wantIds:     []string{webUiTestGenericId},
		},
		{
			name:        "ignores other services",
			serviceName: "missing",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := webUiSidecars(all, tc.serviceName, tc.known)

			var ids []string

			for _, sidecar := range got {
				ids = append(ids, sidecar.ContainerId)
			}

			require.Equal(t, tc.wantIds, ids)
		})
	}
}

func newWebUiRootSummary(serviceName string, publicPort uint16) container.Summary {
	port := container.Port{PrivatePort: 3909, PublicPort: publicPort, Type: tcpProtocol}

	return container.Summary{
		ID:     "root-" + serviceName,
		Labels: map[string]string{labels.VervServiceLabel: serviceName},
		Ports:  []container.Port{port},
	}
}

func Test_WebUiAddresses_Layouts(t *testing.T) {
	rootService := domain.S3ServiceName(webUiTestInstance)

	sidecar := container.Summary{
		ID: "sidecar",
		Labels: map[string]string{
			labels.VervServiceLabel: rootService,
			labels.Sidecar:          webUiTestSidecarValue,
			labels.S3WebUiLabel:     webUiTestInstance,
			labels.WebUiForLabel:    rootService,
			labels.WebUiPortLabel:   "3909",
		},
	}

	legacyPort := container.Port{PrivatePort: 3909, PublicPort: 40000, Type: tcpProtocol}
	legacy := container.Summary{
		ID:     webUiTestLegacyId,
		Labels: map[string]string{labels.S3WebUiLabel: webUiTestInstance},
		Ports:  []container.Port{legacyPort},
	}

	cases := []struct {
		name       string
		containers []container.Summary
		want       map[string]webUiAddress
	}{
		{
			name:       "sidecar reads port from root container",
			containers: []container.Summary{newWebUiRootSummary(rootService, 32001), sidecar},
			want:       map[string]webUiAddress{rootService: {host: webUiTestHost, port: 32001}},
		},
		{
			name:       "sidecar listed before root",
			containers: []container.Summary{sidecar, newWebUiRootSummary(rootService, 32001)},
			want:       map[string]webUiAddress{rootService: {host: webUiTestHost, port: 32001}},
		},
		{
			name:       "sidecar without root is skipped",
			containers: []container.Summary{sidecar},
			want:       map[string]webUiAddress{},
		},
		{
			name:       "sidecar ignores root without published port",
			containers: []container.Summary{newWebUiRootSummary(rootService, 0), sidecar},
			want:       map[string]webUiAddress{},
		},
		{
			name:       "legacy reads port from its own container",
			containers: []container.Summary{newWebUiRootSummary(rootService, 0), legacy},
			want:       map[string]webUiAddress{rootService: {host: webUiTestHost, port: 40000}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, webUiAddresses(tc.containers, webUiTestHost))
		})
	}
}

func Test_WebUiRootService_OwnWebUiMatchesSidecarAddress(t *testing.T) {
	rootService := domain.S3ServiceName(webUiTestInstance)

	sidecar := container.Summary{
		ID: "sidecar",
		Labels: map[string]string{
			labels.VervServiceLabel: rootService,
			labels.Sidecar:          webUiTestSidecarValue,
			labels.WebUiForLabel:    rootService,
			labels.WebUiPortLabel:   "3909",
		},
	}
	containers := []container.Summary{newWebUiRootSummary(rootService, 32001), sidecar}

	resource := domain.BoundResource{ResourceType: webUiResourceType, Name: webUiResourceType}

	key, ok := webUiRootService(resource, rootService)
	require.True(t, ok)

	address, found := webUiAddresses(containers, webUiTestHost)[key]
	require.True(t, found)
	require.Equal(t, webUiAddress{host: webUiTestHost, port: 32001}, address)
}
