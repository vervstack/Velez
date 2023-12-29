package cron

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

type KeepAliveService interface {
	GetName() string
	GetDuration() time.Duration

	Start() error
	IsAlive() (bool, error)
	Kill() error
}

func KeepAlive(ctx context.Context, s KeepAliveService) {
	t := time.NewTicker(s.GetDuration())

	for {
		select {
		case <-t.C:
			ok, err := s.IsAlive()
			if err != nil {
				logrus.Errorf(`error checking if "%s" alive`, s.GetName())
			}
			if ok {
				continue
			}

			err = s.Start()
			if err != nil {
				logrus.Errorf(`error keeping "%s" alive `, s.GetName())
			}
			logrus.Infof(`successfully started "%s"`, s.GetName())

		case <-ctx.Done():
			logrus.Errorf(`got termination call. Killing "%s"`, s.GetName())

			err := s.Kill()
			if err != nil {
				logrus.Errorf(`error killing "%s"`, s.GetName())
			} else {
				logrus.Infof(`successfully killed "%s"`, s.GetName())
			}

			return
		}
	}
}
