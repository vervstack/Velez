package resource_manager

import (
	"github.com/godverv/matreshka"

	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/pkg/velez_api"
)

func Postgres(resources matreshka.DataSources, resourceName string) (deps domain.Dependencies, err error) {
	pg, err := resources.Postgres(resourceName)
	if err != nil {
		return domain.Dependencies{}, err
	}

	if pg.Pwd == "" {
		pg.Pwd = genPass()
	}

	if pg.Host == "" {
		pg.Host = pg.GetName()
	}

	smerdDep := domain.SmerdDependency{
		Constructor: &velez_api.CreateSmerd_Request{
			ImageName: "postgres:13.6",
			Hardware:  &velez_api.Container_Hardware{},
			Settings:  &velez_api.Container_Settings{},
			Env: map[string]string{
				"POSTGRES_USER":     pg.User,
				"POSTGRES_PASSWORD": pg.Pwd,
				"POSTGRES_DB":       pg.DbName,
			},
			Healthcheck: &velez_api.Container_Healthcheck{
				Command:        "pg_isready -U postgres",
				IntervalSecond: 5,
				TimeoutSecond:  5,
				Retries:        3,
			},
		}}

	deps.Smerds = append(deps.Smerds, smerdDep)
	return deps, nil
}
