package velez_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) RegisterContainer(
	_ context.Context,
	_ *velez_api.RegisterContainer_Request,
) (*velez_api.RegisterContainer_Response, error) {
	return nil, rerrors.Wrap(errRegisterContainerNotImplemented)
}
