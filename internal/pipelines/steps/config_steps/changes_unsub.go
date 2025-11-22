package config_steps

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/service"
)

type unsubscribeForConfigChangesStep struct {
	cfg  service.ConfigurationService
	name string
}

func UnSubscribeForConfigChanges(
	srv service.Services,
	name string,
) *unsubscribeForConfigChangesStep {
	return &unsubscribeForConfigChangesStep{
		cfg:  srv.ConfigurationService(),
		name: name,
	}
}

func (s *unsubscribeForConfigChangesStep) Do(_ context.Context) error {
	err := s.cfg.UnsubscribeFromChanges(s.name)
	if err != nil {
		logrus.Error(rerrors.Wrap(err, "error handling subscription cancel on service with name: "+s.name))
	}

	return nil
}
