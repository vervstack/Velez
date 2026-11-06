package velez_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/cluster/env"
)

func (impl *Impl) SelfRedeploy(
	ctx context.Context,
	req *velez_api.SelfRedeploy_Request,
) (*velez_api.SelfRedeploy_Response, error) {
	if !env.IsInContainer() {
		return nil, rerrors.New("self redeploy is only available when velez runs inside a container")
	}

	containerId := env.GetContainerId()
	if containerId == nil || *containerId == "" {
		return nil, rerrors.New("self redeploy: container id of this velez instance is unknown")
	}

	err := impl.smerdService.SelfRedeploy(ctx, req.GetImage(), *containerId)
	if err != nil {
		return nil, rerrors.Wrap(err, "error self redeploying")
	}

	resp := &velez_api.SelfRedeploy_Response{Accepted: true}

	return resp, nil
}
