package pgaas_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/transport/common"
)

func (impl *Impl) ListPgInstances(
	ctx context.Context,
	req *pb.ListPgInstances_Request,
) (*pb.ListPgInstances_Response, error) {
	serviceReq := domain.ListPgInstancesReq{
		Paging: common.FromPaging(req.GetPaging()),
	}

	list, err := impl.postgresService.ListPgInstances(ctx, serviceReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing pg instances")
	}

	out := make([]*pb.PgInstance, 0, len(list.Instances))
	for _, instance := range list.Instances {
		out = append(out, pgInstanceToPb(instance))
	}

	resp := &pb.ListPgInstances_Response{
		Instances: out,
		Total:     list.Total,
	}

	return resp, nil
}

// pgInstanceToPb maps the domain view onto the wire type. Shared by every
// pg instance RPC in this package. Never populates a password field -
// pb.PgInstance deliberately has none; only GetPgInstanceCredentials
// resolves the secret.
func pgInstanceToPb(view domain.PgInstanceView) *pb.PgInstance {
	out := &pb.PgInstance{
		Name:        view.Name,
		DbName:      view.DbName,
		Username:    view.Username,
		Port:        uint32(view.Port),
		Environment: view.Environment,
		Status:      view.Status,
	}

	if view.OwnerService != "" {
		out.OwnerService = &view.OwnerService
	}

	if !view.CreatedAt.IsZero() {
		out.CreatedAt = timestamppb.New(view.CreatedAt)
	}

	if !view.UpdatedAt.IsZero() {
		out.UpdatedAt = timestamppb.New(view.UpdatedAt)
	}

	return out
}
