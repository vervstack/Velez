package vervonomicon

import (
	"testing"

	"github.com/stretchr/testify/require"

	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
)

func baseFiles() map[string][]byte {
	return map[string][]byte{
		indexFileName: []byte(`
version: "1"
service:
  name: zpotify
`),
		"deployment.yaml": []byte(`
app:
  command: run.sh
  use_image_ports: true
  auto_upgrade: false
  env:
    SOME_FLAG: "1"
  labels:
    team: media
`),
		"ingress.yaml": []byte(`
domains:
  - host: zpotify.ru
    ssl: enabled
    listen: 443
  - host: old.zpotify.ru
    ssl: disabled
`),
		"ingress.conf": []byte("base config"),
	}
}

func TestMergeEnvironment_MapsDeepMergeByKey(t *testing.T) {
	files := baseFiles()

	files["prod/deployment.yaml"] = []byte(`
app:
  env:
    OTHER_FLAG: "2"
`)

	descriptor, err := MergeEnvironment(files, "prod")
	require.NoError(t, err)
	require.Equal(t, "run.sh", descriptor.Deployment.App.Command)
	require.Equal(t, map[string]string{"SOME_FLAG": "1", "OTHER_FLAG": "2"}, descriptor.Deployment.App.Env)
	require.Equal(t, "prod", descriptor.Environment)
}

func TestMergeEnvironment_ScalarOverlayReplacesBase(t *testing.T) {
	files := baseFiles()

	files["prod/deployment.yaml"] = []byte(`
app:
  command: prod_run.sh
`)

	descriptor, err := MergeEnvironment(files, "prod")
	require.NoError(t, err)
	require.Equal(t, "prod_run.sh", descriptor.Deployment.App.Command)
}

func TestMergeEnvironment_OverlayFalseZeroDoesNotClobberUnmentionedBaseValues(t *testing.T) {
	files := baseFiles()

	// use_image_ports: true and command: run.sh live only in the base. The
	// overlay only ever mentions auto_upgrade - it must not reset the other
	// fields to their zero values.
	files["prod/deployment.yaml"] = []byte(`
app:
  auto_upgrade: true
`)

	descriptor, err := MergeEnvironment(files, "prod")
	require.NoError(t, err)
	require.True(t, descriptor.Deployment.App.UseImagePorts)
	require.Equal(t, "run.sh", descriptor.Deployment.App.Command)
	require.True(t, descriptor.Deployment.App.AutoUpgrade)
}

func TestMergeEnvironment_ListsOfObjectsMergeByName(t *testing.T) {
	files := baseFiles()

	files["resources.yaml"] = []byte(`
- name: pg
  type: postgres
  isolation: shared_pool
  box: small
- name: storage
  type: local_volume
  container_path: /data
`)
	files["prod/resources.yaml"] = []byte(`
- name: pg
  box: medium
- name: cache
  type: redis
  box: small
`)

	descriptor, err := MergeEnvironment(files, "prod")
	require.NoError(t, err)
	require.Len(t, descriptor.Resources, 3)

	byName := make(map[string]verv.Resource, len(descriptor.Resources))
	for _, r := range descriptor.Resources {
		byName[r.Name] = r
	}

	// merged in place - type/isolation untouched, box overridden
	require.Equal(t, "postgres", byName["pg"].Type)
	require.Equal(t, verv.IsolationSharedPool, byName["pg"].Isolation)
	require.Equal(t, "medium", byName["pg"].Box)

	// untouched entry survives
	require.Equal(t, "/data", byName["storage"].ContainerPath)

	// unseen name appended
	require.Equal(t, "redis", byName["cache"].Type)
}

// TestMergeEnvironment_ListsOfObjectsWithoutNameReplaceWholesale documents a
// consequence of the spec's literal "merge by name" wording: ingress.yaml's
// domains are keyed by "host", not "name", so they don't qualify for the
// merge-by-name rule and fall back to wholesale replacement like any other
// list of objects without a "name" field.
func TestMergeEnvironment_ListsOfObjectsWithoutNameReplaceWholesale(t *testing.T) {
	files := baseFiles()

	files["prod/ingress.yaml"] = []byte(`
domains:
  - host: new.zpotify.ru
    ssl: enabled
`)

	descriptor, err := MergeEnvironment(files, "prod")
	require.NoError(t, err)
	require.Len(t, descriptor.Ingress.Domains, 1)
	require.Equal(t, "new.zpotify.ru", descriptor.Ingress.Domains[0].Host)
}

func TestMergeEnvironment_NonYamlFileReplacedWholesale(t *testing.T) {
	files := baseFiles()

	files["prod/ingress.conf"] = []byte("prod config, totally different shape")

	descriptor, err := MergeEnvironment(files, "prod")
	require.NoError(t, err)
	require.Equal(t, []byte("prod config, totally different shape"), descriptor.Raw["ingress.conf"])
}

func TestMergeEnvironment_OtherEnvironmentDirIsDropped(t *testing.T) {
	files := baseFiles()

	files["local/deployment.yaml"] = []byte(`
app:
  command: local_run.sh
`)

	descriptor, err := MergeEnvironment(files, "prod")
	require.NoError(t, err)
	require.Equal(t, "run.sh", descriptor.Deployment.App.Command)
	require.NotContains(t, descriptor.Raw, "local/deployment.yaml")
}

func TestMergeEnvironment_NoEnvironmentAppliesNoOverlay(t *testing.T) {
	files := baseFiles()

	files["prod/deployment.yaml"] = []byte(`
app:
  command: prod_run.sh
`)

	descriptor, err := MergeEnvironment(files, "")
	require.NoError(t, err)
	require.Equal(t, "run.sh", descriptor.Deployment.App.Command)
	require.Equal(t, "", descriptor.Environment)
}

func TestMergeEnvironment_OverlayOnlyFileIsAdded(t *testing.T) {
	files := baseFiles()

	files["local/ingress.conf"] = []byte("local-only config")

	descriptor, err := MergeEnvironment(files, "local")
	require.NoError(t, err)
	require.Equal(t, []byte("local-only config"), descriptor.Raw["ingress.conf"])
}
