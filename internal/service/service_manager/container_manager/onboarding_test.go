package container_manager

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
	"go.vervstack.ru/Velez/internal/domain/labels"
)

func Test_replacedContainerIds_NoLabels(t *testing.T) {
	list := []container.Summary{
		{ID: "cont-1", Labels: map[string]string{}},
	}

	got := replacedContainerIds(list)

	require.Empty(t, got)
}

func Test_replacedContainerIds_OneOnboarded(t *testing.T) {
	oldId := "old-container-id"
	newId := "new-container-id"
	list := []container.Summary{
		{
			ID: newId,
			Labels: map[string]string{
				labels.OnboardedFromLabel: oldId,
			},
		},
	}

	got := replacedContainerIds(list)

	require.Equal(t, map[string]string{oldId: newId}, got)
}

func Test_replacedContainerIds_TwoContainers(t *testing.T) {
	oldId1 := "old-1"
	newId1 := "new-1"
	oldId2 := "old-2"
	newId2 := "new-2"
	list := []container.Summary{
		{
			ID: newId1,
			Labels: map[string]string{
				labels.OnboardedFromLabel: oldId1,
			},
		},
		{
			ID: newId2,
			Labels: map[string]string{
				labels.OnboardedFromLabel: oldId2,
			},
		},
	}

	got := replacedContainerIds(list)

	require.Equal(t, map[string]string{oldId1: newId1, oldId2: newId2}, got)
}
