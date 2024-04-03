package config_manager

import (
	"context"
	"os"
	"path"
	"strings"

	errors "github.com/Red-Sock/trace-errors"
	"github.com/godverv/matreshka"
	"github.com/godverv/matreshka-be/pkg/matreshka_api"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

func (c *Configurator) getConfig(ctx context.Context, serviceName string) (*matreshka.AppConfig, error) {
	g, gctx := errgroup.WithContext(ctx)

	var apiCfg, conCfg *matreshka.AppConfig
	g.Go(func() (err error) {
		apiCfg, err = c.fetchFromApi(gctx, serviceName)
		return
	})

	g.Go(func() (err error) {
		conCfg, err = c.readFromMount(serviceName)
		return
	})

	err := g.Wait()
	if err != nil {
		logrus.Warnf("error obtaining configs for service '%s' :%s", serviceName, err)
	}

	if apiCfg == conCfg {
		return nil, errors.Wrap(err, "no configs found")
	}

	if apiCfg == nil {
		return conCfg, nil
	}

	if conCfg == nil {
		return apiCfg, nil
	}

	res := matreshka.MergeConfigs(*apiCfg, *conCfg)

	return &res, nil
}

func (c *Configurator) fetchFromApi(ctx context.Context, serviceName string) (*matreshka.AppConfig, error) {
	var apiConfig *matreshka.AppConfig

	matreshkaConfig, err := c.matreshkaClient.GetConfigRaw(ctx,
		&matreshka_api.GetConfigRaw_Request{
			ServiceName: serviceName,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "error obtaining raw config")
	}

	err = apiConfig.Unmarshal(matreshkaConfig.Config)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling config from api")
	}

	return apiConfig, nil
}

func (c *Configurator) readFromMount(serviceName string) (*matreshka.AppConfig, error) {
	dirPath := c.getMountPoint(serviceName)
	dirs, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching service config dir")
	}

	if len(dirs) == 0 {
		return nil, nil
	}

	var configPaths = make([]string, 0, len(dirs))

	for _, item := range dirs {
		if item.IsDir() {
			continue
		}
		name := item.Name()
		if !strings.HasSuffix(name, ".yaml") {
			continue
		}

		configPaths = append(configPaths, path.Join(dirPath, name))
	}

	cfg, err := matreshka.ReadConfigs(configPaths...)
	if err != nil {
		return nil, errors.Wrap(err, "error reading configs from mounts")
	}

	return cfg, nil
}
