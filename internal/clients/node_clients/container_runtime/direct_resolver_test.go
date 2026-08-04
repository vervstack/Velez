package container_runtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testDirectResolverIdentifier = "mysvc"
)

func TestDirectResolver_ContainerName(t *testing.T) {
	resolver := &directResolver{}

	require.Equal(t, testDirectResolverIdentifier, resolver.ContainerName(testDirectResolverIdentifier))
	require.Equal(t, "arbitrary-name", resolver.ContainerName("arbitrary-name"))
}

func TestDirectResolver_NetworkName(t *testing.T) {
	resolver := &directResolver{}

	require.Equal(t, testDirectResolverIdentifier, resolver.NetworkName(testDirectResolverIdentifier))
	require.Equal(t, "arbitrary-net", resolver.NetworkName("arbitrary-net"))
}

func TestDirectResolver_VirtualContainerName(t *testing.T) {
	resolver := &directResolver{}

	require.Equal(t, testDirectResolverIdentifier, resolver.VirtualContainerName(testDirectResolverIdentifier))
	require.Equal(t, "arbitrary-docker-name", resolver.VirtualContainerName("arbitrary-docker-name"))
}

func TestDirectResolver_CandidateNames(t *testing.T) {
	resolver := &directResolver{}

	got := resolver.CandidateNames(testDirectResolverIdentifier)
	require.Equal(t, []string{testDirectResolverIdentifier}, got)
}

func TestDirectResolver_Owns(t *testing.T) {
	resolver := &directResolver{}

	tests := []struct {
		name            string
		containerLabels map[string]string
	}{
		{name: "nil labels", containerLabels: nil},
		{name: "empty labels", containerLabels: map[string]string{}},
		{name: "arbitrary labels", containerLabels: map[string]string{"suffix": "other-env"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.True(t, resolver.Owns(tt.containerLabels))
		})
	}
}

func TestDirectResolver_StampLabels(t *testing.T) {
	resolver := &directResolver{}

	containerLabels := map[string]string{"a": "1", "b": "2"}
	expected := map[string]string{"a": "1", "b": "2"}

	resolver.StampLabels(containerLabels)

	require.Equal(t, expected, containerLabels, "StampLabels must be a true no-op")
}

func TestDirectResolver_ListFilterLabels(t *testing.T) {
	resolver := &directResolver{}

	filterLabels := map[string]string{"a": "1", "b": "2"}
	expected := map[string]string{"a": "1", "b": "2"}

	resolver.ListFilterLabels(filterLabels)

	require.Equal(t, expected, filterLabels, "ListFilterLabels must be a true no-op")
}
