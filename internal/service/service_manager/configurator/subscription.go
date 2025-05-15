package configurator

import (
	"io"
	"time"

	"github.com/sirupsen/logrus"
	"go.redsock.ru/rerrors"
	api "go.vervstack.ru/matreshka/pkg/matreshka_be_api"

	"github.com/godverv/Velez/internal/domain"
)

func (c *Configurator) SubscribeOnChanges(serviceNames ...string) error {
	//subReq := &api.SubscribeOnChanges_Request{
	//	SubscribeServiceNames: serviceNames,
	//}
	//err := c.subscriptionStream.Send(subReq)
	//if err != nil {
	//	return rerrors.Wrap(err, "error sending subscription request to stream")
	//}

	return nil
}

func (c *Configurator) UnsubscribeFromChanges(serviceNames ...string) error {
	unsubReq := &api.SubscribeOnChanges_Request{
		UnsubscribeConfigNames: serviceNames,
	}

	err := c.subscriptionStream.Send(unsubReq)
	if err != nil {
		return rerrors.Wrap(err, "error sending subscription request to stream")
	}

	return nil
}

func (c *Configurator) GetUpdates() <-chan domain.ConfigurationPatch {
	return c.updatesChan
}

func handleSubscriptionStream(stream api.MatreshkaBeAPI_SubscribeOnChangesClient) chan domain.ConfigurationPatch {
	errorsCount := 3
	patchesChan := make(chan domain.ConfigurationPatch)

	go func() {
		defer close(patchesChan)

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
				ServiceName: changes.ConfigName,
			}
			for _, p := range changes.Patches {
				switch v := p.GetPatch().(type) {
				case *api.PatchConfig_Patch_UpdateValue:
					//	TODO implement
				case *api.PatchConfig_Patch_Rename:
				//	TODO implement
				case *api.PatchConfig_Patch_Delete:
				//	TODO implement
				default:
					_ = v
				}
			}
			if len(patch.EnvVarsMap) != 0 {
				patchesChan <- patch
			}
		}
	}()

	return patchesChan
}
