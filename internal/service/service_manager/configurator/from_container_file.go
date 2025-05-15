package configurator

import (
	"archive/tar"
	"context"
	stderrs "errors"
	"io"

	errors "go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka"
)

const defaultConfigPath = "/app/config/config.yaml"

func (c *Configurator) GetFromContainer(ctx context.Context, contId string) (conf matreshka.AppConfig, err error) {
	rc, _, err := c.dockerAPI.CopyFromContainer(ctx, contId, defaultConfigPath)
	if err != nil {
		return matreshka.AppConfig{}, errors.Wrap(err, "error coping from container")
	}

	defer func() {
		errClose := rc.Close()
		if errClose == nil {
			return
		}

		if err == nil {
			err = errClose
		} else {
			stderrs.Join(err, errClose)
		}
	}()

	reader := tar.NewReader(rc)
	_, err = reader.Next()
	if err != nil {
		return matreshka.AppConfig{}, errors.Wrap(err, "error getting next")
	}

	res, err := io.ReadAll(reader)
	if err != nil {
		return matreshka.AppConfig{}, errors.Wrap(err, "error reading config from tar")
	}

	confVal := matreshka.NewEmptyConfig()
	err = confVal.Unmarshal(res)
	if err != nil {
		return matreshka.AppConfig{}, errors.Wrap(err, "error unmarshalling config")
	}

	return confVal, nil
}
