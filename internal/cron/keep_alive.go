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
	_ = start(s)
	for {
		select {
		case <-t.C:
			if !start(s) {
				return
			}
		case <-ctx.Done():
			logrus.Infof(`Got termination call. Killing "%s"`, s.GetName())

			err := s.Kill()
			if err != nil {
				logrus.Errorf(`Error killing "%s"`, s.GetName())
			} else {
				logrus.Infof(`Successfully killed "%s"`, s.GetName())
			}

			return
		}
	}
}

func start(s KeepAliveService) bool {
	ok, err := s.IsAlive()
	if err != nil {
		logrus.Errorf(`error checking if "%s" alive`, s.GetName())
		return false
	}
	if ok {
		return true
	}
	_ = s.Kill()

	err = s.Start()
	if err != nil {
		logrus.Errorf(`error keeping "%s" alive: %s `, s.GetName(), err)
		return false
	}
	logrus.Infof(`successfully started "%s"`, s.GetName())

	return true
}
