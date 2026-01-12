package headscale

import (
	"context"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/cluster/env"
)

type Client struct {
	apiKey          string
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

func ConnectToContainer(ctx context.Context, nc node_clients.NodeClients, containerName string) (*Client, error) {
	var srv Client
	var err error

	srv.headscaleApiUrl, err = getApiAddress(ctx, nc.Docker(), containerName)
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting api address")
	}

	srv.apiKey, err = issueNewApiKey(ctx, nc.Docker(), containerName)
	if err != nil {
		return nil, rerrors.Wrap(err, "error issuing api key")
	}

	stateManager := nc.LocalStateManager()
	localState := stateManager.GetForUpdate()
	defer stateManager.SetAndRelease(localState)

	localState.HeadscaleServerUrl = srv.headscaleApiUrl
	localState.HeadscaleKey = srv.apiKey

	return &srv, nil
}

func getApiAddress(ctx context.Context, dockerClient node_clients.Docker, containerName string) (address string, err error) {
	dockerApi := dockerClient.Client()

	cont, err := dockerApi.ContainerInspect(ctx, containerName)
	if err != nil {
		return "", rerrors.Wrap(err, "error inspecting headscale container")
	}

	var contPort, hostPort string

	for port, pb := range cont.HostConfig.PortBindings {
		contPort = port.Port()
		for _, p := range pb {
			hostPort = p.HostPort
		}
	}

	if !env.IsInContainer() {
		return "http://localhost:" + hostPort, nil
	}
	vervNet, isExists := cont.NetworkSettings.Networks[env.VervNetwork]
	if !isExists {
		return "", rerrors.New("headscale container isn't connected to vervnet")
	}

	if len(vervNet.Aliases) == 0 {
		return "", rerrors.New("headscale container doesn't have any aliases")
	}

	return "http://" + vervNet.Aliases[0] + ":" + contPort, nil
}
