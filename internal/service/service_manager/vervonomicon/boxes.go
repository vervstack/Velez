package vervonomicon

import (
	"context"
	"strings"

	"go.redsock.ru/rerrors"

	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
)

// BoxLookup is the narrow storage dependency box resolution needs: reading
// sizing tiers out of velez.resource_boxes. Declared here, where it's
// consumed, rather than next to the storage implementation -
// internal/storage/postgres/resource_boxes.go implements it.
type BoxLookup interface {
	GetBox(ctx context.Context, name string) (verv.Box, error)
	ListBoxes(ctx context.Context) ([]verv.Box, error)
}

// BoxResolver resolves a `box` name or an exact `resources` block into
// concrete cpu/ram/disk, per docs/features/vervonomicon.md's "Boxes" section.
type BoxResolver struct {
	boxes BoxLookup
}

func NewBoxResolver(boxes BoxLookup) *BoxResolver {
	return &BoxResolver{
		boxes: boxes,
	}
}

// ResolveApp resolves deployment.yaml's app.box / app.resources against
// rootBox (vervonomicon.yaml's root-level box), in the precedence order the
// spec fixes, most specific first: an exact resources block -> the app's own
// box -> the root-level box -> no limits.
func (r *BoxResolver) ResolveApp(ctx context.Context, app verv.App, rootBox string) (verv.Sizing, error) {
	if hasExactSizing(app.Resources) {
		return app.Resources, nil
	}

	boxName := app.Box
	if boxName == "" {
		boxName = rootBox
	}

	if boxName == "" {
		return verv.Sizing{}, nil
	}

	box, err := r.resolveBox(ctx, boxName)
	if err != nil {
		return verv.Sizing{}, err
	}

	sizing := verv.Sizing{
		Cpu:   box.Cpu,
		RamMb: box.RamMb,
	}

	return sizing, nil
}

// ResolveResource resolves one resources.yaml entry's own box/resources
// against rootBox, same precedence as ResolveApp, but returns ResourceSizing
// (adds the resource's own disk allocation).
func (r *BoxResolver) ResolveResource(
	ctx context.Context, res verv.Resource, rootBox string,
) (verv.ResourceSizing, error) {
	if hasExactSizing(res.Resources.Sizing) {
		return res.Resources, nil
	}

	boxName := res.Box
	if boxName == "" {
		boxName = rootBox
	}

	if boxName == "" {
		return verv.ResourceSizing{}, nil
	}

	box, err := r.resolveBox(ctx, boxName)
	if err != nil {
		return verv.ResourceSizing{}, err
	}

	sizing := verv.ResourceSizing{
		Sizing: verv.Sizing{
			Cpu:   box.Cpu,
			RamMb: box.RamMb,
		},
		DiskMb: box.DiskMb,
	}

	return sizing, nil
}

func (r *BoxResolver) resolveBox(ctx context.Context, name string) (verv.Box, error) {
	box, err := r.boxes.GetBox(ctx, name)
	if err == nil {
		return box, nil
	}

	return verv.Box{}, r.unknownBoxError(ctx, name, err)
}

// unknownBoxError names the box that wasn't found and lists the known ones,
// per the spec: "An unknown box name is an error naming the box and listing
// the known ones."
func (r *BoxResolver) unknownBoxError(ctx context.Context, name string, cause error) error {
	boxes, listErr := r.boxes.ListBoxes(ctx)
	if listErr != nil {
		return rerrors.Wrap(cause, "unknown box '"+name+"'")
	}

	knownNames := make([]string, len(boxes))
	for i, box := range boxes {
		knownNames[i] = box.Name
	}

	msg := "unknown box '" + name + "', known boxes: " + strings.Join(knownNames, ", ")

	return rerrors.Wrap(cause, msg)
}

// hasExactSizing reports whether a Sizing block was actually set - any
// nonzero field wins over a box, per the spec's precedence rule.
func hasExactSizing(s verv.Sizing) bool {
	return s.Cpu != 0 || s.RamMb != 0 || s.MemorySwapMb != 0
}
