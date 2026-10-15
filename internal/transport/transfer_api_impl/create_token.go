package transfer_api_impl

import (
	"context"

	"go.redsock.ru/rerrors"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
)

func (impl *Impl) CreateTransferToken(
	ctx context.Context,
	_ *pb.CreateTransferToken_Request,
) (*pb.CreateTransferToken_Response, error) {
	token, expiresAt, err := impl.transfersIn.CreateToken(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating transfer token")
	}

	resp := &pb.CreateTransferToken_Response{
		Token:     token,
		ExpiresAt: timestamppb.New(expiresAt),
	}

	return resp, nil
}
