package postgres

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/sqldb"
	"go.vervstack.ru/Velez/internal/domain"
	pg_queries "go.vervstack.ru/Velez/internal/storage/postgres/generated/nodes_queries"
	"go.vervstack.ru/Velez/internal/utils/common"
)

const (
	defaultNodeListLimit = 3
)

type nodeStorage struct {
	db      sqldb.DB
	querier pg_queries.Querier
}

func newNodeStorage(db sqldb.DB) *nodeStorage {
	return &nodeStorage{
		querier: pg_queries.New(db),
		db:      db,
	}
}

func (n *nodeStorage) InitNode(ctx context.Context, region string) error {
	err := n.querier.InitNode(ctx, region)
	if err != nil {
		return rerrors.Wrap(err, "error initializing new node")
	}

	storedRegion, err := n.querier.GetOwnRegion(ctx)
	if err != nil {
		return rerrors.Wrap(err, "error reading back stored node region")
	}

	if region != "" && region != storedRegion {
		log.Warn().
			Str("config_region", region).
			Str("stored_region", storedRegion).
			Msg("node region in config differs from region already stored in postgres; keeping the stored value")
	}

	return nil
}

func (n *nodeStorage) UpdateOnline(ctx context.Context, cpuPercent, memPercent float64) error {
	params := pg_queries.UpdateOnlineParams{
		CpuPercent: sql.NullFloat64{Float64: cpuPercent, Valid: true},
		MemPercent: sql.NullFloat64{Float64: memPercent, Valid: true},
	}

	err := n.querier.UpdateOnline(ctx, params)
	if err != nil {
		return rerrors.Wrap(err, "error updating node's last online")
	}

	return nil
}

func (n *nodeStorage) List(ctx context.Context, req domain.ListNodesReq) (domain.NodesList, error) {
	builder := sq.Select().
		From("velez.nodes")

	// TODO filters
	totalNodes, err := n.countTotal(ctx, builder)
	if err != nil {
		return domain.NodesList{}, rerrors.Wrap(err, "error counting nodes")
	}

	builder = builder.Columns(
		"id",
		"name",
		"last_online",
		"is_enabled",
		"addr",
		"cpu_percent",
		"mem_percent",
		"region",
	)

	servicesCountColumn := "(SELECT count(*) FROM velez.deployments d " +
		"WHERE d.node_id = velez.nodes.id AND d.status = 'RUNNING') AS services_count"

	builder = builder.Column(servicesCountColumn)

	if req.Paging.Limit == 0 {
		req.Paging.Limit = defaultNodeListLimit
	}

	builder = builder.
		Limit(min(req.Paging.Limit, totalNodes)).
		Offset(req.Paging.Offset)

	query, args, err := builder.ToSql()
	if err != nil {
		return domain.NodesList{}, rerrors.Wrap(err, "erorr building list query")
	}

	rows, err := n.db.QueryContext(ctx, query, args...)
	if err != nil {
		return domain.NodesList{}, rerrors.Wrap(err, "error querying list nodes")
	}

	defer common.CloseWithLog(rows.Close, "list nodes query")

	list := make([]domain.NodeBaseInfo, 0, totalNodes)

	for rows.Next() {
		var node domain.NodeBaseInfo

		node, err = scanNode(rows)
		if err != nil {
			return domain.NodesList{}, rerrors.Wrap(err, "error scanning list nodes")
		}

		list = append(list, node)
	}

	err = rows.Err()
	if err != nil {
		return domain.NodesList{}, rerrors.Wrap(err, "error iterating list nodes")
	}

	return domain.NodesList{
		Nodes: list,
		Total: totalNodes,
	}, nil
}

func (n *nodeStorage) countTotal(ctx context.Context, builder sq.SelectBuilder) (uint64, error) {
	var total uint64

	builder = builder.Column("count(*)")

	q, args, err := builder.ToSql()
	if err != nil {
		return 0, rerrors.Wrap(err, "error building count total query")
	}

	err = n.db.QueryRowContext(ctx, q, args...).
		Scan(&total)
	if err != nil {
		return 0, rerrors.Wrap(err, "error scanning count total query")
	}

	return total, nil
}

func scanNode(scannable sqldb.Scannable) (domain.NodeBaseInfo, error) {
	node := domain.NodeBaseInfo{}

	var cpuPercent sql.NullFloat64

	var memPercent sql.NullFloat64

	err := scannable.Scan(
		&node.Id,
		&node.Name,
		&node.LastOnline,
		&node.IsEnabled,
		&node.Addr,
		&cpuPercent,
		&memPercent,
		&node.Region,
		&node.ServicesCount,
	)
	if err != nil {
		return node, rerrors.Wrap(err, "error scanning result for node base info")
	}

	node.CpuPercent = cpuPercent.Float64
	node.MemPercent = memPercent.Float64

	return node, nil
}
