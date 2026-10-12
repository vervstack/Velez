package verv_services

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testHost     = "a.local"
	testDindName = "arteldind"
)

func Test_WithoutDindHost_Scenarios(t *testing.T) {
	cases := []struct {
		name     string
		hosts    []string
		dindName string
		want     []string
	}{
		{"dind host is removed", []string{testHost, testDindName}, testDindName, []string{testHost}},
		{"empty dind name keeps everything", []string{testHost, testDindName}, "", []string{testHost, testDindName}},
		{"absent dind host keeps everything", []string{testHost}, testDindName, []string{testHost}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, withoutDindHost(tc.hosts, tc.dindName))
		})
	}
}
