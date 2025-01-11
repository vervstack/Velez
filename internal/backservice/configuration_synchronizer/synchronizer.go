package configuration_synchronizer

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.redsock.ru/rerrors"
	"go.verv.tech/matreshka-be/pkg/matreshka_be_api"

	"github.com/godverv/Velez/internal/clients/matreshka"
)

type Synchronizer struct {
	configApi matreshka.Client

	nodeName string

	stopCall func() error
}

func New(matreshkaClient matreshka.Client, nodeName string) *Synchronizer {
	return &Synchronizer{
		configApi: matreshkaClient,
		// TODO: VERV-123
		nodeName: "steel_owl",
	}
}

func (s *Synchronizer) Start(ctx context.Context) error {
	initReq := &matreshka_be_api.RefreshConfigHook_Init{
		NodeName: s.nodeName,
	}

	updateLoop, err := s.configApi.RefreshConfigHook(ctx, initReq)
	if err != nil {
		return rerrors.Wrap(err, "error connecting to configuration")
	}

	s.stopCall = updateLoop.CloseSend

	for {
		msg, err := updateLoop.Recv()
		if err != nil {
			logrus.Error("error looping over configuration service update events")
			break
		}

		_ = msg
	}

	return nil
}

func (s *Synchronizer) Stop() error {
	return s.stopCall()
}
