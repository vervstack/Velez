package steps

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/service"
)

type subscribeForConfigChangesStep struct {
	cfg  service.ConfigurationService
	name string
}

func SubscribeForConfigChanges(
	srv service.Services,
	name string,
) *subscribeForConfigChangesStep {
	return &subscribeForConfigChangesStep{
		cfg:  srv.ConfigurationService(),
		name: name,
	}
}

func (s *subscribeForConfigChangesStep) Do(_ context.Context) error {
	err := s.cfg.SubscribeOnChanges(s.name)
	if err != nil {
		logrus.Error(rerrors.Wrap(err, "error handling subscription on service with name: "+s.name))
	}

	return nil
}
