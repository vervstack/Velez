package settings

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
)

func Test_MergeSettings_OnlyNonNilFieldsApplied(t *testing.T) {
	isTrue := true
	isFalse := false

	cases := []struct {
		name    string
		current domain.Settings
		req     domain.UpdateSettingsReq
		want    domain.Settings
	}{
		{
			"empty request keeps current",
			domain.Settings{IsSysboxEnabled: true},
			domain.UpdateSettingsReq{},
			domain.Settings{IsSysboxEnabled: true},
		},
		{
			"sysbox only",
			domain.Settings{IsSysboxWhitelistIgnored: true},
			domain.UpdateSettingsReq{IsSysboxEnabled: &isTrue},
			domain.Settings{IsSysboxEnabled: true, IsSysboxWhitelistIgnored: true},
		},
		{
			"both fields",
			domain.Settings{IsSysboxEnabled: true},
			domain.UpdateSettingsReq{IsSysboxEnabled: &isFalse, IsSysboxWhitelistIgnored: &isTrue},
			domain.Settings{IsSysboxWhitelistIgnored: true},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, mergeSettings(tc.current, tc.req))
		})
	}
}

func Test_IsRootless_DetectsRootlessSecurityOption(t *testing.T) {
	cases := []struct {
		name    string
		options []string
		want    bool
	}{
		{"nil options", nil, false},
		{"default options", []string{"name=seccomp,profile=builtin", "name=cgroupns"}, false},
		{"rootless option", []string{"name=seccomp,profile=builtin", "name=rootless"}, true},
		{"rootless with cgroupns", []string{"name=rootless", "name=cgroupns"}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, isRootless(tc.options))
		})
	}
}

func Test_IsSnapRoot_DetectsSnapPath(t *testing.T) {
	cases := []struct {
		name string
		root string
		want bool
	}{
		{"empty path", "", false},
		{"default root", "/var/lib/docker", false},
		{"snap root", "/var/snap/docker/common/var-lib-docker", true},
		{"snap at root", "/snap/docker/current", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, isSnapRoot(tc.root))
		})
	}
}
