package resource_manager

import (
	"path"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/matreshka"

	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/pkg/velez_api"
)

func Sqlite(resources matreshka.DataSources, resourceName string) (deps domain.Dependencies, err error) {
	res, err := resources.Sqlite(resourceName)
	if err != nil {
		return deps, errors.Wrap(err, "error extracting resource cfg")
	}

	volDep := domain.VolumeDependency{
		Constructor: &velez_api.VolumeBindings{
			Volume:        resourceName,
			ContainerPath: path.Dir(res.Path),
		},
	}

	deps.Volumes = append(deps.Volumes, volDep)

	return deps, nil
}
