package resource_manager

import (
	"github.com/godverv/matreshka"

	"github.com/godverv/Velez/pkg/velez_api"
)

func Postgres(resources matreshka.Resources, resourceName string) (*velez_api.CreateSmerd_Request, error) {
	pg, err := resources.Postgres(resourceName)
	if err != nil {
		return nil, err
	}

	if pg.Pwd == "" {
		pg.Pwd = genPass()
	}

	if pg.Host == "" {
		pg.Host = pg.GetName()
	}

	return &velez_api.CreateSmerd_Request{
		ImageName: "postgres:13.6",
		Hardware:  &velez_api.Container_Hardware{},
		Settings:  &velez_api.Container_Settings{},
		Env: map[string]string{
			"POSTGRES_USER":     pg.User,
			"POSTGRES_PASSWORD": pg.Pwd,
			"POSTGRES_DB":       pg.DbName,
		},
	}, nil
}
