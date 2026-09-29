package runners_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	rtb "go.redsock.ru/toolbox"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func (impl *Impl) GetRunnerConfig(
	ctx context.Context,
	req *pb.GetRunnerConfig_Request,
) (*pb.GetRunnerConfig_Response, error) {
	config, err := impl.runnersService.GetRunnerConfig(ctx, req.GetName())
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting runner config")
	}

	resp := &pb.GetRunnerConfig_Response{
		BaseUrl:             config.BaseUrl,
		DockerImage:         config.DockerImage,
		DockerSocketAddress: config.DockerSocketAddress,
	}

	return resp, nil
}

func (impl *Impl) UpdateRunnerConfig(
	ctx context.Context,
	req *pb.UpdateRunnerConfig_Request,
) (*pb.UpdateRunnerConfig_Response, error) {
	updateReq := domain.UpdateRunnerConfigReq{
		Name: req.GetName(),
	}

	if req.BaseUrl != nil {
		updateReq.BaseUrl = rtb.NewOptional(req.GetBaseUrl())
	}

	if req.DockerImage != nil {
		updateReq.DockerImage = rtb.NewOptional(req.GetDockerImage())
	}

	if req.DockerSocketAddress != nil {
		updateReq.DockerSocketAddress = rtb.NewOptional(req.GetDockerSocketAddress())
	}

	result, err := impl.runnersService.UpdateRunnerConfig(ctx, updateReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error updating runner config")
	}

	resp := &pb.UpdateRunnerConfig_Response{
		RequiresReregister: result.RequiresReregister,
		RequiresRedeploy:   result.RequiresRedeploy,
	}

	return resp, nil
}
