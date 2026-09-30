package verv_services

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
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
