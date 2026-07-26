package service_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) GetServiceGraph(
	ctx context.Context,
	pbReq *pb.GetServiceGraph_Request,
) (*pb.GetServiceGraph_Response, error) {
	graph, err := impl.servicesService.GetServiceGraph(ctx, pbReq.GetServiceName())
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	pbDependencies := make([]*pb.ServiceDependencyInfo, 0, len(graph.Dependencies))
	for _, dep := range graph.Dependencies {
		pbDep := &pb.ServiceDependencyInfo{
			ServiceName: dep.TargetService,
			Proto:       dep.Proto,
			RequestRate: dep.RequestRate,
			NodeType:    pb.NodeType(dep.NodeType),
		}

		pbDependencies = append(pbDependencies, pbDep)
	}

	pbCallers := make([]*pb.ServiceDependencyInfo, 0, len(graph.Callers))
	for _, caller := range graph.Callers {
		pbCaller := &pb.ServiceDependencyInfo{
			ServiceName: caller.SourceService,
			Proto:       caller.Proto,
			RequestRate: caller.RequestRate,
			NodeType:    pb.NodeType_NODE_TYPE_SERVICE,
		}

		pbCallers = append(pbCallers, pbCaller)
	}

	resp := &pb.GetServiceGraph_Response{
		Dependencies: pbDependencies,
		Callers:      pbCallers,
	}

	return resp, nil
}
