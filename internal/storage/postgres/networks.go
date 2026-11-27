package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/service/user_errors"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/vcn_networks_queries"
)

type networksStorage struct {
	querier *vcn_networks_queries.Queries
}

func newNetworksStorage(db *sql.DB) *networksStorage {
	return &networksStorage{querier: vcn_networks_queries.New(db)}
}

func (n *networksStorage) Create(ctx context.Context, network domain.Network) error {
	id, err := uuid.Parse(network.Id)
	if err != nil {
		return rerrors.Wrap(err, "error parsing network id")
	}

	params := vcn_networks_queries.CreateVcnNetworkParams{
		ID:          id,
		Name:        network.Name,
		Provider:    network.Provider,
		Source:      network.Source,
		Url:         network.Url,
		Token:       network.Token,
		ServiceName: network.ServiceName,
		ContainerID: network.ContainerId,
		NodeName:    network.NodeName,
		CreatedAt:   network.CreatedAt,
	}

	err = n.querier.CreateVcnNetwork(ctx, params)
	if err != nil {
		return rerrors.Wrap(wrapPgErr(err), "error creating network")
	}

	return nil
}

func (n *networksStorage) Get(ctx context.Context, id string) (domain.Network, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return domain.Network{}, rerrors.Wrap(user_errors.ErrStorageNotFound)
	}

	row, err := n.querier.GetVcnNetworkById(ctx, parsed)
	if err != nil {
		return domain.Network{}, rerrors.Wrap(wrapPgErr(err), "error getting network")
	}

	return networkFromRow(row), nil
}

func (n *networksStorage) List(ctx context.Context) ([]domain.Network, error) {
	rows, err := n.querier.ListVcnNetworks(ctx)
	if err != nil {
		return nil, rerrors.Wrap(wrapPgErr(err), "error listing networks")
	}

	networks := make([]domain.Network, 0, len(rows))
	for _, row := range rows {
		networks = append(networks, networkFromRow(row))
	}

	return networks, nil
}

func (n *networksStorage) Drop(ctx context.Context, id string) error {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return rerrors.Wrap(user_errors.ErrStorageNotFound)
	}

	count, err := n.querier.DropVcnNetworkById(ctx, parsed)
	if err != nil {
		return rerrors.Wrap(wrapPgErr(err), "error dropping network")
	}

	if count == 0 {
		return rerrors.Wrap(user_errors.ErrStorageNotFound)
	}

	return nil
}

func (n *networksStorage) CurrentNode(ctx context.Context) (string, error) {
	name, err := n.querier.GetCurrentNode(ctx)
	if err != nil {
		return "", rerrors.Wrap(wrapPgErr(err), "error looking up current node")
	}

	return name, nil
}

func networkFromRow(row vcn_networks_queries.VelezNetwork) domain.Network {
	return domain.Network{
		Id:          row.ID.String(),
		Name:        row.Name,
		Provider:    row.Provider,
		Source:      row.Source,
		Url:         row.Url,
		Token:       row.Token,
		ServiceName: row.ServiceName,
		ContainerId: row.ContainerID,
		NodeName:    row.NodeName,
		CreatedAt:   row.CreatedAt,
	}
}
