package configuration

import (
	"context"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"
	"go.redsock.ru/toolbox"
	"go.redsock.ru/toolbox/closer"
	"go.redsock.ru/toolbox/keep_alive"
	"go.vervstack.ru/makosh/pkg/makosh_be"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients"
	"go.vervstack.ru/Velez/internal/clients/cluster_clients/matreshka"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/cluster/env"
	"go.vervstack.ru/Velez/internal/cluster/env/container_service_task"
	"go.vervstack.ru/Velez/internal/config"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/pipelines"
	"go.vervstack.ru/Velez/internal/pipelines/steps"
)

// sharedInstanceCtxKey is an unexported context key used to hand an
// already-running matreshka container to SetupMatreshka, so it skips
// creating (and keep-alive managing) a second one. See SharedInstance and
// WithSharedInstance — this only exists for test fixtures that need many
// SetupMatreshka callers within one process to share exactly one container,
// same as production's "exactly one matreshka instance per node" assumption.
type sharedInstanceCtxKey struct{}

// SharedInstance is a handle to an already-running matreshka container and
// its background keep-alive loop, created once via StartSharedInstance.
// Production code never creates one of these — cluster.Setup calls
// SetupMatreshka directly, one instance per node. It exists so a test
// binary (see tests/e2e/main_test.go) can create a single shared instance
// and hand it to every SetupMatreshka call via WithSharedInstance.
type SharedInstance struct {
	task *container_service_task.TaskV2
}

// StartSharedInstance creates the matreshka sidecar container and its
// keep-alive loop exactly once, the same way SetupMatreshka would for a
// single node. Intended for a package-level test fixture that then hands
// the result to every SetupMatreshka call via WithSharedInstance.
func StartSharedInstance(ctx context.Context, cfg config.Config, nc node_clients.NodeClients) (*SharedInstance, error) {
	key, err := initKey(ctx, nc)
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting key")
	}

	task, _, err := startContainer(ctx, cfg, nc, key)
	if err != nil {
		return nil, rerrors.Wrap(err, "error starting shared matreshka instance")
	}

	instance := &SharedInstance{task: task}

	return instance, nil
}

// Stop removes the shared container. It intentionally does not call the
// keep-alive loop's Stop() — go.redsock.ru/toolbox@v0.0.13's
// AliveKeeper.Stop() blocks on a channel that a stopped time.Ticker never
// closes, so it never returns. The background ticker goroutine dies with
// the process on exit regardless, same as any other leftover goroutine in
// a test binary.
func (s *SharedInstance) Stop() error {
	return s.task.Kill()
}

// WithSharedInstance returns a context that makes SetupMatreshka reuse an
// already-running shared container instead of creating (and keep-alive
// managing) a second one. See SharedInstance's doc comment for the
// single-process constraint this implies.
func WithSharedInstance(ctx context.Context, instance *SharedInstance) context.Context {
	return context.WithValue(ctx, sharedInstanceCtxKey{}, instance.task)
}

func sharedTaskFromContext(ctx context.Context) (*container_service_task.TaskV2, bool) {
	task, ok := ctx.Value(sharedInstanceCtxKey{}).(*container_service_task.TaskV2)
	return task, ok
}

// startContainer builds the matreshka container request, creates the
// container, and starts its keep-alive loop. Extracted out of
// SetupMatreshka so StartSharedInstance can reuse the exact same
// container-config logic instead of duplicating it.
func startContainer(ctx context.Context, cfg config.Config, nc node_clients.NodeClients, key string) (*container_service_task.TaskV2, *keep_alive.AliveKeeper, error) {
	//region Create Container request
	matreshkaContainerCfg := container.CreateRequest{
		Config: &container.Config{
			Hostname: Name,
			Env: []string{
				passEnv + ":" + key,
			},
			ExposedPorts: map[nat.Port]struct{}{},
			Image:        toolbox.Coalesce(cfg.Environment.MatreshkaImage, image),
			Volumes: map[string]struct{}{
				Name: {},
			},
			Labels: map[string]string{
				labels.VervServiceLabel:  "true",
				labels.ComposeGroupLabel: Name,
			},
		},
		HostConfig: &container.HostConfig{
			PortBindings: map[nat.Port][]nat.PortBinding{
				grpcPort: {
					{
						HostPort: grpcPort,
					},
				},
			},
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: Name,
					Target: defaultDataPath,
				},
			},
		},
	}
	//endregion

	if cfg.Environment.MatreshkaPort > 0 {
		matreshkaContainerCfg.Config.ExposedPorts[grpcPort] = struct{}{}

		matreshkaContainerCfg.HostConfig.PortBindings = map[nat.Port][]nat.PortBinding{
			grpcPort: {
				{
					HostPort: strconv.Itoa(cfg.Environment.MatreshkaPort),
				},
			},
		}
	}

	task, err := container_service_task.NewTaskV2(nc.Docker(), matreshkaContainerCfg)
	if err != nil {
		return nil, nil, rerrors.Wrap(err, "error creating verv service task")
	}

	log.Info().Msg("Preparing matreshka service background task")

	ka := keep_alive.KeepAlive(task, keep_alive.WithCancel(ctx.Done()))

	return task, ka, nil
}

func SetupMatreshka(
	ctx context.Context,
	cfg config.Config,
	nc node_clients.NodeClients,
	sdClient cluster_clients.ServiceDiscovery,
	vcnClient cluster_clients.VervClosedNetworkClient,
) (matreshka.Client, error) {
	key, err := initKey(ctx, nc)
	if err != nil {
		return nil, rerrors.Wrap(err, "error getting key")
	}

	task, hasSharedInstance := sharedTaskFromContext(ctx)
	if !hasSharedInstance {
		var ka *keep_alive.AliveKeeper

		task, ka, err = startContainer(ctx, cfg, nc, key)
		if err != nil {
			return nil, rerrors.Wrap(err, "error creating verv service task")
		}

		if cfg.Environment.ShutDownOnExit {
			closer.Add(func() error {
				ka.Stop()
				return nil
			})
		}
	}

	mClient, err := newClient(nc)
	if err != nil {
		return nil, rerrors.Wrap(err, "error creating matreshka client")
	}

	vcnReq := domain.ConnectServiceToVcn{
		ServiceName: Name,
	}

	if !env.IsInContainer() {
		_, port := task.GetPortBinding(grpcPort)
		r := &makosh_be.UpsertEndpoints_Request{
			Endpoints: []*makosh_be.Endpoint{
				{
					ServiceName: "matreshka",
					Addrs:       []string{"0.0.0.0:" + port},
				},
			},
		}
		_, err = sdClient.UpsertEndpoints(ctx, r)
		if err != nil {
			return nil, rerrors.Wrap(err, "error upserting endpoints")
		}

		return mClient, nil
	}

	runner := pipelines.ConnectServiceToVpn(vcnReq, nc, vcnClient, sdClient)
	err = runner.Run(ctx)
	if err != nil {
		if rerrors.Is(err, steps.ErrAlreadyExists) {
			return mClient, nil
		}

		if rerrors.Is(err, cluster_clients.ErrServiceIsDisabled) {
			return mClient, nil
		}

		return nil, rerrors.Wrap(err, "error connecting matreshka to verv closed network")
	}

	return mClient, nil
}
