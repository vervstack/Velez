package service_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	pb "go.vervstack.ru/Velez/pkg/velez_api"
)

func (impl *Impl) CreateDeploy(ctx context.Context, apiReq *pb.CreateDeploy_Request) (
	*pb.CreateDeploy_Response, error) {
	var err error

	switch payload := apiReq.Specification.(type) {
	case *pb.CreateDeploy_Request_New:
		return impl.handleNewDeployment(ctx, apiReq, payload)
	case *pb.CreateDeploy_Request_Upgrade_:
		return impl.handleUpgradeDeployment(ctx, apiReq, payload)
	}

	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return &pb.CreateDeploy_Response{}, nil
}

func (impl *Impl) handleNewDeployment(ctx context.Context, apiReq *pb.CreateDeploy_Request, payload *pb.CreateDeploy_Request_New) (*pb.CreateDeploy_Response, error) {
	req := domain.CreateDeployReq{
		ServiceId: apiReq.ServiceId,
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
	*pb.CreateDeploy_Response, error) {
	req := domain.UpgradeDeployReq{
		ServiceId:    apiReq.ServiceId,
		DeploymentId: payload.Upgrade.DeploymentId,
		NewImage:     payload.Upgrade.Image,
	}
	err := impl.servicesService.UpgradeDeploy(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error upgrading deployment")
	}

	return &pb.CreateDeploy_Response{}, nil
}
