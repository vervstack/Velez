package service_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

func (impl *Impl) SetServiceProxy(
	ctx context.Context, apiReq *pb.SetServiceProxy_Request,
) (*pb.SetServiceProxy_Response, error) {
	_, err := impl.resolveEnvironment(ctx, apiReq.GetEnvironment())
	if err != nil {
		return nil, err
	}

	req := domain.SetServiceProxyReq{
		ServiceName:      apiReq.GetServiceName(),
		ProxyUrl:         apiReq.GetProxyUrl(),
		ProxyBypassHosts: apiReq.GetProxyBypassHosts(),
	}

	err = impl.servicesService.SetServiceProxy(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error setting service proxy")
	}

	return &pb.SetServiceProxy_Response{}, nil
}
