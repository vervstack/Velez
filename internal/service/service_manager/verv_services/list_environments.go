package verv_services

import (
	"context"

	"go.redsock.ru/rerrors"
)

func (v *VervService) ListEnvironments(ctx context.Context) ([]string, error) {
	envs, err := v.environmentsStorage.ListEnvironments(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing environments frpm storage")
	}

	return envs, nil
}
