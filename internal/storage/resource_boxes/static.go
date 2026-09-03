// Package resource_boxes provides the in-memory storage.ResourceBoxesStorage
// backend local_storage (single-node/dev mode, no postgres) uses, mirroring
// internal/storage/environments and internal/storage/registries.
package resource_boxes

import (
	"context"

	"go.redsock.ru/rerrors"

	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
	"go.vervstack.ru/Velez/internal/storage"
)

const (
	boxSmall  = "small"
	boxMedium = "medium"
	boxLarge  = "large"

	// Builtin tier sizing - mirrors
	// migrations/20260902130000_resource_boxes.sql's seed values exactly.
	smallCpu   = 0.5
	smallRam   = 512
	smallDisk  = 2048
	mediumCpu  = 1.0
	mediumRam  = 1024
	mediumDisk = 8192
	largeCpu   = 2.0
	largeRam   = 4096
	largeDisk  = 32768
)

// staticStorage is a read-only in-memory implementation of
// storage.ResourceBoxesStorage, seeded with the same three builtin tiers
// migrations/20260902130000_resource_boxes.sql seeds in cluster mode. There
// is no way to add a custom box in single-node/dev mode today - that's the
// same gap local_storage's other static fallbacks accept (see
// registries.NewStatic).
type staticStorage struct {
	boxes map[string]verv.Box
}

// NewStatic builds an in-memory resource-boxes storage seeded with the three
// builtin tiers.
func NewStatic() storage.ResourceBoxesStorage {
	return &staticStorage{
		boxes: map[string]verv.Box{
			boxSmall:  {Name: boxSmall, Cpu: smallCpu, RamMb: smallRam, DiskMb: smallDisk, IsBuiltin: true},
			boxMedium: {Name: boxMedium, Cpu: mediumCpu, RamMb: mediumRam, DiskMb: mediumDisk, IsBuiltin: true},
			boxLarge:  {Name: boxLarge, Cpu: largeCpu, RamMb: largeRam, DiskMb: largeDisk, IsBuiltin: true},
		},
	}
}

func (s *staticStorage) GetBox(_ context.Context, name string) (verv.Box, error) {
	box, ok := s.boxes[name]
	if !ok {
		return verv.Box{}, rerrors.Wrap(storage.ErrNotFound)
	}

	return box, nil
}

func (s *staticStorage) ListBoxes(_ context.Context) ([]verv.Box, error) {
	out := make([]verv.Box, 0, len(s.boxes))
	for _, box := range s.boxes {
		out = append(out, box)
	}

	return out, nil
}
