package pgaas_api_impl

import (
	"context"
	"net"
	"strings"

	"go.redsock.ru/rerrors"
	"google.golang.org/grpc/metadata"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
)

const (
	forwardedHostHeader = "x-forwarded-host"
	authorityHeader     = ":authority"
)

func (impl *Impl) GetPgInstanceCredentials(
	ctx context.Context,
	req *pb.GetPgInstanceCredentials_Request,
) (*pb.GetPgInstanceCredentials_Response, error) {
	requestHost := requestHostname(ctx)

	creds, err := impl.postgresService.GetPgInstanceCredentials(ctx, req.GetName(), requestHost)
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting pg instance credentials")
	}

	resp := &pb.GetPgInstanceCredentials_Response{
		DbName:    creds.DbName,
		Username:  creds.Username,
		Password:  creds.Password,
		Dsn:       creds.Dsn,
		InNetwork: pgConnectionToPb(creds.InNetwork),
	}

	if creds.External != nil {
		resp.External = pgConnectionToPb(*creds.External)
	}

	return resp, nil
}

func pgConnectionToPb(conn domain.PgConnectionView) *pb.PgConnection {
	return &pb.PgConnection{
		Dsn:       conn.Dsn,
		MaskedDsn: conn.MaskedDsn,
	}
}

// requestHostname is the bare hostname the client reached Velez on.
// grpc-gateway forwards the HTTP Host as x-forwarded-host; a direct gRPC
// call carries it as :authority.
func requestHostname(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	for _, header := range []string{forwardedHostHeader, authorityHeader} {
		values := md.Get(header)
		if len(values) == 0 {
			continue
		}

		return stripPort(strings.TrimSpace(values[0]))
	}

	return ""
}

func stripPort(hostPort string) string {
	host, _, err := net.SplitHostPort(hostPort)
	if err != nil {
		return strings.Trim(hostPort, "[]")
	}

	return host
}
