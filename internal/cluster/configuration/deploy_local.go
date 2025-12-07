package configuration

import (
	"strings"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"go.redsock.ru/rerrors"
	version "go.vervstack.ru/matreshka/config"
	"golang.org/x/net/context"

	"go.vervstack.ru/Velez/internal/clients/node_clients"
)

const (
	Name         = "matreshka"
	defaultImage = "vervstack/matreshka"
	grpcPort     = "50049"

	passEnv = "pass"

	defaultDataPath = "/app/data"
)

var image string

func init() {
	image = defaultImage + ":" + version.GetVersion()
}

func initKey(ctx context.Context, nodeClients node_clients.NodeClients) (string, error) {
	keyFromLocalState := nodeClients.LocalStateManager().Get().MatreshkaKey

	keyFromCont, err := getKeyFromMatreshkaContainerEnv(ctx, nodeClients.Docker().Client())
	if err != nil {
		return "", rerrors.Wrap(err, "error getting key from container")
	}

	if keyFromCont == "" {
		return keyFromLocalState, nil
	}

	logrus.Infof("Using key from local state: %s", keyFromLocalState)

	stateManager := nodeClients.LocalStateManager()
	localState := stateManager.Get()
	localState.MatreshkaKey = keyFromCont
	stateManager.Set(localState)

	return keyFromCont, nil
}

func getKeyFromMatreshkaContainerEnv(ctx context.Context, docker client.APIClient) (string, error) {
	cont, err := docker.ContainerInspect(ctx, Name)
	if err != nil {
		if !cerrdefs.IsNotFound(err) {
			return "", rerrors.Wrap(err, "")
		}

		return "", nil
	}

	for _, e := range cont.Config.Env {
		if strings.HasPrefix(e, passEnv) {
			return e[len(passEnv)+1:], nil
		}
	}
	return "", nil
}
