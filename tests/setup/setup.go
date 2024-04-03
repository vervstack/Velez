package setup

import (
	"context"

	"github.com/godverv/matreshka-be/pkg/matreshka_api"

	"github.com/godverv/Velez/internal/clients/grpc"
	"github.com/godverv/Velez/internal/config"
)

type App struct {
	Config config.Config

	Matreshka matreshka_api.MatreshkaBeAPIClient
}

func GetApp() (a App, err error) {
	a.Config, err = config.LoadWithPath("./../config/test.yaml")
	if err != nil {
		return a, err
	}
	ctx := context.Background()

	a.Matreshka, err = grpc.NewMatreshkaBeAPIClient(ctx, a.Config)
	if err != nil {
		return a, err
	}

	return a, nil
}
