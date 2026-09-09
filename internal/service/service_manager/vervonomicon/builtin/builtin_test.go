package builtin_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon/builtin"
)

func TestRead_RoundTripsThroughParse(t *testing.T) {
	cases := []struct {
		name       string
		wantImage  string
		wantPort   int
		wantVolume int
	}{
		{name: "postgres", wantImage: "postgres:18", wantPort: 5432, wantVolume: 1},
		{name: "registry", wantImage: "registry:2", wantPort: 5000, wantVolume: 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files, err := builtin.Read(tc.name)
			require.NoError(t, err)
			require.Contains(t, files, "vervonomicon.yaml")
			require.Contains(t, files, "deployment.yaml")

			descriptor, err := vervonomicon.Parse(files)
			require.NoError(t, err)

			require.Equal(t, tc.name, descriptor.Index.Service.Name)
			require.Equal(t, tc.wantImage, descriptor.Deployment.App.Image)
			require.Len(t, descriptor.Deployment.App.Ports, 1)
			require.Equal(t, tc.wantPort, descriptor.Deployment.App.Ports[0].Port)
			require.Equal(t, verv.ProtocolTcp, descriptor.Deployment.App.Ports[0].Protocol)
			require.Len(t, descriptor.Deployment.App.Volumes, tc.wantVolume)
		})
	}
}

func TestRead_UnknownNameErrors(t *testing.T) {
	_, err := builtin.Read("does-not-exist")
	require.Error(t, err)
}
