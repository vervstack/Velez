//go:build github_wf

package tests

import (
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/godverv/Velez/internal/clients/docker"
	"github.com/godverv/Velez/pkg/velez_api"
)

func initEnv() {

	conn, err := grpc.NewClient("[::]:53890",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logrus.Fatalf("error connecting to test grpc server: %s ", err)
	}

	tEnv.velezAPI = velez_api.NewVelezAPIClient(conn)
	tEnv.docker, err = docker.NewClient()
	if err != nil {
		logrus.Fatalf("error creating docker client: %s ", err)
	}
}
