package container_manager

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

func Test_MatchesContainerFilters_Service(t *testing.T) {
	cases := []struct {
		name    string
		dc      *velez_api.DockerContainer
		filters []*velez_api.ContainerFilter
		want    bool
	}{
		{
			name:    "no filters always matches",
			dc:      &velez_api.DockerContainer{},
			filters: nil,
			want:    true,
		},
		{
			name: "matches display name substring case-insensitively",
			dc: &velez_api.DockerContainer{
				Labels: map[string]string{labels.DisplayNameLabel: "Artel"},
			},
			filters: newServiceFilters("artel"),
			want:    true,
		},
		{
			name: "falls back to raw VERV_SERVICE label when no display name",
			dc: &velez_api.DockerContainer{
				Labels: map[string]string{labels.VervServiceLabel: "gitlab_runner_Artel"},
			},
			filters: newServiceFilters("artel"),
			want:    true,
		},
		{
			name: "no match",
			dc: &velez_api.DockerContainer{
				Labels: map[string]string{labels.DisplayNameLabel: "Artel"},
			},
			filters: newServiceFilters("other"),
			want:    false,
		},
		{
			name: "unspecified field ignored",
			dc:   &velez_api.DockerContainer{},
			filters: []*velez_api.ContainerFilter{
				{Field: velez_api.ContainerFilterField_unspecified, Value: "anything"},
			},
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := matchesContainerFilters(tc.dc, tc.filters)
			require.Equal(t, tc.want, got)
		})
	}
}

func newServiceFilters(value string) []*velez_api.ContainerFilter {
	filter := &velez_api.ContainerFilter{
		Field: velez_api.ContainerFilterField_service,
		Value: value,
	}

	return []*velez_api.ContainerFilter{filter}
}
