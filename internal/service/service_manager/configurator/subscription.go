package configurator

import (
	"io"
	"time"

	"github.com/sirupsen/logrus"
	"go.redsock.ru/rerrors"
	api "go.verv.tech/matreshka-be/pkg/matreshka_be_api"

	"github.com/godverv/Velez/internal/domain"
)

func (c *Configurator) SubscribeOnChanges(serviceNames ...string) error {
	subReq := &api.SubscribeOnChanges_Request{
		SubscribeServiceNames: serviceNames,
	}
	err := c.subscriptionStream.Send(subReq)
	if err != nil {
		return rerrors.Wrap(err, "error sending subscription request to stream")
	}

	return nil
}

func (c *Configurator) UnsubscribeFromChanges(serviceNames ...string) error {
	unsubReq := &api.SubscribeOnChanges_Request{
		UnsubscribeServiceNames: serviceNames,
	}

	err := c.subscriptionStream.Send(unsubReq)
	if err != nil {
		return rerrors.Wrap(err, "error sending subscription request to stream")
	}

	return nil
}

func handleSubscriptionStream(stream api.MatreshkaBeAPI_SubscribeOnChangesClient) chan domain.ConfigurationPatch {
	errorsCount := 3

	patchesChan := make(chan domain.ConfigurationPatch)
	defer close(patchesChan)

	go func() {
		for {
			changes, err := stream.Recv()
			if err != nil {
				if !rerrors.Is(err, io.EOF) {
					logrus.Errorf("Error recieving changes from matreshka subscription stream: %s", err)
					errorsCount--
					time.Sleep(time.Second)
					if errorsCount <= 0 {
						return
					}
					continue
				}
				return
			}

			patch := domain.ConfigurationPatch{
				ServiceName: changes.ServiceName,
			}

			switch ch := changes.Changes.(type) {
			case *api.SubscribeOnChanges_Response_EnvVariables:
				patch.EnvVarsMap = make(map[string]*string, len(ch.EnvVariables.EnvVariables))

				for _, node := range ch.EnvVariables.EnvVariables {
					patch.EnvVarsMap[node.Name] = node.Value
				}
			}

			patchesChan <- patch
		}
	}()

	return patchesChan
}
