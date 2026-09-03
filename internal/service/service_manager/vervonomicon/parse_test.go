package vervonomicon

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
)

func TestParse_NoDescriptor(t *testing.T) {
	files := map[string][]byte{
		"deployment.yaml": []byte("app:\n  command: hi\n"),
	}

	_, err := Parse(files)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrNoDescriptor))
}

func TestParse_MinimalDescriptor(t *testing.T) {
	files := map[string][]byte{
		indexFileName: []byte(`
version: "1"
service:
  name: zpotify
  description: Self-hosted spotify alternative
box: small
`),
	}

	descriptor, err := Parse(files)
	require.NoError(t, err)
	require.Equal(t, "zpotify", descriptor.Index.Service.Name)
	require.Equal(t, "small", descriptor.Index.Box)
	require.Equal(t, verv.Deployment{}, descriptor.Deployment)
	require.Nil(t, descriptor.Resources)
	require.Equal(t, files, descriptor.Raw)
}

func TestParse_UnrecognisedVersion(t *testing.T) {
	files := map[string][]byte{
		indexFileName: []byte(`
version: "2"
service:
  name: zpotify
`),
	}

	descriptor, err := Parse(files)
	require.Error(t, err)
	require.Contains(t, err.Error(), "2")
	require.Equal(t, verv.Descriptor{}, descriptor)
}

func TestParse_PathOverrideResolvesCustomFile(t *testing.T) {
	files := map[string][]byte{
		indexFileName: []byte(`
version: "1"
service:
  name: zpotify
deployment: custom_deploy.yaml
`),
		"custom_deploy.yaml": []byte(`
app:
  command: run.sh
`),
		// deployment.yaml, the conventional default, is deliberately absent
		// to prove the override actually took effect.
	}

	descriptor, err := Parse(files)
	require.NoError(t, err)
	require.Equal(t, "run.sh", descriptor.Deployment.App.Command)
}

func TestParse_MissingOptionalFileIsNotAnError(t *testing.T) {
	files := map[string][]byte{
		indexFileName: []byte(`
version: "1"
service:
  name: zpotify
`),
	}

	descriptor, err := Parse(files)
	require.NoError(t, err)
	require.Equal(t, verv.Ingress{}, descriptor.Ingress)
	require.Equal(t, verv.Auth{}, descriptor.Auth)
}

func TestParse_MalformedReferencedFileIsAnError(t *testing.T) {
	files := map[string][]byte{
		indexFileName: []byte(`
version: "1"
service:
  name: zpotify
`),
		"resources.yaml": []byte("this is not: [a valid, resources list"),
	}

	_, err := Parse(files)
	require.Error(t, err)
}

func TestParse_ResourcesIsAFlatList(t *testing.T) {
	files := map[string][]byte{
		indexFileName: []byte(`
version: "1"
service:
  name: zpotify
`),
		"resources.yaml": []byte(`
- name: pg
  type: postgres
  isolation: shared_pool
  box: medium
  binds_to: data_sources.postgres
- name: storage
  type: local_volume
  container_path: /data
  sticky: true
`),
	}

	descriptor, err := Parse(files)
	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 2)
	require.Equal(t, "pg", descriptor.Resources[0].Name)
	require.Equal(t, verv.IsolationSharedPool, descriptor.Resources[0].Isolation)
	require.Equal(t, "storage", descriptor.Resources[1].Name)
	require.True(t, descriptor.Resources[1].Sticky)
}
