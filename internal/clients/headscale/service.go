package headscale

import (
	"context"

	"go.vervstack.ru/Velez/internal/clients"
)

type Client struct {
	docker        clients.Docker
	containerName string
	apiKey        string

	headscaleApiUrl string
}

func New(ctx context.Context, nc clients.NodeClients, containerName string) (srv *Client, err error) {
	srv = &Client{
		docker:        nc.Docker(),
		containerName: containerName,
		// TODO change onto actual
		headscaleApiUrl: "http://localhost:8080",

		// Remove
		apiKey: "TCXqxOn.Apx28pH4_ELTJ1Ts_xunI4r67nmjSHyU",
	}
	//TODO return to live
	//srv.apiKey, err = issueNewApiKey(ctx, docker, containerName)
	//if err != nil {
	//	return nil, rerrors.Wrap(err, "error issuing api key")
	//}

	return srv, nil
}
