package transfer_out

import (
	"encoding/json"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

// containerManifests lists the root first, then the sidecars sharing its network namespace.
func containerManifests(
	root container.InspectResponse,
	sidecars []container.InspectResponse,
) ([]*velez_api.Transfer_ContainerManifest, error) {
	manifests := make([]*velez_api.Transfer_ContainerManifest, 0, 1+len(sidecars))

	rootManifest, err := toContainerManifest(root, "")
	if err != nil {
		return nil, rerrors.Wrap(err)
	}

	manifests = append(manifests, rootManifest)

	for _, sidecar := range sidecars {
		sidecarManifest, sidecarErr := toContainerManifest(sidecar, containerName(root))
		if sidecarErr != nil {
			return nil, rerrors.Wrap(sidecarErr)
		}

		manifests = append(manifests, sidecarManifest)
	}

	return manifests, nil
}

func toContainerManifest(
	inspected container.InspectResponse,
	networkOwner string,
) (*velez_api.Transfer_ContainerManifest, error) {
	inspectJson, err := json.Marshal(inspected)
	if err != nil {
		return nil, rerrors.Wrap(err, "error marshalling container inspect")
	}

	manifest := &velez_api.Transfer_ContainerManifest{
		Name:         containerName(inspected),
		NetworkOwner: networkOwner,
		InspectJson:  inspectJson,
	}

	return manifest, nil
}

// distinctImages is the set of image references of the containers, in first-seen order.
func distinctImages(containers []container.InspectResponse) []string {
	seen := make(map[string]struct{})
	images := make([]string, 0, len(containers))

	for _, inspected := range containers {
		ref := imageRef(inspected)

		_, isSeen := seen[ref]
		if ref == "" || isSeen {
			continue
		}

		seen[ref] = struct{}{}
		images = append(images, ref)
	}

	return images
}

// distinctVolumes is the set of named volumes mounted by the containers, in first-seen order.
func distinctVolumes(containers []container.InspectResponse) []string {
	seen := make(map[string]struct{})
	volumes := make([]string, 0)

	for _, inspected := range containers {
		for _, mnt := range inspected.Mounts {
			_, isSeen := seen[mnt.Name]
			if mnt.Type != mount.TypeVolume || mnt.Name == "" || isSeen {
				continue
			}

			seen[mnt.Name] = struct{}{}
			volumes = append(volumes, mnt.Name)
		}
	}

	return volumes
}
