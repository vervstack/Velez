package control_plane_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func (impl *Impl) ListEnvironments(
	ctx context.Context,
	_ *pb.ListEnvironments_Request,
) (*pb.ListEnvironments_Response, error) {
	envs, err := impl.vervServices.ListEnvironments(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing environments")
	}

	out := make([]*pb.Environment, 0, len(envs))
	for _, env := range envs {
		out = append(out, environmentToPb(env))
	}

	resp := &pb.ListEnvironments_Response{
		Environments: out,
	}

	return resp, nil
}

// environmentToPb maps the domain type onto the wire type. Shared by every
// environment RPC in this package.
func environmentToPb(env domain.Environment) *pb.Environment {
	out := &pb.Environment{
		Id:         env.ID,
		Name:       env.Name,
		Suffix:     env.Suffix,
		DockerHost: env.DockerHost,
	}

	if !env.CreatedAt.IsZero() {
		out.CreatedAt = timestamppb.New(env.CreatedAt)
	}

	if !env.UpdatedAt.IsZero() {
		out.UpdatedAt = timestamppb.New(env.UpdatedAt)
	}

	return out
}
