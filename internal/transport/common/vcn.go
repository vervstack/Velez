package common

import (
	"context"
	"net"
	"strconv"
	"strings"

	"google.golang.org/grpc/metadata"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service"
)

func CreateVcnNetworkRequest(ctx context.Context, req *pb.CreateVcnNetwork_Request) service.CreateVcnNetworkRequest {
	input := service.CreateVcnNetworkRequest{
		Name: req.GetName(),
		Host: vcnRequestHost(ctx),
	}

	if deploy := req.GetDeploy(); deploy != nil {
		input.Deploy = &domain.SetupHeadscaleRequest{CustomImage: deploy.CustomImage}
		if deploy.CustomPort != nil {
			port := strconv.FormatUint(uint64(deploy.GetCustomPort()), 10)

			input.Deploy.ExposeToPort = &port
		}
	}

	if external := req.GetExternal(); external != nil {
		input.External = &service.VcnExternalSource{
			Url:   external.GetUrl(),
			Token: external.GetToken(),
		}
	}

	return input
}

func vcnRequestHost(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	for _, header := range []string{"x-forwarded-host", ":authority"} {
		values := md.Get(header)
		if len(values) == 0 {
			continue
		}

		authority := strings.TrimSpace(values[0])

		host, _, err := net.SplitHostPort(authority)
		if err != nil {
			host = strings.Trim(authority, "[]")
		}

		return host
	}

	return ""
}
