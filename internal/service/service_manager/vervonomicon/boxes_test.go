package vervonomicon

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.redsock.ru/rerrors"

	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
)

const (
	boxSmall  = "small"
	boxMedium = "medium"
)

var errBoxNotFound = rerrors.New("box not found")

type fakeBoxLookup struct {
	boxes map[string]verv.Box
}

func newFakeBoxLookup() *fakeBoxLookup {
	return &fakeBoxLookup{
		boxes: map[string]verv.Box{
			boxSmall:  {Name: boxSmall, Cpu: 0.5, RamMb: 512, DiskMb: 2048, IsBuiltin: true},
			boxMedium: {Name: boxMedium, Cpu: 1.0, RamMb: 1024, DiskMb: 8192, IsBuiltin: true},
		},
	}
}

func (f *fakeBoxLookup) GetBox(_ context.Context, name string) (verv.Box, error) {
	box, ok := f.boxes[name]
	if !ok {
		return verv.Box{}, errBoxNotFound
	}

	return box, nil
}

func (f *fakeBoxLookup) ListBoxes(_ context.Context) ([]verv.Box, error) {
	result := make([]verv.Box, 0, len(f.boxes))
	for _, box := range f.boxes {
		result = append(result, box)
	}

	return result, nil
}

func TestBoxResolver_ResolveApp_Precedence(t *testing.T) {
	resolver := NewBoxResolver(newFakeBoxLookup())

	cases := []struct {
		name    string
		app     verv.App
		rootBox string
		want    verv.Sizing
	}{
		{
			name:    "exact resources block wins over everything",
			app:     verv.App{Box: boxSmall, Resources: verv.Sizing{Cpu: 4, RamMb: 4096}},
			rootBox: boxMedium,
			want:    verv.Sizing{Cpu: 4, RamMb: 4096},
		},
		{
			name:    "own box wins over root box",
			app:     verv.App{Box: boxSmall},
			rootBox: boxMedium,
			want:    verv.Sizing{Cpu: 0.5, RamMb: 512},
		},
		{
			name:    "root box used when entry has none",
			app:     verv.App{},
			rootBox: boxMedium,
			want:    verv.Sizing{Cpu: 1.0, RamMb: 1024},
		},
		{
			name:    "no limits when nothing set anywhere",
			app:     verv.App{},
			rootBox: "",
			want:    verv.Sizing{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolver.ResolveApp(context.Background(), tc.app, tc.rootBox)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestBoxResolver_ResolveResource_ExactResourcesWinsOverBox(t *testing.T) {
	resolver := NewBoxResolver(newFakeBoxLookup())

	res := verv.Resource{
		Box: boxSmall,
		Resources: verv.ResourceSizing{
			Sizing: verv.Sizing{Cpu: 2, RamMb: 2048},
			DiskMb: 16384,
		},
	}

	got, err := resolver.ResolveResource(context.Background(), res, boxMedium)
	require.NoError(t, err)
	require.Equal(t, res.Resources, got)
}

func TestBoxResolver_ResolveResource_FallsBackToBoxAndAddsDisk(t *testing.T) {
	resolver := NewBoxResolver(newFakeBoxLookup())

	res := verv.Resource{Box: boxMedium}

	got, err := resolver.ResolveResource(context.Background(), res, boxSmall)
	require.NoError(t, err)
	require.Equal(t, verv.ResourceSizing{
		Sizing: verv.Sizing{Cpu: 1.0, RamMb: 1024},
		DiskMb: 8192,
	}, got)
}

func TestBoxResolver_UnknownBoxNamesItAndListsKnownOnes(t *testing.T) {
	resolver := NewBoxResolver(newFakeBoxLookup())

	_, err := resolver.ResolveApp(context.Background(), verv.App{Box: "gigantic"}, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "gigantic")
	require.Contains(t, err.Error(), boxSmall)
	require.Contains(t, err.Error(), boxMedium)
}
