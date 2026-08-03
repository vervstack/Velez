package service_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func (impl *Impl) CreateDeploy(ctx context.Context, apiReq *pb.CreateDeploy_Request) (
	*pb.CreateDeploy_Response, error,
) {
	// Environment is required and must resolve to a real environment. The
	// suffix itself is re-resolved later by internal/workers/deploy_watcher.go
	// when the scheduled deployment is actually launched.
	_, err := impl.resolveEnvironment(ctx, apiReq.GetEnvironment())
	if err != nil {
		return nil, err
	}

	switch payload := apiReq.GetSpecification().(type) {
	case *pb.CreateDeploy_Request_New:
		return impl.handleNewDeployment(ctx, apiReq, payload)
	case *pb.CreateDeploy_Request_Upgrade_:
		return impl.handleUpgradeDeployment(ctx, apiReq, payload)
	}

	return &pb.CreateDeploy_Response{}, nil
}

func (impl *Impl) handleNewDeployment(
	ctx context.Context,
	apiReq *pb.CreateDeploy_Request,
	payload *pb.CreateDeploy_Request_New,
) (*pb.CreateDeploy_Response, error) {
	// Carry the deploy's environment into the smerd spec that gets persisted:
	// deploy_watcher.go only ever sees the stored CreateSmerd_Request, so the
	// environment has to live there for the suffix to be resolvable later.
	if payload.New != nil && payload.New.GetEnvironment() == "" {
		payload.New.Environment = apiReq.GetEnvironment()
	}

	req := domain.CreateDeployReq{
		ServiceName: apiReq.GetServiceName(),
		LaunchSmerd: domain.LaunchSmerd{
			CreateSmerd_Request: payload.New,
		},
	}

	err := impl.servicesService.CreateNewDeploy(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating new deploy")
	}

	return &pb.CreateDeploy_Response{}, nil
}

func (impl *Impl) handleUpgradeDeployment(ctx context.Context,
	apiReq *pb.CreateDeploy_Request, payload *pb.CreateDeploy_Request_Upgrade_) (
	*pb.CreateDeploy_Response, error,
) {
	req := domain.UpgradeDeployReq{
		ServiceName:  apiReq.GetServiceName(),
		DeploymentId: payload.Upgrade.GetDeploymentId(),
		NewImage:     payload.Upgrade.Image,
	}

	err := impl.servicesService.UpgradeDeploy(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error upgrading deployment")
	}

	return &pb.CreateDeploy_Response{}, nil
}
