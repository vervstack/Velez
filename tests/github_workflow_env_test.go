//go:build github_wf

package tests

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/godverv/Velez/internal/clients/docker"
	"github.com/godverv/Velez/pkg/velez_api"
)

func initEnv() {
	var err error
	tEnv.docker, err = docker.NewClient()
	if err != nil {
		logrus.Fatalf("error initializing docker client %s", err)
	}

	conn, err := grpc.NewClient("localhost:53890",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logrus.Fatalf("error connecting to test grpc server: %s ", err)
	}

	tEnv.velezAPI = velez_api.NewVelezAPIClient(conn)

	_, err = tEnv.velezAPI.Version(context.Background(), &velez_api.Version_Request{})
	if err != nil {
		logrus.Fatalf("error pinging api %s", err)
	}
}
