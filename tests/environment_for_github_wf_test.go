//go:build github_wf

package tests

import (
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.vervstack.ru/Velez/internal/clients/docker"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

func initEnv() {

	conn, err := grpc.NewClient("[::]:53890",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logrus.Fatalf("error connecting to test grpc server: %s ", err)
	}

	testEnvironment.velezAPI = velez_api.NewVelezAPIClient(conn)
	testEnvironment.docker, err = docker.NewClient()
	if err != nil {
		logrus.Fatalf("error creating docker client: %s ", err)
	}

	testEnvironment.dockerAPI = testEnvironment.docker.Client()
}
