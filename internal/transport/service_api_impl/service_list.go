package service_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	rtb "go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/transport/common"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

func (impl *Impl) ListServices(ctx context.Context, pbReq *velez_api.ListServices_Request) (*velez_api.ListServices_Response, error) {
	req := fromListServiceRequest(pbReq)

	services, err := impl.servicesService.List(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return toListServiceResponse(services), nil
}

func fromListServiceRequest(pbReq *velez_api.ListServices_Request) domain.ListServicesReq {
	req := domain.ListServicesReq{
		Paging: common.FromPaging(pbReq.GetPaging()),
	}

	if pbReq.SearchPattern != nil {
		req.NamePattern = rtb.NewOptional[string](*pbReq.SearchPattern)
	}

	return req
}

func toListServiceResponse(list domain.ServiceList) *velez_api.ListServices_Response {
	out := &velez_api.ListServices_Response{
		Total:    list.Total,
		Services: toServiceBaseInfoList(list.Services),
	}

	return out
}
