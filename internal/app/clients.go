package app

import (
	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/clients/managers"
)

func (a *App) MustInitExternalClients() {
	var err error

	a.ExternalClients, err = managers.NewExternalClients(a.Ctx, a.Cfg, a.InternalClients)
	if err != nil {
		logrus.Fatalf("error initializing external clients %s", err)
	}
}
