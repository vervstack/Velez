package pgaas_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
)

func (impl *Impl) DropPgInstance(
	ctx context.Context,
	req *pb.DropPgInstance_Request,
) (*pb.DropPgInstance_Response, error) {
	err := impl.postgresService.DropPgInstance(ctx, req.GetName())
	if err != nil {
		return nil, rerrors.Wrap(err, "error dropping pg instance")
	}

	resp := &pb.DropPgInstance_Response{
		EntityId: req.GetName(),
		Action:   jobs.DropPgInstanceAction,
	}

	return resp, nil
}
