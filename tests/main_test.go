//go:build integration

package tests

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/app"
	"github.com/godverv/Velez/pkg/velez_api"
)

type TestEnv struct {
	grpcApi velez_api.VelezAPIClient
}

var testEnv TestEnv

func TestMain(m *testing.M) {
	err := setupTestEnv()
	if err != nil {
		logrus.Fatal(err)
	}

	var code int
	defer func() {
		testEnv.clean()
		os.Exit(code)
	}()

	code = m.Run()
}

func setupTestEnv() error {
	app.New
	return nil
}

func (t *TestEnv) clean() {

}
