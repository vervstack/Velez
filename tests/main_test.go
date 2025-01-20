//go:build integration

package tests

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/godverv/Velez/internal/app"
	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
	"github.com/godverv/Velez/internal/pipelines/deploy_steps"
	"github.com/godverv/Velez/pkg/velez_api"
)

const (
	integrationTest = "integration_test"

	minPortToExposeTo = uint32(18501)
)

type testEnv struct {
	app      app.App
	velezAPI velez_api.VelezAPIClient
}

var tEnv testEnv

func TestMain(m *testing.M) {
	initEnv()

	var code int
	defer func() {
		tEnv.clean()
		os.Exit(code)
	}()

	code = m.Run()
}

func initEnv() {
	var err error

	tEnv.app, err = app.New()
	if err != nil {
		logrus.Fatal(err)
	}

	tEnv.clean()
	go func() {
		startErr := tEnv.app.Start()
		if startErr != nil {
			logrus.Fatal(startErr)
		}
	}()

	const bufSize = 1024 * 1024

	lis := bufconn.Listen(bufSize)

	serv := grpc.NewServer()

	velez_api.RegisterVelezAPIServer(serv, tEnv.app.Custom.ApiGrpcImpl)
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
}

func (t *testEnv) callCreate(ctx context.Context, req *velez_api.CreateSmerd_Request) (smerd *velez_api.Smerd, err error) {
	if req.Labels == nil {
		req.Labels = map[string]string{}
	}

	req.Labels[integrationTest] = "true"
	return tEnv.velezAPI.CreateSmerd(ctx, req)
}

func (t *testEnv) clean() {
	ctx := context.Background()

	listReq := &velez_api.ListSmerds_Request{
		Label: map[string]string{
			integrationTest: "true",
		},
	}
	cList, err := dockerutils.ListContainers(ctx, t.app.Custom.NodeClients.Docker(), listReq)
	if err != nil {
		logrus.Fatal(err)
	}

	for _, cont := range cList {
		err = t.app.Custom.NodeClients.Docker().ContainerRemove(ctx, cont.ID,
			container.RemoveOptions{
				Force: true,
			})
		if err != nil {
			logrus.Fatal(err)
		}
	}

}

func (t *testEnv) getExpectedLabels() map[string]string {
	return map[string]string{
		deploy_steps.CreatedWithVelezLabel: "true",
		deploy_steps.MatreshkaConfigLabel:  "false",
		integrationTest:                    "true",
	}
}
