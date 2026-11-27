package haproxy_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/transport/common"
)

func (impl *Impl) CreateHaproxyInstance(
	ctx context.Context,
	req *pb.CreateHaproxyInstance_Request,
) (*pb.CreateHaproxyInstance_Response, error) {
	err := impl.haproxyService.CreateInstance(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating HAProxy instance")
	}

	resp := &pb.CreateHaproxyInstance_Response{EntityId: req.GetName(), Action: jobs.CreateHaproxyInstanceAction}

	return resp, nil
}

func (impl *Impl) ValidateHaproxyTarget(
	ctx context.Context,
	req *pb.ValidateHaproxyTarget_Request,
) (*pb.ValidateHaproxyTarget_Response, error) {
	network, err := impl.haproxyService.ValidateTarget(ctx, req)

	resp := &pb.ValidateHaproxyTarget_Response{Ok: err == nil, DockerNetwork: network}
	if err != nil {
		resp.ErrorMessage = err.Error()
	}

	return resp, nil
}

func (impl *Impl) ListHaproxyInstances(
	ctx context.Context,
	req *pb.ListHaproxyInstances_Request,
) (*pb.ListHaproxyInstances_Response, error) {
	paging := common.FromPaging(req.GetPaging())

	instances, total, err := impl.haproxyService.ListInstances(ctx, paging)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing HAProxy instances")
	}

	tasks, err := impl.provisioning.List(ctx, jobs.CreateHaproxyInstanceAction, jobs.DropHaproxyInstanceAction)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing HAProxy provisioning tasks")
	}

	out := make([]*pb.HaproxyInstance, 0, len(instances))
	for _, instance := range instances {
		item := &pb.HaproxyInstance{
			Name:             instance.Name,
			MasterPgInstance: instance.MasterPgInstance,
			Port:             instance.Port,
			Status:           instance.Status,
			Environment:      instance.Environment,
			CreatedAt:        timestamppb.New(instance.CreatedAt),
		}

		out = append(out, item)
	}

	resp := &pb.ListHaproxyInstances_Response{
		Instances:    out,
		Total:        total,
		Provisioning: common.ProvisioningTasksToPb(tasks),
	}

	return resp, nil
}

func (impl *Impl) DropHaproxyInstance(
	ctx context.Context,
	req *pb.DropHaproxyInstance_Request,
) (*pb.DropHaproxyInstance_Response, error) {
	err := impl.haproxyService.DropInstance(ctx, req.GetName())
	if err != nil {
		return nil, rerrors.Wrap(err, "error dropping HAProxy instance")
	}

	resp := &pb.DropHaproxyInstance_Response{EntityId: req.GetName(), Action: jobs.DropHaproxyInstanceAction}

	return resp, nil
}

func (impl *Impl) SelectHaproxyMaster(
	ctx context.Context,
	req *pb.SelectHaproxyMaster_Request,
) (*pb.SelectHaproxyMaster_Response, error) {
	err := impl.haproxyService.SelectMaster(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error selecting HAProxy master")
	}

	resp := &pb.SelectHaproxyMaster_Response{EntityId: req.GetName(), Action: jobs.SelectHaproxyMasterAction}

	return resp, nil
}

func (impl *Impl) BootstrapVcnSidecar(
	ctx context.Context, req *pb.BootstrapVcnSidecar_Request,
) (*pb.BootstrapVcnSidecar_Response, error) {
	resp, err := impl.haproxyService.BootstrapVcnSidecar(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	return resp, nil
}

func (impl *Impl) ValidatePgBackend(
	ctx context.Context, req *pb.ValidatePgBackend_Request,
) (*pb.ValidatePgBackend_Response, error) {
	err := impl.haproxyService.ValidatePgBackend(ctx, req)

	resp := &pb.ValidatePgBackend_Response{Ok: err == nil}
	if err != nil {
		resp.ErrorMessage = err.Error()
	}

	return resp, nil
}
