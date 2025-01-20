package deploy_steps

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.redsock.ru/rerrors"

	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/pkg/velez_api"
)

type subscribeForConfigChangesStep struct {
	cfg service.ConfigurationService
	req *velez_api.CreateSmerd_Request
}

func SubscribeForConfigChanges(cfg service.ConfigurationService, req *velez_api.CreateSmerd_Request) *subscribeForConfigChangesStep {
	return &subscribeForConfigChangesStep{
		cfg: cfg,
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
