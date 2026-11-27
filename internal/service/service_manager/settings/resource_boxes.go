package settings

import (
	"context"
	"slices"
	"strings"

	"go.redsock.ru/rerrors"

	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
	"go.vervstack.ru/Velez/internal/storage/resource_boxes"
)

func (s *SettingsServiceImpl) ListResourceBoxes(ctx context.Context) ([]verv.Box, error) {
	boxStorage := s.dataStorage.ResourceBoxes()
	if boxStorage == nil {
		boxStorage = resource_boxes.NewStatic()
	}

	boxes, err := boxStorage.ListBoxes(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing resource boxes")
	}

	slices.SortFunc(boxes, func(a, b verv.Box) int { return strings.Compare(a.Name, b.Name) })

	return boxes, nil
}
