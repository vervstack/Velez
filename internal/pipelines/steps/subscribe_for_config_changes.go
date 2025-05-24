package steps

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/service"
	"go.vervstack.ru/Velez/pkg/velez_api"
)

type subscribeForConfigChangesStep struct {
	cfg service.ConfigurationService
	req *velez_api.CreateSmerd_Request
}

func SubscribeForConfigChanges(
	srv service.Services, req *velez_api.CreateSmerd_Request,
) *subscribeForConfigChangesStep {
	return &subscribeForConfigChangesStep{
		cfg: srv.ConfigurationService(),
		req: req,
	}
}

func (s *subscribeForConfigChangesStep) Do(_ context.Context) error {
	err := s.cfg.SubscribeOnChanges(s.req.Name)
	if err != nil {
		logrus.Error(rerrors.Wrap(err, "error handling subscription on service with name: "+s.req.Name))
	}

	return nil
}
