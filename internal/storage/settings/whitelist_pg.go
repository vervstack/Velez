package settings

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/whitelisted_images_queries"
)

type pgWhitelist struct {
	querier *whitelisted_images_queries.Queries
}

func NewPgWhitelist(db *sql.DB) storage.WhitelistStorage {
	return &pgWhitelist{
		querier: whitelisted_images_queries.New(db),
	}
}

func (p *pgWhitelist) List(ctx context.Context) ([]domain.WhitelistedImage, error) {
	rows, err := p.querier.ListWhitelistedImages(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing whitelisted images")
	}

	images := make([]domain.WhitelistedImage, 0, len(rows))

	for _, row := range rows {
		image := domain.WhitelistedImage{
			Id:         row.ID,
			Repository: row.Repository,
		}

		images = append(images, image)
	}

	return images, nil
}

func (p *pgWhitelist) Add(ctx context.Context, repository string) (uuid.UUID, bool, error) {
	id, err := p.querier.InsertWhitelistedImage(ctx, repository)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, false, nil
	}

	if err != nil {
		return uuid.Nil, false, rerrors.Wrap(err, "error adding whitelisted image")
	}

	return id, true, nil
}

func (p *pgWhitelist) Remove(ctx context.Context, id uuid.UUID) error {
	err := p.querier.DeleteWhitelistedImage(ctx, id)
	if err != nil {
		return rerrors.Wrap(err, "error removing whitelisted image")
	}

	return nil
}

func (p *pgWhitelist) RemoveByRepository(ctx context.Context, repository string) error {
	err := p.querier.DeleteWhitelistedImageByRepository(ctx, repository)
	if err != nil {
		return rerrors.Wrap(err, "error removing whitelisted image by repository")
	}

	return nil
}
