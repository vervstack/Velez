package configuration_synchronizer

import (
	"context"

	"go.redsock.ru/rerrors"
	"go.vervstack.ru/matreshka/pkg/matreshka_api"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients/matreshka"
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

	//for {
	//updates, err := s.stream.Recv()
	//if err != nil {
	//	if !rerrors.Is(err, io.EOF) {
	//		logrus.Errorf("error recieving message from stream %s", err)
	//		continue
	//	}
	//
	//	return nil
	//}
	//envVars := make([]string, len(updates.Changes))
	//
	//s.updatesChan <- updates
	//}

	return nil
}

func (s *Synchronizer) Updates() <-chan []string {
	return s.updatesChan
}

func (s *Synchronizer) Stop() error {
	return s.stream.CloseSend()
}
