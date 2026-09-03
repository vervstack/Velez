package postgres

import (
	"context"
	"database/sql"
	"strconv"

	"go.redsock.ru/rerrors"

	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/resource_boxes_queries"
)

// resourceBoxesStorage implements
// internal/service/service_manager/vervonomicon.BoxLookup over
// velez.resource_boxes.
type resourceBoxesStorage struct {
	querier resource_boxes_queries.Querier
}

func newResourceBoxesStorage(db *sql.DB) *resourceBoxesStorage {
	return &resourceBoxesStorage{
		querier: resource_boxes_queries.New(db),
	}
}

func (s *resourceBoxesStorage) GetBox(ctx context.Context, name string) (verv.Box, error) {
	row, err := s.querier.GetResourceBoxByName(ctx, name)
	if err != nil {
		return verv.Box{}, wrapPgErr(err)
	}

	return boxFromRow(row)
}

func (s *resourceBoxesStorage) ListBoxes(ctx context.Context) ([]verv.Box, error) {
	rows, err := s.querier.ListResourceBoxes(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing resource boxes")
	}

	result := make([]verv.Box, len(rows))

	for i, row := range rows {
		box, err := boxFromRow(row)
		if err != nil {
			return nil, err
		}

		result[i] = box
	}

	return result, nil
}

// boxFromRow converts a generated row into the domain type. Cpu is stored as
// Postgres NUMERIC, which sqlc scans as a string.
func boxFromRow(row resource_boxes_queries.VelezResourceBox) (verv.Box, error) {
	cpu, err := strconv.ParseFloat(row.Cpu, 64)
	if err != nil {
		return verv.Box{}, rerrors.Wrap(err, "error parsing box cpu '"+row.Cpu+"'")
	}

	box := verv.Box{
		Name:      row.Name,
		Cpu:       cpu,
		RamMb:     row.RamMb,
		DiskMb:    row.DiskMb,
		IsBuiltin: row.IsBuiltin,
	}

	return box, nil
}
