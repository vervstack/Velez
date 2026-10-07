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
