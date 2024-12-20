package velez_api_impl

import (
	"context"

	errors "go.redsock.ru/rerrors"

	"github.com/godverv/Velez/pkg/velez_api"
)

func (a *Impl) CreateSmerd(ctx context.Context, req *velez_api.CreateSmerd_Request) (*velez_api.Smerd, error) {
	id, err := a.srv.LaunchSmerd(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "error launching smerd")
	}

	smerd, err := a.srv.InspectSmerd(ctx, id)
	if err != nil {
		return nil, err
	}

	return smerd, nil
}
