package tests

import (
	"context"
	_ "embed"
	"os"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"

	"go.vervstack.ru/Velez/internal/app"
	"go.vervstack.ru/Velez/internal/clients"
	"go.vervstack.ru/Velez/internal/clients/docker/dockerutils"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

const (
	integrationTest = "integration_test"

	minPortToExposeTo = uint32(18501)
)

var (
	//go:embed config/test_loki.yaml
	lokiConfig []byte
)

type testEnv struct {
	velezAPI     velez_api.VelezAPIClient
	matreshkaApi matreshka_api.MatreshkaBeAPIClient
	docker       clients.Docker
	dockerAPI    client.APIClient

	app app.App
}

var testEnvironment testEnv

func TestMain(m *testing.M) {
	initVelez()

	ctx := context.Background()

	_, err := testEnvironment.velezAPI.Version(ctx, &velez_api.Version_Request{})
	if err != nil {
		logrus.Fatalf("error pinging service api %s", err)
	}

	_, err = testEnvironment.docker.Client().Ping(ctx)
	if err != nil {
		logrus.Fatalf("error pinging docker %s", err)
	}

	testEnvironment.clean()

	var code int
	defer func() {
		testEnvironment.clean()
		os.Exit(code)
	}()

	code = m.Run()
}

func (t *testEnv) createSmerd(ctx context.Context, req *velez_api.CreateSmerd_Request) (smerd *velez_api.Smerd, err error) {
	if req.Labels == nil {
		req.Labels = map[string]string{}
	}

	req.Labels[integrationTest] = "true"
	return testEnvironment.velezAPI.CreateSmerd(ctx, req)
}

func (t *testEnv) clean() {
	ctx := context.Background()

	listReq := &velez_api.ListSmerds_Request{
		Label: map[string]string{
			integrationTest: "true",
		},
	}
	cList, err := dockerutils.ListContainers(ctx, t.dockerAPI, listReq)
	if err != nil {
		logrus.Fatal(err)
	}

	for _, cont := range cList {
		err = t.dockerAPI.ContainerRemove(ctx, cont.ID,
			container.RemoveOptions{
				Force: true,
			})
		if err != nil {
			logrus.Fatal(err)
		}
	}

}

func getExpectedLabels() map[string]string {
	return map[string]string{
		labels.CreatedWithVelezLabel: "true",
		labels.MatreshkaConfigLabel:  "false",
		integrationTest:              "true",
	}
}

func getLabelsWithMatreshkaConfig() map[string]string {
	l := getExpectedLabels()
	l[labels.MatreshkaConfigLabel] = "true"
	return l
}

func getServiceName(t *testing.T) string {
	return strings.ReplaceAll(t.Name(), "/", "_")
}
