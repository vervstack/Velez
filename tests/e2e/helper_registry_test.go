//go:build e2e_full

package e2e

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/storage/environments"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/tasks_queries"
)

const (
	// registryLoaderContainerName mirrors copy_to_volume.go's
	// "<volume>_loader" naming (registryAuthVolumeName +
	// loaderContainerSuffix), duplicated here as a literal since both are
	// unexported internal/jobs constants.
	registryLoaderContainerName = "registry-auth_loader"

	// registryAuthVolumeName mirrors enable_registry.go's own (unexported)
	// registryAuthVolumeName constant.
	registryAuthVolumeName = "registry-auth"
)

// enableRegistryUnderDind runs EnablePlugin(registry) to completion, in
// plain single-node/local_storage mode (no WithMatreshka/WithClusterPgDsn) -
// this mirrors the bug report this suite covers ("doesn't show up on
// control plane page in single mode"). Unlike enableStatefullPgUnderDind,
// jobs.RegistryServiceName carries no per-suite suffix (see enable_registry.go's
// doc comment: a node only ever runs one registry), so collision-safety here
// is unconditional force-removal of the fixed container/volume names instead
// of a suffixed name - the same trade-off enable_statefull.go's sidecar
// makes, just without a suffix to vary.
//
// EnablePlugin's own task only gets as far as writing a SCHEDULED_DEPLOYMENT
// row (deployRegistryJob hands off to CreateNewDeploy - see its doc comment);
// the deploy watcher (internal/workers/deploy_watcher.go) is a separate
// ticker-driven worker that actually enqueues and runs create_smerd against
// that row. So after the enable_registry task reaches DONE, this helper also
// watches the create_smerd task the deploy watcher dispatches - entity id
// jobs.SmerdEntityID(environments.DefaultEnvironmentName, jobs.RegistryServiceName),
// per deployRegistryJob's registryEnvironment - before returning, so the
// caller can rely on the registry container actually existing.
func enableRegistryUnderDind(t *testing.T) *TestEnvironment {
	t.Helper()

	env := NewEnvironment(t)
	dockerClient := env.Custom.NodeClients.Docker().Client()

	removeRegistrySidecar(dockerClient)
	t.Cleanup(func() { removeRegistrySidecar(dockerClient) })

	// No ExposeToPort: deployRegistryJob.resolvePort auto-assigns and holds a
	// port from the node's shared PortManager pool when none is requested
	// (see its doc comment) - the same atomic pool every other test's smerd
	// port comes from, so it can't collide with one another suite's parallel
	// subtest hands out concurrently. A fixed dindRegistryPort carved out of
	// that pool doesn't work here: unlike the statefull_pg/headscale
	// sidecars (created directly via the docker client), the registry goes
	// through the real create_smerd job, whose prepareSmerdVervConfigJob
	// re-locks the exposed port through the same PortManager - which
	// rejects any port outside its known pool with ErrUnavailablePort. The
	// suite recovers the port PortManager picked from the registries row's
	// Url column after enabling (see registryHost in
	// suite_enable_registry_test.go).
	registryReq := &velez_api.EnableRegistry{}
	payload := &velez_api.EnablePlugin_Request_Registry{Registry: registryReq}
	req := &velez_api.EnablePlugin_Request{
		Plugin:  velez_api.VervPluginType_registry,
		Payload: payload,
	}

	resp, err := env.Custom.ControlPlaneApiImpl.EnablePlugin(t.Context(), req)
	require.NoError(t, err)

	var enableTask tasks_queries.VelezTask

	for task := range env.Custom.JobsEngine.Watch(t.Context(), resp.GetEntityId(), resp.GetAction()) {
		enableTask = task
	}

	require.Equal(t, tasks_queries.VelezTaskStatusDONE, enableTask.Status,
		"enable registry task error: %s", enableTask.Error.String)

	smerdEntityID := jobs.SmerdEntityID(environments.DefaultEnvironmentName, jobs.RegistryServiceName)

	var deployTask tasks_queries.VelezTask

	for task := range env.Custom.JobsEngine.Watch(t.Context(), smerdEntityID, jobs.CreateSmerdAction) {
		deployTask = task
	}

	require.Equal(t, tasks_queries.VelezTaskStatusDONE, deployTask.Status,
		"create_smerd task for registry error: %s", deployTask.Error.String)

	return env
}

// removeRegistrySidecar force-removes the fixed-name registry container, its
// leftover htpasswd loader container (only present if a prior run crashed
// mid-BuildJobs before stepDropContainer ran), and the registry-auth named
// volume, ignoring "no such container/volume".
func removeRegistrySidecar(dockerClient client.APIClient) {
	ctx := context.Background()
	removeOpts := container.RemoveOptions{Force: true}

	_ = dockerClient.ContainerRemove(ctx, jobs.RegistryServiceName, removeOpts)
	_ = dockerClient.ContainerRemove(ctx, registryLoaderContainerName, removeOpts)
	_ = dockerClient.VolumeRemove(ctx, registryAuthVolumeName, true)
}
