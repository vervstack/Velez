package pgaas

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
)

var errBoxNotFound = rerrors.New("box not found")

type fakeBoxLookup struct {
	boxes map[string]verv.Box
}

func (f *fakeBoxLookup) GetBox(_ context.Context, name string) (verv.Box, error) {
	box, ok := f.boxes[name]
	if !ok {
		return verv.Box{}, errBoxNotFound
	}

	return box, nil
}

func (f *fakeBoxLookup) ListBoxes(_ context.Context) ([]verv.Box, error) {
	out := make([]verv.Box, 0, len(f.boxes))
	for _, box := range f.boxes {
		out = append(out, box)
	}

	return out, nil
}

func newFakeBoxLookup() *fakeBoxLookup {
	return &fakeBoxLookup{
		boxes: map[string]verv.Box{
			"small": {Name: "small", Cpu: 0.5, RamMb: 512, DiskMb: 2048, IsBuiltin: true},
		},
	}
}

func TestBuildDeployRequest_ResolvesShapeAndOverlaysCredentialsOnlyOntoRequest(t *testing.T) {
	req := domain.CreatePgInstanceReq{
		Name:         "my-pg",
		Environment:  "prod",
		ExposeToPort: 15432,
	}

	creds := pgCredentials{
		dbName:   "my_pg",
		username: "my_pg_user",
		password: "s3cr3t",
	}

	descriptor, request, err := buildDeployRequest(context.Background(), newFakeBoxLookup(), req, creds)
	require.NoError(t, err)

	require.Equal(t, "postgres:18", request.GetImageName())
	require.Equal(t, "my-pg", request.GetName())
	require.Equal(t, "prod", request.GetEnvironment())

	settings := request.GetSettings()
	require.NotNil(t, settings)
	require.Len(t, settings.GetPorts(), 1)
	require.EqualValues(t, 5432, settings.GetPorts()[0].GetServicePortNumber())
	require.NotNil(t, settings.GetPorts()[0].ExposedTo)
	require.EqualValues(t, 15432, settings.GetPorts()[0].GetExposedTo())

	require.Len(t, settings.GetVolumes(), 1)
	require.Equal(t, "my-pg-data", settings.GetVolumes()[0].GetVolumeName())
	require.Equal(t, "/var/lib/postgresql", settings.GetVolumes()[0].GetContainerPath())

	require.Equal(t, map[string]string{
		"POSTGRES_DB":       "my_pg",
		"POSTGRES_USER":     "my_pg_user",
		"POSTGRES_PASSWORD": "s3cr3t",
	}, request.GetEnv())

	require.Empty(t, descriptor.Deployment.App.Env)
	require.NotContains(t, descriptor.Deployment.App.Env, "POSTGRES_PASSWORD")
}
