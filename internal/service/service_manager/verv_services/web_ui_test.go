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
)

func Test_WebUiRootService_Mapping(t *testing.T) {
	cases := []struct {
		name     string
		resource domain.BoundResource
		want     string
		wantOk   bool
	}{
		{
			name:     "s3 bucket resolves to instance service",
			resource: domain.BoundResource{ResourceType: domain.S3ResourceType, Name: "main/photos"},
			want:     domain.S3ServiceName(webUiTestInstance),
			wantOk:   true,
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
			got, ok := webUiRootService(tc.resource)

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
			ports: []container.Port{{PrivatePort: 3909, PublicPort: 32768, Type: "tcp"}},
			want:  32768,
		},
		{
			name:  "no public port",
			ports: []container.Port{{PrivatePort: 3909, Type: "tcp"}},
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
		ID:     "legacy",
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
			wantIds:     []string{"legacy"},
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
