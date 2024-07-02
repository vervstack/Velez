package resource_manager

import (
	"context"
	"strings"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/docker/docker/client"
	"github.com/godverv/matreshka"
	"github.com/godverv/matreshka/resources"

	"github.com/godverv/Velez/internal/domain"
	"github.com/godverv/Velez/pkg/velez_api"
)

type SmerdConstructor func(resources matreshka.DataSources, resourceName string) (domain.Dependencies, error)

type ResourceManager struct {
	docker client.CommonAPIClient

	images map[string]SmerdConstructor
}

func New(docker client.CommonAPIClient) *ResourceManager {
	return &ResourceManager{
		docker: docker,
		images: map[string]SmerdConstructor{
			resources.PostgresResourceName: Postgres,
			resources.SqliteResourceName:   Sqlite,
		},
	}
}

func (m *ResourceManager) GetDependencies(serviceName string, cfg matreshka.AppConfig) (domain.Dependencies, error) {
	deps := domain.Dependencies{}
	for _, cfgResource := range cfg.DataSources {
		constructor := m.getByType(cfgResource.GetType())
		if constructor == nil {
			continue
		}

		resourceDeps, err := constructor(cfg.DataSources, cfgResource.GetName())
		if err != nil {
			return deps, errors.Wrap(err, "error getting resource-smerd config ")
		}

		for _, smerd := range resourceDeps.Smerds {
			smerd.Constructor.Name = serviceName + "_" + smerd.Constructor.Name

			vervNetworkBind := &velez_api.NetworkBind{
				NetworkName: serviceName,
				Aliases:     []string{cfgResource.GetName()},
			}

			smerd.Constructor.Settings.Networks = append(
				smerd.Constructor.Settings.Networks,
				vervNetworkBind,
			)

			deps.Smerds = append(deps.Smerds, smerd)
		}

		for _, vol := range resourceDeps.Volumes {
			vol.Constructor.Volume = serviceName + "_" + vol.Constructor.Volume

			deps.Volumes = append(deps.Volumes, vol)
		}
	}

	return deps, nil
}

func (m *ResourceManager) getByType(resourceName string) SmerdConstructor {
	return m.images[resourceName]
}

func (m *ResourceManager) FindDependenciesOnThisNode(ctx context.Context, deps domain.Dependencies) (domain.Dependencies, error) {
	for idx := range deps.Smerds {
		cont, err := m.docker.ContainerInspect(ctx, deps.Smerds[idx].Constructor.Name)
		if err != nil {
			return deps, errors.Wrap(err, "error inspecting container")
		}

		if cont.ID != "" {
			deps.Smerds[idx].RunningContainer = &cont
		}
	}

	for idx := range deps.Volumes {
		vol, err := m.docker.VolumeInspect(ctx, deps.Volumes[idx].Constructor.Volume)
		if err != nil {
			if strings.Contains(err.Error(), "no such volume") {
				continue
			}
			return deps, errors.Wrapf(err, "error inspecting volume: %s", deps.Volumes[idx].Constructor.Volume)
		}

		if vol.Name != "" {
			deps.Volumes[idx].ExistingVolume = &vol
		}
	}

	return deps, nil
}
