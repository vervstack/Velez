//go:build integration

package tests

import (
	"context"
	"os"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/app"
	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
	"github.com/godverv/Velez/internal/service/service_manager/smerd_launcher"
	"github.com/godverv/Velez/pkg/velez_api"
)

const (
	integrationTest = "integration_test"
)

type TestEnv struct {
	*app.App
}

var testEnv TestEnv

func TestMain(m *testing.M) {
	testEnv.App = app.New()
	testEnv.clean()
	go testEnv.App.Start()

	var code int
	defer func() {
		testEnv.clean()
		os.Exit(code)
	}()

	code = m.Run()
}

func (t *TestEnv) callCreate(ctx context.Context, req *velez_api.CreateSmerd_Request) (smerd *velez_api.Smerd, err error) {
	if req.Labels == nil {
		req.Labels = map[string]string{}
	}

	req.Labels[integrationTest] = "true"
	return testEnv.GrpcApi.CreateSmerd(ctx, req)
}

func (t *TestEnv) clean() {
	ctx := context.Background()

	listReq := &velez_api.ListSmerds_Request{
		Label: map[string]string{
			integrationTest: "true",
		},
	}
	cList, err := dockerutils.ListContainers(ctx, t.InternalClients.Docker(), listReq)
	if err != nil {
		logrus.Fatal(err)
	}

	for _, cont := range cList {
		err = t.InternalClients.Docker().ContainerRemove(ctx, cont.ID,
			container.RemoveOptions{
				Force: true,
			})
		if err != nil {
			logrus.Fatal(err)
		}
	}

}

func (t *TestEnv) getExpectedLabels() map[string]string {
	return map[string]string{
		smerd_launcher.CreatedWithVelezLabel: "true",
		smerd_launcher.MatreshkaConfigLabel:  "false",
		integrationTest:                      "true",
	}
}
