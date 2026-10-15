package transfer_out

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/storage/postgres/generated/transfers_queries"
)

func Test_FailureOutcome_CancelledWins(t *testing.T) {
	status, text := failureOutcome(errTargetClosedEarly, true)

	require.Equal(t, transfers_queries.VelezTransferStatusCANCELLED, status)
	require.Equal(t, cancelledText, text)
}

func Test_FailureOutcome_KeepsOnlyTheCode(t *testing.T) {
	err := rerrors.Wrap(errTargetClosedEarly, "peer said: secret details")

	status, text := failureOutcome(err, false)

	require.Equal(t, transfers_queries.VelezTransferStatusFAILED, status)
	require.Equal(t, "transfer failed: aborted", text)
	require.NotContains(t, text, "secret")
}
