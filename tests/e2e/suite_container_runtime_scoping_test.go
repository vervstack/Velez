//go:build e2e_full

package e2e

import (
	"testing"

	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

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

func newListContainersProdRequest() *velez_api.CreateSmerd_Request {
	return &velez_api.CreateSmerd_Request{
		Name:         listContainersProdName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
	}
}

func newListContainersStageRequest() *velez_api.CreateSmerd_Request {
	return &velez_api.CreateSmerd_Request{
		Name:         listContainersStageName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Environment:  listContainersStageEnv,
	}
}

func newNetworkIsoProdRequest() *velez_api.CreateSmerd_Request {
	return &velez_api.CreateSmerd_Request{
		Name:         networkIsoProdName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Settings: &velez_api.Container_Settings{
			Ports: []*velez_api.Port{{ServicePortNumber: 8080, Protocol: velez_api.Port_tcp}},
		},
	}
}

func newNetworkIsoStageRequest() *velez_api.CreateSmerd_Request {
	return &velez_api.CreateSmerd_Request{
		Name:         networkIsoStageName,
		ImageName:    HelloWorldAppImage,
		IgnoreConfig: true,
		Environment:  networkIsoStageEnv,
		Settings: &velez_api.Container_Settings{
			Ports: []*velez_api.Port{{ServicePortNumber: 8080, Protocol: velez_api.Port_tcp}},
		},
	}
}

// ContainerRuntimeScopingSuite covers the ContainerRuntime.ListContainers
// environment scoping (docs/container_runtimes/roadmap.md Phase 1) and the
// per-environment Docker network isolation (Stage 4).
type ContainerRuntimeScopingSuite struct {
	suite.Suite

	plane Plane
}

// Test_ListContainers_ScopesToEnvironment is a RED test for
// ContainerManager.ListSmerds, which now routes through
// RuntimeResolver.Runtime(...).ListContainers instead of calling
// node_clients.Docker.ListContainers directly: labelBasedRuntime.ListContainers
// is a deliberate unfiltered stub (see its doc comment), so a STAGE-scoped
// ListSmerds today also returns PROD's container until that stub is fixed.
func (s *ContainerRuntimeScopingSuite) Test_ListContainers_ScopesToEnvironment() {
	t := s.T()
	t.Parallel()

	env := s.plane.NewEnvironment(t,
		WithContainerSuffix(listContainersProdSuffix),
		WithEnvironments([]string{listContainersStageEnv}))

	prodReq := newListContainersProdRequest()

	prodSmerd := env.CreateSmerd(t, prodReq)
	require.Equal(t, velez_api.Smerd_running, prodSmerd.GetStatus())

	stageReq := newListContainersStageRequest()

	stageSmerd := env.CreateSmerd(t, stageReq)
	require.Equal(t, velez_api.Smerd_running, stageSmerd.GetStatus())

	stageListReq := &velez_api.ListSmerds_Request{
		Environment: listContainersStageEnv,
		Label:       map[string]string{testCaseNameLabel: t.Name()},
	}

	stageList := env.ListSmerds(t, t.Context(), stageListReq)

	require.Len(t, stageList.GetSmerds(), 1,
		"a STAGE-scoped ListSmerds must not also return PROD's container")
	require.Equal(t, stageSmerd.GetUuid(), stageList.GetSmerds()[0].GetUuid())

	// Smerd.Name is always the virtual/logical name, never the suffixed
	// Docker container name - see docs/container_runtimes/interface_design.md.
	require.Equal(t, listContainersStageName, stageSmerd.GetName())
	require.Equal(t, listContainersStageName, stageList.GetSmerds()[0].GetName())
}

// Test_Network_PerEnvironmentIsolation covers Stage 4
// (docs/container_runtimes): each environment must get its own Docker bridge
// network ("verv_<suffix>") instead of every environment sharing the single
// hardcoded "verv" network env.StartNetwork creates once at node boot.
func (s *ContainerRuntimeScopingSuite) Test_Network_PerEnvironmentIsolation() {
	t := s.T()
	t.Parallel()

	env := s.plane.NewEnvironment(t,
		WithContainerSuffix(networkIsoProdSuffix),
		WithEnvironments([]string{networkIsoStageEnv}))

	prodReq := newNetworkIsoProdRequest()

	prodSmerd := env.CreateSmerd(t, prodReq)
	require.Equal(t, velez_api.Smerd_running, prodSmerd.GetStatus())
	checkPorts(t, prodSmerd.GetPorts(), prodReq.GetSettings().GetPorts())

	stageReq := newNetworkIsoStageRequest()

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

func Test_ContainerRuntimeScoping(t *testing.T) {
	t.Parallel()
	RunPlaneSuite(t, Planes, func(plane Plane) suite.TestingSuite {
		return &ContainerRuntimeScopingSuite{plane: plane}
	})
}
