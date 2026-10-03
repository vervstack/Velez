package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_IsVelezImage(t *testing.T) {
	cases := []struct {
		name  string
		image string
		want  bool
	}{
		{"plain name", "velez", true},
		{"org and tag", "vervstack/velez:1.2.3", true},
		{"registry with port", "localhost:5000/vervstack/velez:latest", true},
		{"digest", "ghcr.io/vervstack/velez@sha256:abc", true},
		{"uppercase", "Vervstack/Velez:1", true},
		{"unrelated image", "postgres:16", false},
		{"velez only in the tag", "postgres:velez", false},
		{"empty", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, IsVelezImage(tc.image))
		})
	}
}
