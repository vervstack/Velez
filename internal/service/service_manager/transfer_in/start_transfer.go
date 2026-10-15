package transfer_in

import (
	"context"

	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/peer"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/transfers_queries"
)

func (s *Service) startTransfer(ctx context.Context, manifest transferManifest) (domain.Transfer, error) {
	transfer := domain.Transfer{
		IsSource:      false,
		ContainerName: manifest.roots()[0].name,
		PeerAddress:   peerAddress(ctx),
		Status:        transfers_queries.VelezTransferStatusRUNNING,
		TotalBytes:    manifest.totalBytes,
	}

	inserted, err := s.dataStorage.Transfers().InsertTransfer(ctx, transfer)
	if err != nil {
		return domain.Transfer{}, rerrors.Wrap(err, "error inserting transfer")
	}

	return inserted, nil
}

func peerAddress(ctx context.Context) string {
	p, isFound := peer.FromContext(ctx)
	if !isFound || p.Addr == nil {
		return ""
	}

	return p.Addr.String()
}
