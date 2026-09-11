//go:build e2e_full

package e2e

import (
	"testing"

	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

const (
	networkIsoProdSuffix = "e2ecrtnet"
	networkIsoStageEnv   = "E2ECRTNETSTAGE"

	networkIsoProdName  = "e2e_ctr_net_prod"
	networkIsoStageName = "e2e_ctr_net_stage"

	listContainersProdSuffix = "e2ecrtlist"
	listContainersStageEnv   = "E2ECRTLISTSTAGE"

	listContainersProdName  = "e2e_ctr_runtime_list_prod"
	listContainersStageName = "e2e_ctr_runtime_list_stage"
)

// Test_ContainerRuntime_ListContainers_ScopesToEnvironment is a RED test for
// the ContainerRuntime.ListContainers method that backs the ListSmerds RPC
// (see docs/container_runtimes/roadmap.md's Phase 1 and
// internal/clients/node_clients/container_runtime/label_based.go's
// ListContainers doc comment).
//
// It creates one smerd in the default (PROD) environment and a differently
// named one in a second environment, then asserts that a ListSmerds scoped to
// the second environment sees exactly its own container and none of PROD's.
//
// It is EXPECTED TO FAIL today: ContainerManager.ListSmerds now routes through
// RuntimeResolver.Runtime(...).ListContainers instead of calling
// node_clients.Docker.ListContainers directly, and labelBasedRuntime's
// ListContainers is a deliberate stub that does not filter by suffix (see its
// doc comment) - the next agent's job is to make it correct. Until then this
// scoped list returns BOTH environments' containers instead of just the one
// requested.
func Test_ContainerRuntime_ListContainers_ScopesToEnvironment(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t,
		WithContainerSuffix(listContainersProdSuffix),
		WithEnvironments([]string{listContainersStageEnv}))

	prodReq := &velez_api.CreateSmerd_Request{
		Name:         listContainersProdName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
	}

	prodSmerd := env.CreateSmerd(t, prodReq)
	require.Equal(t, velez_api.Smerd_running, prodSmerd.GetStatus())

	stageReq := &velez_api.CreateSmerd_Request{
		Name:         listContainersStageName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Environment:  listContainersStageEnv,
	}

	stageSmerd := env.CreateSmerd(t, stageReq)
	require.Equal(t, velez_api.Smerd_running, stageSmerd.GetStatus())

	stageListReq := &velez_api.ListSmerds_Request{
		Environment: listContainersStageEnv,
		Label:       map[string]string{testCaseNameLabel: t.Name()},
	}

	stageList := env.ListSmerds(t, t.Context(), stageListReq)

	require.Len(t, stageList.GetSmerds(), 1,
		"a STAGE-scoped ListSmerds must not also return PROD's container - "+
			"labelBasedRuntime.ListContainers is a deliberate unfiltered stub, see its doc comment")
	require.Equal(t, stageSmerd.GetUuid(), stageList.GetSmerds()[0].GetUuid())

	// Smerd.Name is always the virtual/logical name - never the suffixed
	// Docker container name - both from CreateSmerd's response (stageSmerd,
	// via InspectSmerd) and from ListSmerds (stageList, via
	// ContainerRuntime.ListContainers). See
	// docs/container_runtimes/interface_design.md.
	require.Equal(t, listContainersStageName, stageSmerd.GetName())
	require.Equal(t, listContainersStageName, stageList.GetSmerds()[0].GetName())
}

// Test_ContainerRuntime_Network_PerEnvironmentIsolation covers Stage 4
// (docs/container_runtimes): each environment must get its OWN Docker bridge
// network ("verv_<suffix>") instead of every environment sharing the single
// hardcoded "verv" network env.StartNetwork creates once at node boot. It
// creates one smerd in PROD and one differently-named smerd in a second
// environment - both with a port binding, so createContainerJob's
// default-network connect actually runs (see label_based.go's
// CreateNetwork/ConnectToNetwork) - then asserts against the real Docker
// daemon that each container ends up on its own suffixed network and that
// neither network contains the other environment's container.
//
// This used to be RED for a structural reason beyond "the methods don't
// exist yet": dockerutils.CreateNetwork hardcoded the exact same IPAM subnet
// (10.0.1.0/24) for every network it created, regardless of name, which was
// fine when only one such network (the single shared "verv") ever existed on
// a node but made a second environment's own "verv_<suffix>" network collide
// with it ("Pool overlaps with other one on this address space"). Fixed by
// dropping the explicit Subnet from dockerutils.CreateNetwork's IPAM config
// and letting Docker's default IPAM driver auto-allocate a non-overlapping
// subnet per network - env.StartNetwork's existing "verv" network is
// untouched by that change (CreateNetwork no-ops when a network with the
// requested name already exists).
func Test_ContainerRuntime_Network_PerEnvironmentIsolation(t *testing.T) {
	t.Parallel()

	env := Planes[0].NewEnvironment(t,
		WithContainerSuffix(networkIsoProdSuffix),
		WithEnvironments([]string{networkIsoStageEnv}))

	prodReq := &velez_api.CreateSmerd_Request{
		Name:         networkIsoProdName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Settings: &velez_api.Container_Settings{
			Ports: []*velez_api.Port{{ServicePortNumber: 8080, Protocol: velez_api.Port_tcp}},
		},
	}

	prodSmerd := env.CreateSmerd(t, prodReq)
	require.Equal(t, velez_api.Smerd_running, prodSmerd.GetStatus())
	checkPorts(t, prodSmerd.GetPorts(), prodReq.GetSettings().GetPorts())

	stageReq := &velez_api.CreateSmerd_Request{
		Name:         networkIsoStageName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Environment:  networkIsoStageEnv,
		Settings: &velez_api.Container_Settings{
			Ports: []*velez_api.Port{{ServicePortNumber: 8080, Protocol: velez_api.Port_tcp}},
		},
	}

	stageSmerd := env.CreateSmerd(t, stageReq)
	require.Equal(t, velez_api.Smerd_running, stageSmerd.GetStatus())
	checkPorts(t, stageSmerd.GetPorts(), stageReq.GetSettings().GetPorts())

	dockerClient := env.Custom.NodeClients.Docker().Client()

	prodNetName := "verv_" + networkIsoProdSuffix
	stageNetName := "verv_" + networkIsoStageEnv

	inspectOpts := network.InspectOptions{Verbose: true}

	prodNet, err := dockerClient.NetworkInspect(t.Context(), prodNetName, inspectOpts)
	require.NoError(t, err, "expected PROD's own network %q to exist", prodNetName)

	stageNet, err := dockerClient.NetworkInspect(t.Context(), stageNetName, inspectOpts)
	require.NoError(t, err, "expected STAGE's own network %q to exist", stageNetName)

	_, prodInProdNet := prodNet.Containers[prodSmerd.GetUuid()]
	require.True(t, prodInProdNet, "expected PROD's container to be connected to %q", prodNetName)

	_, stageInProdNet := prodNet.Containers[stageSmerd.GetUuid()]
	require.False(t, stageInProdNet, "STAGE's container must not be connected to PROD's network")

	_, stageInStageNet := stageNet.Containers[stageSmerd.GetUuid()]
	require.True(t, stageInStageNet, "expected STAGE's container to be connected to %q", stageNetName)

	_, prodInStageNet := stageNet.Containers[prodSmerd.GetUuid()]
	require.False(t, prodInStageNet, "PROD's container must not be connected to STAGE's network")
}
