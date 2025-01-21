//go:build integration && !github_wf

package tests

import (
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/godverv/Velez/internal/app"
	"github.com/godverv/Velez/pkg/velez_api"
)

func initEnv() {
	fullApp, err := app.New()
	if err != nil {
		logrus.Fatalf("error creating app %s", err)
	}

	go func() {
		startErr := fullApp.Start()
		if startErr != nil {
			logrus.Fatal(startErr)
		}
	}()

	conn, err := grpc.NewClient("localhost:53890",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logrus.Fatalf("error connecting to test grpc server: %s ", err)
	}

	tEnv.velezAPI = velez_api.NewVelezAPIClient(conn)
	tEnv.docker = fullApp.Custom.NodeClients.Docker()
}
