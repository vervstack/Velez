package pgaas_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) GetPgInstanceCredentials(
	ctx context.Context,
	req *pb.GetPgInstanceCredentials_Request,
) (*pb.GetPgInstanceCredentials_Response, error) {
	creds, err := impl.postgresService.GetPgInstanceCredentials(ctx, req.GetName())
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting pg instance credentials")
	}

	resp := &pb.GetPgInstanceCredentials_Response{
		DbName:   creds.DbName,
		Username: creds.Username,
		Password: creds.Password,
		Dsn:      creds.Dsn,
	}

	return resp, nil
}
