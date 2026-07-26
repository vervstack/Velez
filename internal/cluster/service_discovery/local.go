package service_discovery

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/makosh/pkg/makosh_be"
	"google.golang.org/grpc"
)

type dockerServiceDiscovery struct {
	docker node_clients.Docker
}

func newDockerServiceDiscovery(d node_clients.Docker) cluster_clients.ServiceDiscovery {
	return &dockerServiceDiscovery{docker: d}
}

func (d *dockerServiceDiscovery) UpsertEndpoints(_ context.Context, _ *makosh_be.UpsertEndpoints_Request, _ ...grpc.CallOption) (*makosh_be.UpsertEndpoints_Response, error) {
	return &makosh_be.UpsertEndpoints_Response{}, nil
}

func (d *dockerServiceDiscovery) ListEndpoints(ctx context.Context, req *makosh_be.ListEndpoints_Request, _ ...grpc.CallOption) (*makosh_be.ListEndpoints_Response, error) {
	opts := container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", labels.ComposeGroupLabel+"="+req.GetServiceName()),
		),
	}

	containers, err := d.docker.Client().ContainerList(ctx, opts)
	if err != nil {
		return nil, err
	}

	var urls []string

	for _, c := range containers {
		for _, p := range c.Ports {
			if p.PublicPort != 0 {
				ip := p.IP
				if ip == "" || ip == "0.0.0.0" {
					ip = "127.0.0.1"
				}

				urls = append(urls, fmt.Sprintf("%s:%d", ip, p.PublicPort))
			}
		}
	}

	resp := &makosh_be.ListEndpoints_Response{
		Urls: urls,
	}

	return resp, nil
}

func (d *dockerServiceDiscovery) Version(_ context.Context, _ *makosh_be.Version_Request, _ ...grpc.CallOption) (*makosh_be.Version_Response, error) {
	return &makosh_be.Version_Response{}, nil
}
