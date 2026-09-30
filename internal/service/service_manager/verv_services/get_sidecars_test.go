package verv_services

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

func Test_ToServiceSidecar_Mapping(t *testing.T) {
	summary := container.Summary{
		ID:    "aaa111",
		Names: []string{"/ts-sidecar"},
		Image: "tailscale:latest",
		State: "running",
	}

	got := toServiceSidecar(summary)

	require.Equal(t, "aaa111", got.ContainerId)
	require.Equal(t, "ts-sidecar", got.ContainerName)
	require.Equal(t, "tailscale:latest", got.ImageName)
	require.Equal(t, velez_api.Smerd_running, got.Status)
}

func Test_ToServiceSidecar_NoNames(t *testing.T) {
	got := toServiceSidecar(container.Summary{ID: "aaa111"})

	require.Equal(t, "", got.ContainerName)
}

func Test_IsServiceRoot_Resolution(t *testing.T) {
	bound := map[string]struct{}{"bound-root": {}}

	cases := []struct {
		name string
		cont container.Summary
		want bool
	}{
		{
			name: "labelled root",
			cont: container.Summary{Labels: map[string]string{labels.VervServiceLabel: "svc"}},
			want: true,
		},
		{
			name: "bound root",
			cont: container.Summary{Names: []string{"/bound-root"}},
			want: true,
		},
		{
			name: "labelled sidecar",
			cont: container.Summary{Labels: map[string]string{
				labels.VervServiceLabel: "svc",
				labels.Sidecar:          "true",
			}},
			want: false,
		},
		{
			name: "other service",
			cont: container.Summary{Labels: map[string]string{labels.VervServiceLabel: "other"}},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, isServiceRoot(tc.cont, "svc", bound))
		})
	}
}
