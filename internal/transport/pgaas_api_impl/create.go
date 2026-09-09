package pgaas_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func (impl *Impl) CreatePgInstance(
	ctx context.Context,
	req *pb.CreatePgInstance_Request,
) (*pb.CreatePgInstance_Response, error) {
	serviceReq := domain.CreatePgInstanceReq{
		Name:         req.GetName(),
		Environment:  req.GetEnvironment(),
		Box:          req.GetBox(),
		ExposeToPort: req.GetExposeToPort(),
		OwnerService: req.GetOwnerService(),
		// The wire request carries no isolation field yet - see
		// domain.CreatePgInstanceReq.Isolation's doc comment. Every request
		// through this transport asks for separate_instance, the only
		// provisionable isolation mode today.
		Isolation: pb.PgInstanceIsolation_PG_INSTANCE_ISOLATION_SEPARATE_INSTANCE,
	}

	instance, err := impl.postgresService.CreatePgInstance(ctx, serviceReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating pg instance")
	}

	resp := &pb.CreatePgInstance_Response{
		Instance: pgInstanceToPb(instance),
	}

	return resp, nil
}
