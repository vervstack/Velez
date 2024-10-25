package configuration

import (
	"sync"
	"time"

	"github.com/Red-Sock/toolbox/keep_alive"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/godverv/Velez/internal/clients"
	"github.com/godverv/Velez/internal/config"
)

type MatreshkaConnect struct {
	Addr  string
	Token string
}

var initOnce sync.Once
var conn MatreshkaConnect

func InitInstance(ctx context.Context, cfg config.Config, clients clients.NodeClients) MatreshkaConnect {
	initOnce.Do(func() {
		conn = initInstance(ctx, cfg, clients)
	})

	return conn
}

func initInstance(ctx context.Context, cfg config.Config, clients clients.NodeClients,
) MatreshkaConnect {
	makoshBackgroundTask, err := newTask(cfg, clients)
	if err != nil {
		logrus.Fatalf("error creating configuration service background task: %s", err)
	}

	logrus.Info("Starting configuration service background task")
	go keep_alive.KeepAlive(makoshBackgroundTask, keep_alive.WithCancel(ctx.Done()))

	t := time.NewTicker(time.Second * 2)
	for range t.C {
		isAlive, err := makoshBackgroundTask.IsAlive()
		if err != nil {
			logrus.Fatalf("error during setting up configuration service %s", err)
		}

		if isAlive {
			t.Stop()
			break
		}
	}

	return MatreshkaConnect{
		Addr: makoshBackgroundTask.Address,
		// TODO add token for authorization
		Token: "",
	}
}
