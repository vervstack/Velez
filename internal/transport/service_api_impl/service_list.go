package service_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	rtb "go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/transport/common"
)

var provisioningActions = []string{
	jobs.CreateServiceAction,
	jobs.RegisterContainerAction,
	jobs.CreatePgInstanceAction,
	jobs.DropPgInstanceAction,
	jobs.CreateS3InstanceAction,
	jobs.DropS3InstanceAction,
	jobs.CreateRegistryInstanceAction,
	jobs.DropRegistryInstanceAction,
	jobs.CreateRunnerAction,
	jobs.DropRunnerAction,
	jobs.CreateDindAction,
	jobs.DropDindAction,
}

func (impl *Impl) ListServices(
	ctx context.Context,
	pbReq *velez_api.ListServices_Request,
) (*velez_api.ListServices_Response, error) {
	req := fromListServiceRequest(pbReq)

	services, err := impl.servicesService.List(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing services")
	}

	tasks, err := impl.provisioning.List(ctx, provisioningActions...)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing provisioning tasks")
	}

	resp := toListServiceResponse(services)

	resp.Provisioning = common.ProvisioningTasksToPb(tasks)

	return resp, nil
}

func fromListServiceRequest(pbReq *velez_api.ListServices_Request) domain.ListServicesReq {
	req := domain.ListServicesReq{
		Paging:          common.FromPaging(pbReq.GetPaging()),
		IncludeInternal: pbReq.GetIncludeInternal(),
	}

	if pbReq.SearchPattern != nil {
		req.NamePattern = rtb.NewOptional[string](pbReq.GetSearchPattern())
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
