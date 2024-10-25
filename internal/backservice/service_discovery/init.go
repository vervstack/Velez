package service_discovery

import (
	"context"
	"sync"

	"github.com/Red-Sock/toolbox/keep_alive"
	errors "github.com/Red-Sock/trace-errors"
	pb "github.com/godverv/makosh/pkg/makosh_be"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/clients/makosh"
	"github.com/godverv/Velez/internal/config"
)

type ServiceDiscoveryConnection struct {
	Addr  []string
	Token string
}

var serviceDiscoveryConn = ServiceDiscoveryConnection{}

var initModeSync = sync.Once{}

func InitInstance(
	ctx context.Context,
	cfg config.Config,
	clients clients.NodeClients,
) (sd ServiceDiscoveryConnection) {
	initModeSync.Do(func() {
		serviceDiscoveryConn = startServiceDiscoveryInstance(ctx, cfg, clients)
	})

	return serviceDiscoveryConn
}

func startServiceDiscoveryInstance(ctx context.Context, cfg config.Config, clients clients.NodeClients,
) ServiceDiscoveryConnection {
	makoshBackgroundTask, err := newKeepAliveTask(cfg, clients)
	if err != nil {
		logrus.Fatalf("error creating service discovery background task: %s", err)
	}

	logrus.Info("Starting service discovery background task")
	ak := keep_alive.KeepAlive(makoshBackgroundTask, keep_alive.WithCancel(ctx.Done()))
	ak.Wait()

	opts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(makosh.HeaderInterceptor(makoshBackgroundTask.AuthToken)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	dial, err := grpc.NewClient(makoshBackgroundTask.Address, opts...)
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "error dialing"))
	}

	makoshClient := pb.NewMakoshBeAPIClient(dial)

	req := &pb.UpsertEndpoints_Request{
		Endpoints: []*pb.Endpoint{
			{
				ServiceName: makosh.ServiceName,
				Addrs:       []string{makoshBackgroundTask.Address},
			},
		},
	}
	_, err = makoshClient.UpsertEndpoints(ctx, req)
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "error upserting makosh endpoint"))
	}

	return ServiceDiscoveryConnection{
		Addr:  []string{makoshBackgroundTask.Address},
		Token: makoshBackgroundTask.AuthToken,
	}
}
