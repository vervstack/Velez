package jobs

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon"
	"go.vervstack.ru/Velez/internal/service/service_manager/vervonomicon/builtin"
)

// TestEnableRegistryTaskPayload_JsonRoundTrip guards against
// docs/features/pgaas_and_registry_plugin.md's jobs-engine rule #1: a oneof
// field (a Go interface) silently survives Marshal but is dropped by
// Unmarshal. EnableRegistry carries no oneof, so this is expected to pass
// with plain encoding/json - this test is what proves that, rather than
// assuming it.
func TestEnableRegistryTaskPayload_JsonRoundTrip(t *testing.T) {
	port := uint32(5432) //nolint:mnd

	want := &velez_api.EnableRegistryTaskPayload{
		Request: &velez_api.EnableRegistry{
			ExposeToPort: &port,
			Username:     stringPtr("verv"),
		},
	}
	want.SetUsername("verv")
	want.SetPassword("s3cr3t")
	want.SetContainerId("cont-123")
	want.SetExposedPort(15000) //nolint:mnd

	raw, err := json.Marshal(want)
	require.NoError(t, err)

	got := &velez_api.EnableRegistryTaskPayload{}

	err = json.Unmarshal(raw, got)
	require.NoError(t, err)

	require.Equal(t, want.GetRequest().GetExposeToPort(), got.GetRequest().GetExposeToPort())
	require.Equal(t, want.GetRequest().GetUsername(), got.GetRequest().GetUsername())
	require.Equal(t, want.GetUsername(), got.GetUsername())
	require.Equal(t, want.GetPassword(), got.GetPassword())
	require.Equal(t, want.GetContainerId(), got.GetContainerId())
	require.Equal(t, want.GetExposedPort(), got.GetExposedPort())
}

func stringPtr(s string) *string {
	return &s
}

func TestHtpasswdLine(t *testing.T) {
	line, err := htpasswdLine("verv", "hunter2")
	require.NoError(t, err)

	username, hash, ok := splitHtpasswdLine(line)
	require.True(t, ok)
	require.Equal(t, "verv", username)

	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte("hunter2"))
	require.NoError(t, err)
}

// splitHtpasswdLine parses "user:hash\n" back apart - a small test-local
// helper, not production code, since nothing in enable_registry.go needs to
// parse its own output.
func splitHtpasswdLine(line string) (username, hash string, ok bool) {
	trimmed := line
	if len(trimmed) > 0 && trimmed[len(trimmed)-1] == '\n' {
		trimmed = trimmed[:len(trimmed)-1]
	}

	for i := range len(trimmed) {
		if trimmed[i] == ':' {
			return trimmed[:i], trimmed[i+1:], true
		}
	}

	return "", "", false
}

// fakeRegistryBoxLookup satisfies vervonomicon.BoxLookup with just the
// builtin registry descriptor's "small" box (see
// builtin/registry/vervonomicon.yaml), avoiding any real storage dependency.
type fakeRegistryBoxLookup struct{}

func (fakeRegistryBoxLookup) GetBox(_ context.Context, name string) (verv.Box, error) {
	return verv.Box{Name: name, Cpu: 1, RamMb: 512}, nil //nolint:mnd
}

func (fakeRegistryBoxLookup) ListBoxes(_ context.Context) ([]verv.Box, error) {
	return nil, nil
}

// TestDeployRegistryJob_ResolveAndOverlay exercises the same
// builtin.Read -> MergeEnvironment -> ResolveRequest -> overlay chain
// deployRegistryJob.Do runs, without any Docker/DB dependency, to guard the
// port-publication and auth-env overlay logic - the part of this handler
// most likely to silently regress (see overlayRegistryPort's doc comment on
// ExposedTo being required for host publication).
func TestDeployRegistryJob_ResolveAndOverlay(t *testing.T) {
	files, err := builtin.Read(registryDescriptorName)
	require.NoError(t, err)

	descriptor, err := vervonomicon.MergeEnvironment(files, registryEnvironment)
	require.NoError(t, err)

	descriptor.Source = verv.SourceKindBuiltin

	resolver := vervonomicon.NewBoxResolver(fakeRegistryBoxLookup{})

	request, err := resolver.ResolveRequest(context.Background(), descriptor, registryEnvironment, "")
	require.NoError(t, err)

	overlayRegistryPort(request, 15000) //nolint:mnd
	overlayRegistryAuthEnv(request)

	var exposedTo *uint32

	for _, p := range request.GetSettings().GetPorts() {
		if p.GetServicePortNumber() == registryContainerPort {
			exposedTo = p.ExposedTo
		}
	}

	require.NotNil(t, exposedTo)
	require.Equal(t, uint32(15000), *exposedTo) //nolint:mnd

	require.Equal(t, "htpasswd", request.GetEnv()[envRegistryAuth])
	require.Equal(t, registryAuthRealm, request.GetEnv()[envRegistryAuthHtpasswdRealm])
	require.Equal(t, registryHtpasswdPath, request.GetEnv()[envRegistryAuthHtpasswdPath])
}

func TestRegistryUrl_InContainerUsesInternalName(t *testing.T) {
	// env.IsInContainer() reads /.dockerenv/cgroup state, which is false in
	// the test runner - this exercises only the bare-binary branch, mirroring
	// enable_statefull_test.go's equivalent coverage gap for
	// applyBareBinaryHostPort.
	got := registryUrl(15000) //nolint:mnd
	require.Equal(t, "http://localhost:15000", got)
}
