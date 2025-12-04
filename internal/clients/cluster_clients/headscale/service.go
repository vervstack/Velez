package headscale

import (
	"context"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
)

type Client struct {
	docker        node_clients.Docker
	containerName string
	apiKey        string

	headscaleApiUrl string
}

func New(ctx context.Context, nc node_clients.NodeClients, containerName string) (srv *Client, err error) {
	srv = &Client{
		docker:        nc.Docker(),
		containerName: containerName,
		// TODO change onto actual
		headscaleApiUrl: "http://localhost:8080",

		// Remove
		apiKey: "_ARnAyX.9BHTGPRGiAPeKhzYLJxNpQABtQUy23Qv",
	}
	//TODO return to live
	//srv.apiKey, err = issueNewApiKey(ctx, docker, containerName)
	//if err != nil {
	//	return nil, rerrors.Wrap(err, "error issuing api key")
	//}

	return srv, nil
}
