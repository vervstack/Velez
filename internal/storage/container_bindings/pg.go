package container_bindings

import (
	"context"
	"database/sql"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage"
	"go.vervstack.ru/Velez/internal/storage/environments"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/container_bindings_queries"
)

type pgStorage struct {
	querier *container_bindings_queries.Queries
}

func NewPg(db *sql.DB) storage.ContainerBindingsStorage {
	return &pgStorage{
		querier: container_bindings_queries.New(db),
	}
}

func (p *pgStorage) WithTx(tx *sql.Tx) storage.ContainerBindingsStorage {
	return &pgStorage{
		querier: p.querier.WithTx(tx),
	}
}

func (p *pgStorage) Upsert(ctx context.Context, binding domain.ContainerBinding) error {
	params := container_bindings_queries.UpsertContainerBindingParams{
		ServiceID:     binding.ServiceId,
		NodeID:        binding.NodeId,
		Environment:   environmentKey(binding.Environment),
		ContainerName: binding.ContainerName,
		IsSidecar:     binding.IsSidecar,
	}

	err := p.querier.UpsertContainerBinding(ctx, params)
	if err != nil {
		return rerrors.Wrap(err, "error upserting container binding")
	}

	return nil
}

func (p *pgStorage) ListByNode(
	ctx context.Context, nodeId int32, environment string,
) ([]domain.ContainerBinding, error) {
	params := container_bindings_queries.ListContainerBindingsByNodeParams{
		NodeID:      nodeId,
		Environment: environmentKey(environment),
	}

	rows, err := p.querier.ListContainerBindingsByNode(ctx, params)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing container bindings by node")
	}

	out := make([]domain.ContainerBinding, 0, len(rows))
	for _, row := range rows {
		binding := domain.ContainerBinding{
			Id:            row.ID,
			ServiceId:     row.ServiceID,
			ServiceName:   row.ServiceName,
			NodeId:        row.NodeID,
			Environment:   row.Environment,
			ContainerName: row.ContainerName,
			IsSidecar:     row.IsSidecar,
		}

		out = append(out, binding)
	}

	return out, nil
}

func (p *pgStorage) Delete(ctx context.Context, nodeId int32, environment, containerName string) error {
	params := container_bindings_queries.DeleteContainerBindingParams{
		NodeID:        nodeId,
		Environment:   environmentKey(environment),
		ContainerName: containerName,
	}

	err := p.querier.DeleteContainerBinding(ctx, params)
	if err != nil {
		return rerrors.Wrap(err, "error deleting container binding")
	}

	return nil
}

// environmentKey stores and matches bindings by environment name; an empty
// name is the default environment, as everywhere else on the wire.
func environmentKey(environment string) string {
	if environment == "" {
		return environments.DefaultEnvironmentName
	}

	return environment
}
