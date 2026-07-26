package sqldb

import (
	"context"
	"database/sql"
	stdErrors "errors"

	errors "go.redsock.ru/rerrors"
)

type TxManager struct {
	conn *sql.DB
}

func NewTxManager(conn *sql.DB) *TxManager {
	return &TxManager{
		conn: conn,
	}
}

func (m *TxManager) Execute(do func(tx *sql.Tx) error) error {
	ctx := context.Background()

	tx, err := m.conn.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err)
	}

	err = do(tx)

	var commitErr error
	if err != nil {
		commitErr = tx.Rollback()
	} else {
		commitErr = tx.Commit()
	}

	if commitErr != nil || err != nil {
		return stdErrors.Join(err, commitErr)
	}

	return nil
}
