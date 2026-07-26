package configurator

import (
	"go.redsock.ru/rerrors"
	api "go.vervstack.ru/matreshka/pkg/matreshka_api"

	"go.vervstack.ru/Velez/internal/domain"
)

func (c *Configurator) SubscribeOnChanges(serviceNames ...string) error {
	// subReq := &api.SubscribeOnChanges_Request{
	//	SubscribeServiceNames: serviceNames,
	//}
	// err := c.subscriptionStream.Send(subReq)
	// if err != nil {
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
