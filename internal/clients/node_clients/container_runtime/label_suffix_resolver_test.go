package container_runtime

import (
	"testing"

	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	testLabelSuffixResolverSuffix     = "stage"
	testLabelSuffixResolverIdentifier = "mysvc"
	testLabelSuffixResolverNetwork    = "vervnet"
)

func TestLabelSuffixResolver_ContainerName(t *testing.T) {
	tests := []struct {
		name     string
		suffix   string
		input    string
		expected string
	}{
		{
			name:     "empty suffix leaves name unchanged",
			suffix:   "",
			input:    testLabelSuffixResolverIdentifier,
			expected: testLabelSuffixResolverIdentifier,
		},
		{
			name:     "non-empty suffix appends _suffix",
			suffix:   testLabelSuffixResolverSuffix,
			input:    testLabelSuffixResolverIdentifier,
			expected: testLabelSuffixResolverIdentifier + "_" + testLabelSuffixResolverSuffix,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &labelSuffixResolver{suffix: tt.suffix}

			got := resolver.ContainerName(tt.input)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestLabelSuffixResolver_NetworkName(t *testing.T) {
	tests := []struct {
		name     string
		suffix   string
		input    string
		expected string
	}{
		{
			name:     "empty suffix leaves name unchanged",
			suffix:   "",
			input:    testLabelSuffixResolverNetwork,
			expected: testLabelSuffixResolverNetwork,
		},
		{
			name:     "non-empty suffix appends _suffix",
			suffix:   testLabelSuffixResolverSuffix,
			input:    testLabelSuffixResolverNetwork,
			expected: testLabelSuffixResolverNetwork + "_" + testLabelSuffixResolverSuffix,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &labelSuffixResolver{suffix: tt.suffix}

			require.Equal(t, resolver.ContainerName(tt.input), resolver.NetworkName(tt.input),
				"NetworkName must delegate to ContainerName")
			require.Equal(t, tt.expected, resolver.NetworkName(tt.input))
		})
	}
}

func TestLabelSuffixResolver_VirtualContainerName(t *testing.T) {
	tests := []struct {
		name   string
		suffix string
		input  string
	}{
		{name: "empty suffix round-trips", suffix: "", input: testLabelSuffixResolverIdentifier},
		{
			name:   "non-empty suffix round-trips",
			suffix: testLabelSuffixResolverSuffix,
			input:  testLabelSuffixResolverIdentifier,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &labelSuffixResolver{suffix: tt.suffix}

			dockerName := resolver.ContainerName(tt.input)
			virtual := resolver.VirtualContainerName(dockerName)

			require.Equal(t, tt.input, virtual)
		})
	}
}

func TestLabelSuffixResolver_CandidateNames(t *testing.T) {
	tests := []struct {
		name     string
		suffix   string
		input    string
		expected []string
	}{
		{
			name:     "empty suffix returns single-element identifier",
			suffix:   "",
			input:    testLabelSuffixResolverIdentifier,
			expected: []string{testLabelSuffixResolverIdentifier},
		},
		{
			name:   "non-empty suffix returns suffixed first, then identifier",
			suffix: testLabelSuffixResolverSuffix,
			input:  testLabelSuffixResolverIdentifier,
			expected: []string{
				testLabelSuffixResolverIdentifier + "_" + testLabelSuffixResolverSuffix,
				testLabelSuffixResolverIdentifier,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &labelSuffixResolver{suffix: tt.suffix}

			got := resolver.CandidateNames(tt.input)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestLabelSuffixResolver_Owns(t *testing.T) {
	tests := []struct {
		name            string
		suffix          string
		containerLabels map[string]string
		expected        bool
	}{
		{
			name:            "empty suffix matches empty label",
			suffix:          "",
			containerLabels: map[string]string{labels.SuffixLabel: ""},
			expected:        true,
		},
		{
			name:            "matching non-empty suffix",
			suffix:          testLabelSuffixResolverSuffix,
			containerLabels: map[string]string{labels.SuffixLabel: testLabelSuffixResolverSuffix},
			expected:        true,
		},
		{
			name:            "non-matching suffix",
			suffix:          testLabelSuffixResolverSuffix,
			containerLabels: map[string]string{labels.SuffixLabel: "other"},
			expected:        false,
		},
		{
			name:            "nil containerLabels with empty suffix must not panic and must match",
			suffix:          "",
			containerLabels: nil,
			expected:        true,
		},
		{
			name:            "nil containerLabels with non-empty suffix must not panic and must not match",
			suffix:          testLabelSuffixResolverSuffix,
			containerLabels: nil,
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &labelSuffixResolver{suffix: tt.suffix}

			require.NotPanics(t, func() {
				got := resolver.Owns(tt.containerLabels)
				require.Equal(t, tt.expected, got)
			})
		})
	}
}

func TestLabelSuffixResolver_StampLabels(t *testing.T) {
	tests := []struct {
		name   string
		suffix string
	}{
		{name: "empty suffix is still stamped", suffix: ""},
		{name: "non-empty suffix is stamped", suffix: testLabelSuffixResolverSuffix},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &labelSuffixResolver{suffix: tt.suffix}

			containerLabels := map[string]string{}

			resolver.StampLabels(containerLabels)

			got, ok := containerLabels[labels.SuffixLabel]
			require.True(t, ok, "the suffix label key must always be present")
			require.Equal(t, tt.suffix, got)
		})
	}
}

func TestLabelSuffixResolver_ListFilterLabels(t *testing.T) {
	tests := []struct {
		name   string
		suffix string
	}{
		{name: "empty suffix is still set", suffix: ""},
		{name: "non-empty suffix is set", suffix: testLabelSuffixResolverSuffix},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &labelSuffixResolver{suffix: tt.suffix}

			filterLabels := map[string]string{"preexisting": "kept"}

			resolver.ListFilterLabels(filterLabels)

			got, ok := filterLabels[labels.SuffixLabel]
			require.True(t, ok, "the suffix label filter key must always be present")
			require.Equal(t, tt.suffix, got)
			require.Equal(t, "kept", filterLabels["preexisting"], "pre-existing keys must be preserved")
		})
	}
}
