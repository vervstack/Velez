package runners_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) GetRunnerCredentials(
	ctx context.Context,
	req *pb.GetRunnerCredentials_Request,
) (*pb.GetRunnerCredentials_Response, error) {
	creds, err := impl.runnersService.GetRunnerCredentials(ctx, req.GetName())
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting runner credentials")
	}

	resp := &pb.GetRunnerCredentials_Response{
		Token:           creds.Token,
		Target:          creds.Target,
		Provider:        creds.Provider,
		RegisterCommand: creds.RegisterCommand,
	}

	return resp, nil
}
