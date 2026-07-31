package state

import (
	"testing"
)

func TestPgName_EmptySuffix_ReturnsBaseName(t *testing.T) {
	got := PgName("")

	want := "verv-cluster-state"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestPgName_WithSuffix_AppendsSuffix(t *testing.T) {
	got := PgName("e2e")

	want := "verv-cluster-state-e2e"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}
