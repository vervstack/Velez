package pgaas_api_impl

import (
	"context"
	"net"
	"strconv"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/jobs"
)

func (impl *Impl) ReplicatePgDatabase(
	ctx context.Context,
	req *pb.ReplicatePgDatabase_Request,
) (*pb.ReplicatePgDatabase_Response, error) {
	err := validateReplicationRequest(req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error validating replication request")
	}

	serviceReq := domain.ReplicatePgDatabaseReq{
		SourceInstanceId: req.GetSourceInstanceId(),
		SourceHost:       req.GetSourceHost(),
		SourcePort:       req.GetSourcePort(),
		SourceUser:       req.GetSourceUser(),
		SourcePassword:   req.GetSourcePassword(),
		SourceDbName:     req.GetSourceDbName(),
		TargetInstanceId: req.GetTargetInstanceId(),
		TargetHost:       req.GetTargetHost(),
		TargetPort:       req.GetTargetPort(),
		TargetUser:       req.GetTargetUser(),
		TargetPassword:   req.GetTargetPassword(),
	}

	entityId, err := impl.postgresService.ReplicatePgDatabase(ctx, serviceReq)
	if err != nil {
		return nil, rerrors.Wrap(err, "error replicating pg database")
	}

	response := &pb.ReplicatePgDatabase_Response{
		EntityId: entityId,
		Action:   jobs.ReplicatePgDatabaseAction,
	}

	return response, nil
}

func validateReplicationRequest(req *pb.ReplicatePgDatabase_Request) error {
	isSourceInstance := req.GetSourceInstanceId() != ""
	isSourcePlain := req.GetSourceHost() != "" && req.GetSourcePort() != 0 &&
		req.GetSourceUser() != "" && req.GetSourcePassword() != "" && req.GetSourceDbName() != ""
	isTargetInstance := req.GetTargetInstanceId() != ""
	isTargetPlain := req.GetTargetHost() != "" && req.GetTargetPort() != 0 &&
		req.GetTargetUser() != "" && req.GetTargetPassword() != ""

	if isSourceInstance == isSourcePlain {
		return rerrors.New("source must specify exactly one of instance id or complete plain connection")
	}

	if isTargetInstance == isTargetPlain {
		return rerrors.New("target must specify exactly one of instance id or complete plain connection")
	}

	if isSourcePlain && isTargetPlain {
		sourcePort := strconv.FormatUint(uint64(req.GetSourcePort()), 10)
		targetPort := strconv.FormatUint(uint64(req.GetTargetPort()), 10)
		source := net.JoinHostPort(req.GetSourceHost(), sourcePort) + "/" + req.GetSourceDbName()
		target := net.JoinHostPort(req.GetTargetHost(), targetPort) + "/" + req.GetSourceDbName()

		if source == target {
			return rerrors.New("source and target database must differ")
		}
	}

	if isSourceInstance && isTargetInstance && req.GetSourceInstanceId() == req.GetTargetInstanceId() {
		return rerrors.New("source and target instance must differ")
	}

	return nil
}
