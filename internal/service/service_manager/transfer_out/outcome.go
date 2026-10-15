package transfer_out

import (
	"strings"

	"google.golang.org/grpc/status"

	"go.vervstack.ru/Velez/internal/storage/postgres/generated/transfers_queries"
)

const (
	cancelledText = "cancelled by user"
	failedText    = "transfer failed: "
)

// failureOutcome picks the stored status and a short text; the error itself can carry
// the target's words, so only its code is kept.
func failureOutcome(err error, isCancelled bool) (transfers_queries.VelezTransferStatus, string) {
	if isCancelled {
		return transfers_queries.VelezTransferStatusCANCELLED, cancelledText
	}

	code := strings.ToLower(status.Code(err).String())

	return transfers_queries.VelezTransferStatusFAILED, failedText + code
}
