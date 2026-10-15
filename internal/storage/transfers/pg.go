// Package transfers provides the postgres backend for storage.TransfersStorage -
// see internal/storage/storage.go.
package transfers

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/transfers_queries"
)

type pgStorage struct {
	querier *transfers_queries.Queries
}

func NewPg(db *sql.DB) storage.TransfersStorage {
	return &pgStorage{
		querier: transfers_queries.New(db),
	}
}

func (p *pgStorage) CreateToken(ctx context.Context, tokenHash string, expiresAt time.Time) error {
	params := transfers_queries.InsertTransferTokenParams{
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}

	err := p.querier.InsertTransferToken(ctx, params)
	if err != nil {
		return rerrors.Wrap(err, "error inserting transfer token")
	}

	return nil
}

func (p *pgStorage) ConsumeToken(ctx context.Context, tokenHash string) (bool, error) {
	_, err := p.querier.ConsumeTransferToken(ctx, tokenHash)
	if err != nil {
		if rerrors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, rerrors.Wrap(err, "error consuming transfer token")
	}

	return true, nil
}

func (p *pgStorage) InsertTransfer(ctx context.Context, transfer domain.Transfer) (domain.Transfer, error) {
	params := transfers_queries.InsertTransferParams{
		IsSource:      transfer.IsSource,
		ContainerName: transfer.ContainerName,
		PeerAddress:   transfer.PeerAddress,
		Status:        transfer.Status,
	}

	row, err := p.querier.InsertTransfer(ctx, params)
	if err != nil {
		return domain.Transfer{}, rerrors.Wrap(err, "error inserting transfer")
	}

	return transferFromRow(row), nil
}

func (p *pgStorage) GetTransfer(ctx context.Context, id uuid.UUID) (sql.Null[domain.Transfer], error) {
	row, err := p.querier.GetTransfer(ctx, id)
	if err != nil {
		if rerrors.Is(err, sql.ErrNoRows) {
			return sql.Null[domain.Transfer]{}, nil
		}

		return sql.Null[domain.Transfer]{}, rerrors.Wrap(err, "error getting transfer")
	}

	return sql.Null[domain.Transfer]{V: transferFromRow(row), Valid: true}, nil
}

func (p *pgStorage) UpdateProgress(ctx context.Context, id uuid.UUID, totalBytes, doneBytes int64) error {
	params := transfers_queries.UpdateTransferProgressParams{
		ID:         id,
		TotalBytes: totalBytes,
		DoneBytes:  doneBytes,
	}

	err := p.querier.UpdateTransferProgress(ctx, params)
	if err != nil {
		return rerrors.Wrap(err, "error updating transfer progress")
	}

	return nil
}

func (p *pgStorage) Finish(
	ctx context.Context, id uuid.UUID, status transfers_queries.VelezTransferStatus, errMessage string,
) error {
	params := transfers_queries.FinishTransferParams{
		ID:     id,
		Status: status,
		Error:  errMessage,
	}

	err := p.querier.FinishTransfer(ctx, params)
	if err != nil {
		return rerrors.Wrap(err, "error finishing transfer")
	}

	return nil
}

func (p *pgStorage) List(ctx context.Context) ([]domain.Transfer, error) {
	rows, err := p.querier.ListTransfers(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing transfers")
	}

	out := make([]domain.Transfer, 0, len(rows))
	for _, row := range rows {
		out = append(out, transferFromRow(row))
	}

	return out, nil
}

func (p *pgStorage) FailRunning(ctx context.Context) error {
	err := p.querier.FailRunningTransfers(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error failing running transfers")
	}

	return nil
}

func transferFromRow(row transfers_queries.VelezTransfer) domain.Transfer {
	return domain.Transfer{
		Uuid:          row.ID,
		IsSource:      row.IsSource,
		ContainerName: row.ContainerName,
		PeerAddress:   row.PeerAddress,
		Status:        row.Status,
		Error:         row.Error,
		TotalBytes:    row.TotalBytes,
		DoneBytes:     row.DoneBytes,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}
