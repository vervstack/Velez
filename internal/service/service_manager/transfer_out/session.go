package transfer_out

import (
	"context"

	"go.redsock.ru/rerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/utils/common"
)

// session is one open stream to the target; it owns the connection and the stream's context.
type session struct {
	conn   *grpc.ClientConn
	stream velez_api.TransferAPI_TransferClient
	cancel context.CancelFunc
}

func (s *session) close() {
	s.cancel()
	common.CloseWithLog(s.conn.Close, "transfer connection")
}

// openSession dials the target, sends Begin and waits for the target's verdict, so a bad
// token or a name collision reaches the caller before anything is stopped.
// TLS is terminated by the operator's reverse proxy in front of the target.
func openSession(
	ctx context.Context,
	cancel context.CancelFunc,
	address string,
	begin *velez_api.Transfer_Begin,
) (*session, error) {
	creds := insecure.NewCredentials()

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, rerrors.Wrap(err, "error dialing target")
	}

	client := velez_api.NewTransferAPIClient(conn)

	stream, err := client.Transfer(ctx)
	if err != nil {
		common.CloseWithLog(conn.Close, "transfer connection")

		return nil, rerrors.Wrap(err, "error opening transfer stream")
	}

	opened := &session{conn: conn, stream: stream, cancel: cancel}

	err = handshake(stream, begin)
	if err != nil {
		opened.close()

		return nil, rerrors.Wrap(err)
	}

	return opened, nil
}

func handshake(stream velez_api.TransferAPI_TransferClient, begin *velez_api.Transfer_Begin) error {
	err := stream.Send(beginRequest(begin))
	if err != nil {
		return peerError(err, "error sending begin")
	}

	resp, err := stream.Recv()
	if err != nil {
		return peerError(err, "error receiving acceptance")
	}

	if resp.GetAccepted() == nil {
		return rerrors.Wrap(errTargetNotAccepted)
	}

	return nil
}
