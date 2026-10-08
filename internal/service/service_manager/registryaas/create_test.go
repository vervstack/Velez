package registryaas

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/user_errors"
)

func Test_CreateRegistryInstance_RejectsInvalidInstanceNameBeforeAnyWork(t *testing.T) {
	svc := New(nil, nil, nil, nil, nil)

	err := svc.CreateRegistryInstance(t.Context(), domain.CreateRegistryInstanceReq{Name: "Bad Name"})

	require.ErrorIs(t, err, user_errors.ErrInvalidInstanceName)
}
