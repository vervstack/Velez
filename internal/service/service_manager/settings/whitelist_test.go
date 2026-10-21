package settings

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"go.vervstack.ru/Velez/internal/domain"
)

const (
	dockerRepository = "docker"
)

func Test_BuildWhitelistEntries_BuiltinsFirst(t *testing.T) {
	storedUuid := uuid.New()
	stored := []domain.WhitelistedImage{{Id: storedUuid, Repository: "acme/agent"}}

	entries := buildWhitelistEntries([]string{dockerRepository, "moby/buildkit"}, stored)

	require.Len(t, entries, 3)
	require.Equal(t, domain.WhitelistedImageEntry{Repository: dockerRepository, IsBuiltin: true}, entries[0])
	require.Equal(t, domain.WhitelistedImageEntry{Repository: "moby/buildkit", IsBuiltin: true}, entries[1])
	require.Equal(t, domain.WhitelistedImageEntry{Id: storedUuid.String(), Repository: "acme/agent"}, entries[2])
}

func Test_BuildWhitelistEntries_NoStoredKeepsBuiltins(t *testing.T) {
	entries := buildWhitelistEntries([]string{dockerRepository}, nil)

	require.Len(t, entries, 1)
	require.True(t, entries[0].IsBuiltin)
}
