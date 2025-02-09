package tests

import (
	"context"
	"os"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/clients/docker/dockerutils"
	"github.com/godverv/Velez/internal/domain/labels"
	"github.com/godverv/Velez/pkg/velez_api"
)

const (
	integrationTest = "integration_test"

	minPortToExposeTo = uint32(18501)
)

type testEnv struct {
	velezAPI velez_api.VelezAPIClient
	docker   clients.Docker
}

var tEnv testEnv

func TestMain(m *testing.M) {
	initEnv()

	ctx := context.Background()

	_, err := tEnv.velezAPI.Version(ctx, &velez_api.Version_Request{})
	if err != nil {
		logrus.Fatalf("error pinging service api %s", err)
	}

	_, err = tEnv.docker.Ping(ctx)
	if err != nil {
		logrus.Fatalf("error pinging docker %s", err)
	}

	tEnv.clean()

	var code int
	defer func() {
		tEnv.clean()
		os.Exit(code)
	}()

	code = m.Run()
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
	cList, err := dockerutils.ListContainers(ctx, t.docker, listReq)
	if err != nil {
		logrus.Fatal(err)
	}

	for _, cont := range cList {
		err = t.docker.ContainerRemove(ctx, cont.ID,
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
		labels.CreatedWithVelezLabel: "true",
		labels.MatreshkaConfigLabel:  "false",
		integrationTest:              "true",
	}
}
