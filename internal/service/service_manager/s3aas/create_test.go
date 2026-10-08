package s3aas

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/user_errors"
)

func Test_CreateInstance_RejectsInvalidInstanceName(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"uppercase", "MyS3"},
		{"single char", "a"},
		{"dot", "my.s3"},
		{"empty", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := New(nil, nil, nil, nil, nil, nil, nil)
			req := &velez_api.CreateS3Instance_Request{Name: tc.input}

			err := svc.CreateInstance(context.Background(), req)

			require.ErrorIs(t, err, user_errors.ErrInvalidInstanceName)
		})
	}
}
