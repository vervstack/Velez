package user_errors

import (
	"go.redsock.ru/rerrors"
)

// ErrClusterStatePgHasNoState is returned by cluster_state.SetupMasterPg when
// the cluster-state postgres container exists but Docker reports a nil
// State - an inspect-result invariant violation, not something the caller
// can recover from.
var ErrClusterStatePgHasNoState = rerrors.New("postgres container for cluster state exists but has no state")
