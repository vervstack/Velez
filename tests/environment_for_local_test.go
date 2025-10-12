//go:build !github_wf

package tests

import (
	"context"
	"net"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"go.vervstack.ru/Velez/internal/app"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

func initVelez() {
	var err error
	testEnvironment.app, err = app.New()
	if err != nil {
		logrus.Fatalf("error creating app %s", err)
	}

	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)

	serv := grpc.NewServer()
	velez_api.RegisterVelezAPIServer(serv, testEnvironment.app.Custom.ApiGrpcImpl)
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

	testEnvironment.velezAPI = velez_api.NewVelezAPIClient(conn)
	testEnvironment.docker = testEnvironment.app.Custom.NodeClients.Docker()
	testEnvironment.dockerAPI = testEnvironment.docker.Client()
}
