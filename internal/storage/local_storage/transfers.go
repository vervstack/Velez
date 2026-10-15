package local_storage

import (
	"context"
	"database/sql"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/storage/postgres/generated/transfers_queries"
)

type transferToken struct {
	expiresAt time.Time
	isUsed    bool
}

// memoryTransfers is the single-node/dev storage.TransfersStorage: there is no
// velez.transfers table, so history lives in memory and is lost on restart.
type memoryTransfers struct {
	mu        sync.Mutex
	tokens    map[string]*transferToken
	transfers map[uuid.UUID]domain.Transfer
}

func newTransfersStorage() *memoryTransfers {
	return &memoryTransfers{
		tokens:    make(map[string]*transferToken),
		transfers: make(map[uuid.UUID]domain.Transfer),
	}
}

func (m *memoryTransfers) CreateToken(_ context.Context, tokenHash string, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tokens[tokenHash] = &transferToken{expiresAt: expiresAt}

	return nil
}

func (m *memoryTransfers) ConsumeToken(_ context.Context, tokenHash string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	token, isFound := m.tokens[tokenHash]
	if !isFound || token.isUsed || !token.expiresAt.After(time.Now()) {
		return false, nil
	}

	token.isUsed = true

	return true, nil
}

func (m *memoryTransfers) InsertTransfer(_ context.Context, transfer domain.Transfer) (domain.Transfer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	transfer.Uuid = uuid.New()
	transfer.CreatedAt = now
	transfer.UpdatedAt = now

	m.transfers[transfer.Uuid] = transfer

	return transfer, nil
}

func (m *memoryTransfers) GetTransfer(_ context.Context, id uuid.UUID) (sql.Null[domain.Transfer], error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	transfer, isFound := m.transfers[id]
	if !isFound {
		return sql.Null[domain.Transfer]{}, nil
	}

	return sql.Null[domain.Transfer]{V: transfer, Valid: true}, nil
}

func (m *memoryTransfers) UpdateProgress(_ context.Context, id uuid.UUID, totalBytes, doneBytes int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	transfer, isFound := m.transfers[id]
	if !isFound {
		return nil
	}

	transfer.TotalBytes = totalBytes
	transfer.DoneBytes = doneBytes
	transfer.UpdatedAt = time.Now()

	m.transfers[id] = transfer

	return nil
}

func (m *memoryTransfers) Finish(
	_ context.Context, id uuid.UUID, status transfers_queries.VelezTransferStatus, errMessage string,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	transfer, isFound := m.transfers[id]
	if !isFound {
		return nil
	}

	transfer.Status = status
	transfer.Error = errMessage
	transfer.UpdatedAt = time.Now()

	m.transfers[id] = transfer

	return nil
}

func (m *memoryTransfers) List(_ context.Context) ([]domain.Transfer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make([]domain.Transfer, 0, len(m.transfers))
	for _, transfer := range m.transfers {
		out = append(out, transfer)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})

	return out, nil
}

func (m *memoryTransfers) FailRunning(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, transfer := range m.transfers {
		isActive := transfer.Status == transfers_queries.VelezTransferStatusPENDING ||
			transfer.Status == transfers_queries.VelezTransferStatusRUNNING
		if !isActive {
			continue
		}

		transfer.Status = transfers_queries.VelezTransferStatusFAILED
		transfer.Error = "interrupted by restart"
		transfer.UpdatedAt = time.Now()

		m.transfers[id] = transfer
	}

	return nil
}
