package runners_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/transport/common"
)

func (impl *Impl) ListRunners(
	ctx context.Context,
	req *pb.ListRunners_Request,
) (*pb.ListRunners_Response, error) {
	serviceReq := domain.ListRunnersReq{
		Paging: common.FromPaging(req.GetPaging()),
	}

	list, err := impl.runnersService.ListRunners(ctx, serviceReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing runners")
	}

	out := make([]*pb.Runner, 0, len(list.Runners))
	for _, view := range list.Runners {
		out = append(out, runnerToPb(view))
	}

	resp := &pb.ListRunners_Response{
		Runners: out,
		Total:   list.Total,
	}

	return resp, nil
}

// runnerToPb maps the domain view onto the wire type. Shared by every runner
// RPC in this package. Never populates a credential field - pb.Runner
// deliberately has none; the access token is never read back once stored.
func runnerToPb(view domain.RunnerView) *pb.Runner {
	out := &pb.Runner{
		Name:        view.Name,
		Provider:    view.Provider,
		Scope:       view.Scope,
		Target:      view.Target,
		Labels:      view.Labels,
		Environment: view.Environment,
		Status:      view.Status,
	}

	if !view.CreatedAt.IsZero() {
		out.CreatedAt = timestamppb.New(view.CreatedAt)
	}

	if !view.UpdatedAt.IsZero() {
		out.UpdatedAt = timestamppb.New(view.UpdatedAt)
	}

	return out
}
