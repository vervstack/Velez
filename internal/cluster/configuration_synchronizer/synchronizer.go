package configuration_synchronizer

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/matreshka"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"
)

type Synchronizer struct {
	stream matreshka_api.MatreshkaBeAPI_SubscribeOnChangesClient

	updatesChan chan []string
}

func New(ctx context.Context, matreshkaClient matreshka.Client) (*Synchronizer, error) {
	s := &Synchronizer{
		updatesChan: make(chan []string),
	}

	var err error

	s.stream, err = matreshkaClient.SubscribeOnChanges(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error during configuration subscription")
	}

	return s, nil
}

func (s *Synchronizer) Start() error {
	// for {
	// updates, err := s.stream.Recv()
	// if err != nil {
	// if !rerrors.Is(err, io.EOF) {
	// logrus.Errorf("error receiving message from stream %s", err)
	// continue
	// }
	//
	// return nil
	// }
	// envVars := make([]string, len(updates.Changes))
	//
	// s.updatesChan <- updates
	// }
	return nil
}

func (s *Synchronizer) Updates() <-chan []string {
	return s.updatesChan
}

func (s *Synchronizer) Stop() error {
	err := s.stream.CloseSend()
	if err != nil {
		return rerrors.Wrap(err, "error closing stream")
	}

	return nil
}
