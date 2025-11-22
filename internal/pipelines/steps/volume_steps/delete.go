package volume_steps

import (
	"context"
)

type deleteVolume struct {
}

func Delete() *deleteVolume {
	return &deleteVolume{}
}

func (s *deleteVolume) Do(ctx context.Context) error {
	return nil
}
