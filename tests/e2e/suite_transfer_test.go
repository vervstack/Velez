//go:build e2e_full

package e2e

import (
	"testing"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TransferSuite runs two Velez instances against two Docker daemons: the
// source on the shared DinD, the target on a second DinD (see
// transferTargetDockerHost), reachable over a loopback TCP port.
type TransferSuite struct {
	suite.Suite

	plane Plane
}

func (s *TransferSuite) Test_MoveContainerWithVolume() {
	t := s.T()
	pair := newTransferPair(t, s.plane)
	rootName, volumeName := transferNames(t)

	createVolumeRoot(t, pair.source, rootName, volumeName, "")
	sourceDigest := awaitSeededVolume(t, pair.source, volumeName)

	plan := pair.plan(t, rootName)
	require.Empty(t, plan.GetUnsupportedMounts())

	pair.moveSucceeds(t, rootName, nil, pair.issueToken(t), true)

	requireStopped(t, pair.target, rootName)
	requireStopped(t, pair.source, rootName)
	require.Equal(t, sourceDigest, volumeDigest(t, pair.target, volumeName))
}

func (s *TransferSuite) Test_Sidecars_OnlySelectedMove() {
	t := s.T()
	pair := newTransferPair(t, s.plane)
	rootName, volumeName := transferNames(t)
	keptName := rootName + "_kept"
	leftName := rootName + "_left"

	root := createVolumeRoot(t, pair.source, rootName, volumeName, "")
	createTransferSidecar(t, pair.source, keptName, root.ID)
	createTransferSidecar(t, pair.source, leftName, root.ID)

	plan := pair.plan(t, rootName)
	require.Len(t, plan.GetSidecars(), 2)

	names := []string{plan.GetSidecars()[0].GetContainerName(), plan.GetSidecars()[1].GetContainerName()}
	require.ElementsMatch(t, []string{keptName, leftName}, names)

	pair.moveSucceeds(t, rootName, []string{keptName}, pair.issueToken(t), true)

	movedRoot := requireStopped(t, pair.target, rootName)
	movedSidecar := requireStopped(t, pair.target, keptName)
	require.Equal(t, container.NetworkMode("container:"+movedRoot.ID), movedSidecar.HostConfig.NetworkMode)

	_, err := dockerOf(pair.target).ContainerInspect(t.Context(), leftName)
	require.True(t, errdefs.IsNotFound(err), "unselected sidecar must not reach the target: %v", err)
}

func (s *TransferSuite) Test_Consumers_RequireAcknowledgement() {
	t := s.T()
	pair := newTransferPair(t, s.plane)
	rootName, volumeName := transferNames(t)
	consumerName := rootName + "_consumer"
	networkName := rootName + "_net"

	createUserNetwork(t, pair.source, networkName)
	createVolumeRoot(t, pair.source, rootName, volumeName, networkName)
	createNetworkConsumer(t, pair.source, consumerName, networkName)

	plan := pair.plan(t, rootName)
	require.Len(t, plan.GetConsumers(), 1)
	require.Equal(t, consumerName, plan.GetConsumers()[0].GetContainerName())
	require.Equal(t, networkName, plan.GetConsumers()[0].GetNetworkName())

	token := pair.issueToken(t)

	_, err := pair.start(t, rootName, nil, token, false)
	require.Equal(t, codes.FailedPrecondition, status.Code(err), "error: %v", err)
	requireRunning(t, pair.source, rootName)

	pair.moveSucceeds(t, rootName, nil, token, true)

	requireStopped(t, pair.target, rootName)
}

func (s *TransferSuite) Test_Token_IsSingleUse() {
	t := s.T()
	pair := newTransferPair(t, s.plane)
	firstName, firstVolume := transferNames(t)
	secondName := firstName + "_second"
	secondVolume := firstVolume + "_second"

	createVolumeRoot(t, pair.source, firstName, firstVolume, "")
	createVolumeRoot(t, pair.source, secondName, secondVolume, "")

	token := pair.issueToken(t)

	pair.moveSucceeds(t, firstName, nil, token, true)

	_, err := pair.start(t, secondName, nil, token, true)
	require.Equal(t, codes.PermissionDenied, status.Code(err), "error: %v", err)
	requireRunning(t, pair.source, secondName)
}

func (s *TransferSuite) Test_BindMounts_BlockTransfer() {
	t := s.T()
	pair := newTransferPair(t, s.plane)
	rootName, _ := transferNames(t)
	hostPath := "/tmp"

	createBindRoot(t, pair.source, rootName, hostPath)

	plan := pair.plan(t, rootName)
	require.Equal(t, []string{hostPath}, plan.GetUnsupportedMounts())

	_, err := pair.start(t, rootName, nil, pair.issueToken(t), true)
	require.Equal(t, codes.FailedPrecondition, status.Code(err), "error: %v", err)
	requireRunning(t, pair.source, rootName)
}

func (s *TransferSuite) Test_TargetNameCollision_LeavesSourceRunning() {
	t := s.T()
	pair := newTransferPair(t, s.plane)
	rootName, volumeName := transferNames(t)

	createVolumeRoot(t, pair.source, rootName, volumeName, "")
	createCollidingContainer(t, pair.target, rootName)

	_, err := pair.start(t, rootName, nil, pair.issueToken(t), true)
	require.Equal(t, codes.AlreadyExists, status.Code(err), "error: %v", err)
	requireRunning(t, pair.source, rootName)
}

func Test_Transfer(t *testing.T) {
	t.Parallel()

	planes := []Plane{singleNodePlane()}

	RunPlaneSuite(t, planes, func(plane Plane) suite.TestingSuite {
		return &TransferSuite{plane: plane}
	})
}
