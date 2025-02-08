package tests

import (
	"context"
	"net"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/godverv/Velez/internal/app"
	"github.com/godverv/Velez/pkg/velez_api"
)

func initEnv() {
	a, err := app.New()
	if err != nil {
		logrus.Fatalf("error creating app %s", err)
	}

	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)

	serv := grpc.NewServer()
	velez_api.RegisterVelezAPIServer(serv, a.Custom.ApiGrpcImpl)
	go func() {
		if err := serv.Serve(lis); err != nil {
			logrus.Fatalf("error serving grpc server for tests %s", err)
		}
	}()

	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logrus.Fatalf("error connecting to test grpc server: %s ", err)
	}

	tEnv.velezAPI = velez_api.NewVelezAPIClient(conn)
	tEnv.docker = a.Custom.NodeClients.Docker()
}
