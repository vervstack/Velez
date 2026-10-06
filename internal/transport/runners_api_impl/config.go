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
		Concurrent:          config.Concurrent,
		PullPolicy:          pullPolicyEnums(config.Settings.PullPolicy),
		AllowedPullPolicies: pullPolicyEnums(config.Settings.AllowedPullPolicies),
		CheckInterval:       config.Settings.CheckInterval,
		LogLevel:            logLevelEnum(config.Settings.LogLevel),
		ShutdownTimeout:     config.Settings.ShutdownTimeout,
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

	if req.Concurrent != nil {
		updateReq.Concurrent = rtb.NewOptional(req.GetConcurrent())
	}

	if req.GetPullPolicy() != nil {
		updateReq.PullPolicy = rtb.NewOptional(pullPolicyStrings(req.GetPullPolicy().GetValues()))
	}

	if req.GetAllowedPullPolicies() != nil {
		updateReq.AllowedPullPolicies = rtb.NewOptional(pullPolicyStrings(req.GetAllowedPullPolicies().GetValues()))
	}

	if req.CheckInterval != nil {
		updateReq.CheckInterval = rtb.NewOptional(req.GetCheckInterval())
	}

	if req.LogLevel != nil {
		updateReq.LogLevel = rtb.NewOptional(logLevelString(req.GetLogLevel()))
	}

	if req.ShutdownTimeout != nil {
		updateReq.ShutdownTimeout = rtb.NewOptional(req.GetShutdownTimeout())
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
