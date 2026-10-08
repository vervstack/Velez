package runneraas

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

func Test_CreateRunner_RejectsInvalidNameBeforeAnyWork(t *testing.T) {
	svc := New(nil, nil, nil, nil, nil)

	req := domain.CreateRunnerReq{Name: "Bad Name!"}

	err := svc.CreateRunner(t.Context(), req)
	require.ErrorIs(t, err, user_errors.ErrInvalidInstanceName)
}
