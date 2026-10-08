package pgaas

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

func Test_CreatePgInstance_RejectsInvalidNameBeforeAnyWork(t *testing.T) {
	svc := New(nil, nil, nil, nil)

	req := domain.CreatePgInstanceReq{Name: "Bad Name!"}

	err := svc.CreatePgInstance(t.Context(), req)
	require.ErrorIs(t, err, user_errors.ErrInvalidInstanceName)
}
