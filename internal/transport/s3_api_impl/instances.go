package s3_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
	"go.vervstack.ru/Velez/internal/transport/common"
)

func (impl *Impl) ListS3Instances(
	ctx context.Context,
	req *pb.ListS3Instances_Request,
) (*pb.ListS3Instances_Response, error) {
	paging := common.FromPaging(req.GetPaging())

	instances, total, err := impl.s3Service.ListInstances(ctx, paging)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing s3 instances")
	}

	out := make([]*pb.S3Instance, 0, len(instances))
	for _, instance := range instances {
		out = append(out, s3InstanceToPb(instance))
	}

	tasks, err := impl.provisioning.List(ctx, jobs.CreateS3InstanceAction)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing provisioning tasks")
	}

	resp := &pb.ListS3Instances_Response{
		Instances: out,
		Total:     total,

		Provisioning: common.ProvisioningTasksToPb(tasks),
	}

	return resp, nil
}

func (impl *Impl) DropS3Instance(
	ctx context.Context,
	req *pb.DropS3Instance_Request,
) (*pb.DropS3Instance_Response, error) {
	err := impl.s3Service.DropInstance(ctx, req.GetName())
	if err != nil {
		return nil, rerrors.Wrap(err, "error dropping s3 instance")
	}

	return &pb.DropS3Instance_Response{}, nil
}

func (impl *Impl) GetS3InstanceCredentials(
	ctx context.Context,
	req *pb.GetS3InstanceCredentials_Request,
) (*pb.GetS3InstanceCredentials_Response, error) {
	credentials, err := impl.s3Service.GetInstanceCredentials(ctx, req.GetName())
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting s3 instance credentials")
	}

	resp := &pb.GetS3InstanceCredentials_Response{
		AdminToken:         credentials.AdminToken,
		S3Endpoint:         credentials.S3Endpoint,
		InternalS3Endpoint: credentials.InternalS3Endpoint,
		Region:             credentials.Region,
	}

	if credentials.WebUiUrl != "" {
		resp.WebUiUrl = &credentials.WebUiUrl
		resp.WebUiUsername = &credentials.WebUiUsername
		resp.WebUiPassword = &credentials.WebUiPassword
	}

	return resp, nil
}
