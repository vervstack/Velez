package transfer_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/transfers_queries"
)

func (impl *Impl) ListTransfers(
	ctx context.Context,
	_ *pb.ListTransfers_Request,
) (*pb.ListTransfers_Response, error) {
	transfers, err := impl.transfersOut.List(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing transfers")
	}

	out := make([]*pb.TransferInfo, 0, len(transfers))
	for _, transfer := range transfers {
		out = append(out, transferToPb(transfer))
	}

	resp := &pb.ListTransfers_Response{Transfers: out}

	return resp, nil
}

func transferToPb(transfer domain.Transfer) *pb.TransferInfo {
	return &pb.TransferInfo{
		Id:            transfer.Uuid.String(),
		IsSource:      transfer.IsSource,
		ContainerName: transfer.ContainerName,
		PeerAddress:   transfer.PeerAddress,
		Status:        transferStatusToPb(transfer.Status),
		Error:         transfer.Error,
		TotalBytes:    transfer.TotalBytes,
		DoneBytes:     transfer.DoneBytes,
		CreatedAt:     timestamppb.New(transfer.CreatedAt),
		UpdatedAt:     timestamppb.New(transfer.UpdatedAt),
	}
}

func transferStatusToPb(status transfers_queries.VelezTransferStatus) pb.TransferStatus {
	switch status {
	case transfers_queries.VelezTransferStatusPENDING:
		return pb.TransferStatus_TRANSFER_STATUS_PENDING
	case transfers_queries.VelezTransferStatusRUNNING:
		return pb.TransferStatus_TRANSFER_STATUS_RUNNING
	case transfers_queries.VelezTransferStatusFAILED:
		return pb.TransferStatus_TRANSFER_STATUS_FAILED
	case transfers_queries.VelezTransferStatusDONE:
		return pb.TransferStatus_TRANSFER_STATUS_DONE
	case transfers_queries.VelezTransferStatusCANCELLED:
		return pb.TransferStatus_TRANSFER_STATUS_CANCELLED
	default:
		return pb.TransferStatus_TRANSFER_STATUS_UNSPECIFIED
	}
}
