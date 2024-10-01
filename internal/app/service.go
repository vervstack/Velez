package app

import (
	"context"

	"github.com/Red-Sock/toolbox/closer"
	"github.com/sirupsen/logrus"

	"github.com/godverv/Velez/internal/service"
	"github.com/godverv/Velez/internal/service/service_manager"
	"github.com/godverv/Velez/pkg/velez_api"
)

func (c *Custom) InitServiceManager() {
	c.Services = service_manager.New(c.InternalClients, c.ExternalClients)

	logrus.Warn("shut down on exit is ", c.Cfg.GetEnvironment().ShutDownOnExit)

	if c.Cfg.GetEnvironment().ShutDownOnExit {
		closer.Add(smerdsDropper(c.Services))
	}
}

func smerdsDropper(manager service.Services) func() error {
	return func() error {
		logrus.Infof("ShutDownOnExit env variable is set to TRUE. Dropping launched smerds")
		logrus.Infof("Listing launched smerds")
		ctx := context.Background()

		smerds, err := manager.ListSmerds(ctx, &velez_api.ListSmerds_Request{})
		if err != nil {
			return err
		}

		names := make([]string, 0, len(smerds.Smerds))

		for _, sm := range smerds.Smerds {
			names = append(names, sm.Name)
		}

		logrus.Infof("%d smerds is active. %v", len(smerds.Smerds), names)

		dropReq := &velez_api.DropSmerd_Request{
			Uuids: make([]string, len(smerds.Smerds)),
		}

		for i := range smerds.Smerds {
			dropReq.Uuids[i] = smerds.Smerds[i].Uuid
		}

		logrus.Infof("Dropping %d smerds", len(smerds.Smerds))

		dropSmerds, err := manager.DropSmerds(ctx, dropReq)
		if err != nil {
			return err
		}

		logrus.Infof("%d smerds dropped successfully", len(dropSmerds.Successful))
		if len(dropSmerds.Successful) != 0 {
			logrus.Infof("Dropped smerds: %v", dropSmerds.Successful)
		}

		if len(dropSmerds.Failed) != 0 {
			logrus.Errorf("%d smerds failed to drop", len(dropSmerds.Failed))
			for _, f := range dropSmerds.Failed {
				logrus.Errorf("error dropping %s. Cause: %s", f.Uuid, f.Cause)
			}
		}

		return nil
	}
}
