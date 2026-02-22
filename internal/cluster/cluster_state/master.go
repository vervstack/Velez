package cluster_state

import (
	"context"
	"time"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/cluster_clients/state"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/clients/sqldb"
)

func SetupMasterPg(
	ctx context.Context,
	nodeClients node_clients.NodeClients,
) error {
	dockerClient := nodeClients.Docker().Client()
	// TODO think about multi cluster on one node - must support multiple postgres instances on one physical node
	// maybe get cont name from local state?
	containerName := state.PgName

	contInspect, err := dockerClient.ContainerInspect(ctx, containerName)
	if err != nil {
		if errdefs.IsNotFound(err) {
			log.Warn().
				Err(err).
				Str("container_name", containerName).
				Msg("PgRootDsn was defined but postgres for cluster state isn't running on this node")
			return nil
		}

		return rerrors.Wrap(err, "error inspecting postgres cluster state container")
	}

	if contInspect.State == nil {
		return rerrors.New("Postgres container for Cluster state exists but don't have a state")
	}

	if contInspect.State.Status != container.StateRunning {
		log.Info().
			Str("container_name", containerName).
			Str("state", contInspect.State.Status).
			Msg("Found Postgres container for Cluster state not running. Trying to start it")

		startOps := container.StartOptions{}

		err = dockerClient.ContainerStart(ctx, contInspect.ID, startOps)
		if err != nil {
			return rerrors.Wrap(err, "Failed to start Postgres container for Cluster state. Fallback to noImpl")
		}

		contInspect, err = dockerClient.ContainerInspect(ctx, containerName)
		if err != nil {
			return rerrors.Wrap(err, "error inspecting postgres cluster state container after start")
		}
	}

	for try := range 5 {
		contInspect, err = dockerClient.ContainerInspect(ctx, containerName)
		if err != nil {
			return rerrors.Wrap(err, "error inspecting postgres cluster state container after start")
		}

		if contInspect.State.Health.Status == container.Healthy {
			break
		}

		time.Sleep(time.Second * 5 * time.Duration(try))
	}

	if contInspect.State.Health.Status != container.Healthy {
		return rerrors.Wrap(err, "Postgres container for cluster state isn't healthy. Falling back to noImpl")
	}

	localState := nodeClients.LocalStateManager().Get().ClusterState.PgRootDsn

	err = sqldb.RollMigration(localState)
	if err != nil {
		return rerrors.Wrap(err, "Failed to roll Postgres migration")
	}

	return nil
}
