package pgaas_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/jobs"
)

func (impl *Impl) EnablePgReplication(
	ctx context.Context,
	req *pb.EnablePgReplication_Request,
) (*pb.EnablePgReplication_Response, error) {
	name := req.GetName()

	err := impl.postgresService.EnablePgReplication(ctx, name)
	if err != nil {
		return nil, rerrors.Wrap(err, "error enabling pg replication")
	}

	response := &pb.EnablePgReplication_Response{
		EntityId: name,
		Action:   jobs.EnablePgReplicationAction,
	}

	return response, nil
}
