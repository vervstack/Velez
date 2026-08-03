package ports

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"
)

// availablePortsForTest picks a high, unlikely-to-be-bound range. GetPort
// actually dials each candidate, so these tests stick to LockPort where the
// pool state is deterministic, except where dialing is explicitly wanted.
func availablePortsForTest() []int {
	return []int{45101, 45102, 45103}
}

// Environments share one pool: a port locked by environment A must not be
// lockable by environment B. This is the core collision-awareness requirement -
// there are deliberately NO per-environment reserved ranges.
func TestLockPortForEnvironment_CrossEnvironmentCollisionRejected(t *testing.T) {
	pm := NewPortManager(availablePortsForTest(), nil)

	err := pm.LockPortForEnvironment("prod", 45101)
	require.NoError(t, err)

	err = pm.LockPortForEnvironment("stage", 45101)
	require.Error(t, err)
	require.True(t, rerrors.Is(err, ErrPortAlreadyLocked))
	require.Contains(t, err.Error(), "prod", "error should name the owning environment")
}

// Same-environment double-lock is still an error - the pre-existing
// ErrPortAlreadyLocked contract is unchanged, only enriched.
func TestLockPortForEnvironment_SameEnvironmentDoubleLockRejected(t *testing.T) {
	pm := NewPortManager(availablePortsForTest(), nil)

	require.NoError(t, pm.LockPortForEnvironment("prod", 45101))

	err := pm.LockPortForEnvironment("prod", 45101)
	require.True(t, rerrors.Is(err, ErrPortAlreadyLocked))
}

func TestLockPortForEnvironment_RecordsOwner(t *testing.T) {
	pm := NewPortManager(availablePortsForTest(), nil)

	require.NoError(t, pm.LockPortForEnvironment("stage", 45102))

	owner, held := pm.PortOwner(45102)
	require.True(t, held)
	require.Equal(t, "stage", owner)
}

// Unlocking clears ownership so another environment can take the port.
func TestUnlockPorts_ClearsOwnership(t *testing.T) {
	pm := NewPortManager(availablePortsForTest(), nil)

	require.NoError(t, pm.LockPortForEnvironment("prod", 45103))

	pm.UnlockPorts([]uint32{45103})

	_, held := pm.PortOwner(45103)
	require.False(t, held)

	require.NoError(t, pm.LockPortForEnvironment("stage", 45103))
}

// A partially-failed multi-port lock must roll back every port it took, owner
// entries included - otherwise a retry from a different environment would hit
// a phantom collision.
func TestLockPortForEnvironment_RollsBackOwnershipOnFailure(t *testing.T) {
	pm := NewPortManager(availablePortsForTest(), nil)

	require.NoError(t, pm.LockPortForEnvironment("prod", 45103))

	// 45101 is free, 45103 is taken - the whole call must fail and release 45101.
	err := pm.LockPortForEnvironment("stage", 45101, 45103)
	require.Error(t, err)

	_, held := pm.PortOwner(45101)
	require.False(t, held, "45101 must have been rolled back")

	require.NoError(t, pm.LockPortForEnvironment("qa", 45101))
}

// The unscoped API keeps working and is just LockPortForEnvironment with the
// UnscopedEnvironment owner.
func TestLockPort_UnscopedStillCollidesWithScoped(t *testing.T) {
	pm := NewPortManager(availablePortsForTest(), nil)

	require.NoError(t, pm.LockPort(45101))

	owner, held := pm.PortOwner(45101)
	require.True(t, held)
	require.Equal(t, UnscopedEnvironment, owner)

	err := pm.LockPortForEnvironment("prod", 45101)
	require.True(t, rerrors.Is(err, ErrPortAlreadyLocked))
}

// Ports reported as already occupied on the host are owned by nobody in
// particular, but still block every environment.
func TestNewPortManager_PreOccupiedPortsBlockEveryEnvironment(t *testing.T) {
	pm := NewPortManager(availablePortsForTest(), []uint32{45101})

	err := pm.LockPortForEnvironment("prod", 45101)
	require.True(t, rerrors.Is(err, ErrPortAlreadyLocked))
}

func TestLockPortForEnvironment_UnknownPortRejected(t *testing.T) {
	pm := NewPortManager(availablePortsForTest(), nil)

	err := pm.LockPortForEnvironment("prod", 1)
	require.True(t, rerrors.Is(err, ErrUnavailablePort))
}

// GetPortForEnvironment stamps the caller's environment onto whatever port it
// hands out, so a subsequent cross-environment lock of that port collides.
func TestGetPortForEnvironment_RecordsOwner(t *testing.T) {
	pm := NewPortManager(availablePortsForTest(), nil)

	port, err := pm.GetPortForEnvironment("prod")
	require.NoError(t, err)

	owner, held := pm.PortOwner(port)
	require.True(t, held)
	require.Equal(t, "prod", owner)

	err = pm.LockPortForEnvironment("stage", port)
	require.True(t, rerrors.Is(err, ErrPortAlreadyLocked))
}

// The Container wrapper must forward the environment-scoped API too, or the
// swap-capable indirection would silently drop ownership information.
func TestContainer_ForwardsEnvironmentScopedAPI(t *testing.T) {
	c := NewContainer(NewPortManager(availablePortsForTest(), nil))

	require.NoError(t, c.LockPortForEnvironment("prod", 45102))

	owner, held := c.PortOwner(45102)
	require.True(t, held)
	require.Equal(t, "prod", owner)

	err := c.LockPortForEnvironment("stage", 45102)
	require.Error(t, err)
}
