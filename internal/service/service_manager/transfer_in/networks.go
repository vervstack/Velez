package transfer_in

import (
	"context"
	"slices"

	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/clients/node_clients/container_runtime"
)

// ensureNetworks creates the networks the manifest's roots are attached to
// that this node does not have yet.
func ensureNetworks(
	ctx context.Context,
	runtime container_runtime.ContainerRuntime,
	manifest transferManifest,
) error {
	existing, err := runtime.ListNetworks(ctx, true)
	if err != nil {
		return rerrors.Wrap(err, "error listing networks")
	}

	known := make(map[string]struct{}, len(existing))

	for _, n := range existing {
		known[n.Name] = struct{}{}
		known[n.DockerName] = struct{}{}
	}

	for _, name := range requiredNetworks(manifest) {
		_, isKnown := known[name]
		if isKnown {
			continue
		}

		err = runtime.CreateNetwork(ctx, name)
		if err != nil {
			return rerrors.Wrap(err, "error creating network")
		}
	}

	return nil
}

func requiredNetworks(manifest transferManifest) []string {
	roots := manifest.roots()
	names := make([]string, 0, len(roots))

	for _, c := range roots {
		names = append(names, networkNames(c.inspect)...)
	}

	slices.Sort(names)

	return slices.Compact(names)
}
