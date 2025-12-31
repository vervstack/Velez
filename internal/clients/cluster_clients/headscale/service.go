package headscale

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
)

type Client struct {
	docker node_clients.Docker

	// deprecated
	containerName string
	apiKey        string

	headscaleApiUrl string
}

func Connect(ctx context.Context, url, token string) (*Client, error) {
	srv := &Client{
		headscaleApiUrl: url,
		apiKey:          token,
	}

	namespaces, err := srv.ListNamespaces(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing namespaces")
	}

	_ = namespaces

	return srv, nil
}

func New(ctx context.Context, nc node_clients.NodeClients, containerName string) (srv *Client, err error) {
	srv = &Client{
		docker:        nc.Docker(),
		containerName: containerName,
		// TODO change onto actual
		//headscaleApiUrl: "http://localhost:8080",
		headscaleApiUrl: "https://vcn.redsock.ru",
		// Remove
		//apiKey: "_ARnAyX.9BHTGPRGiAPeKhzYLJxNpQABtQUy23Qv",
		apiKey: "v6h7IEa.vqf_q8lZTmBgphgMqfYP8Q6qKRrtW4K4",
	}
	//TODO return to live
	//srv.apiKey, err = issueNewApiKey(ctx, docker, containerName)
	//if err != nil {
	//	return nil, rerrors.Wrap(err, "error issuing api key")
	//}

	return srv, nil
}
