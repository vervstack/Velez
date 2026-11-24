package jobs

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

func TestChangePgMasterStrategy(t *testing.T) {
	t.Parallel()

	req := domain.ChangePgMasterReq{
		BalancerInstance: "balancer", NewMasterInstance: "pg-standby", OldMasterStrategy: pgFailoverRewind,
	}

	err := ValidateChangePgMaster(req)
	require.NoError(t, err)

	req.OldMasterStrategy = pgFailoverRecopy
	err = ValidateChangePgMaster(req)
	require.NoError(t, err)

	req.ConfirmOutOfSync = true
	err = ValidateChangePgMaster(req)
	require.ErrorIs(t, err, user_errors.ErrPgFailoverNoOverride)

	req.ConfirmOutOfSync = false
	req.OldMasterStrategy = "unknown"
	err = ValidateChangePgMaster(req)
	require.ErrorIs(t, err, user_errors.ErrPgFailoverStrategy)
}

func TestChangePgMasterInSyncGate(t *testing.T) {
	t.Parallel()

	const upstream = "primary"

	err := checkPgFailoverSync(upstream, upstream, pgStreamingState, true)
	require.NoError(t, err)

	err = checkPgFailoverSync(upstream, upstream, pgStreamingState, false)
	require.ErrorIs(t, err, user_errors.ErrPgFailoverOutOfSync)

	err = checkPgFailoverSync(upstream, upstream, "catchup", true)
	require.ErrorIs(t, err, user_errors.ErrPgFailoverOutOfSync)

	err = checkPgFailoverSync(upstream, "other", pgStreamingState, true)
	require.ErrorIs(t, err, user_errors.ErrPgFailoverUpstream)
}

func TestChangePgMasterReintegrateResume(t *testing.T) {
	t.Parallel()

	payload := &velez_api.SetupPgReplicationTaskPayload{SourceInstance: "new-primary", OperationId: "current-operation"}

	require.True(t, isPgRewindComplete(payload, payload.GetSourceInstance(), payload.GetOperationId()))
	require.False(t, isPgRewindComplete(payload, payload.GetSourceInstance(), "previous-operation"))
	require.False(t, isPgRewindComplete(payload, "other-primary", payload.GetOperationId()))
	require.False(t, isPgRewindComplete(payload, "", ""))
}
