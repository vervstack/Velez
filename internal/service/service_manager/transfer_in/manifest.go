package transfer_in

import (
	"encoding/json"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

type transferManifest struct {
	containers []manifestContainer
	items      []manifestItem
	totalBytes int64
}

type manifestContainer struct {
	name         string
	networkOwner string
	inspect      container.InspectResponse
}

type manifestItem struct {
	name       string
	kind       velez_api.TransferItemKind
	totalBytes int64
}

func (c manifestContainer) isRoot() bool {
	return c.networkOwner == ""
}

func (m transferManifest) roots() []manifestContainer {
	roots := make([]manifestContainer, 0, len(m.containers))

	for _, c := range m.containers {
		if c.isRoot() {
			roots = append(roots, c)
		}
	}

	return roots
}

func (m transferManifest) sidecars() []manifestContainer {
	sidecars := make([]manifestContainer, 0, len(m.containers))

	for _, c := range m.containers {
		if !c.isRoot() {
			sidecars = append(sidecars, c)
		}
	}

	return sidecars
}

func (m transferManifest) item(name string) (manifestItem, bool) {
	for _, it := range m.items {
		if it.name == name {
			return it, true
		}
	}

	return manifestItem{}, false
}

func parseManifest(begin *velez_api.Transfer_Begin) (transferManifest, error) {
	var manifest transferManifest

	if len(begin.GetContainers()) == 0 {
		return manifest, rerrors.Wrap(errManifestEmpty)
	}

	for _, c := range begin.GetContainers() {
		parsed, err := parseContainer(c)
		if err != nil {
			return transferManifest{}, rerrors.Wrap(err)
		}

		manifest.containers = append(manifest.containers, parsed)
	}

	err := validateContainers(manifest)
	if err != nil {
		return transferManifest{}, rerrors.Wrap(err)
	}

	for _, it := range begin.GetItems() {
		item := manifestItem{name: it.GetName(), kind: it.GetKind(), totalBytes: it.GetTotalBytes()}

		manifest.items = append(manifest.items, item)

		manifest.totalBytes += item.totalBytes
	}

	err = validateItems(manifest.items)
	if err != nil {
		return transferManifest{}, rerrors.Wrap(err)
	}

	return manifest, nil
}

func parseContainer(c *velez_api.Transfer_ContainerManifest) (manifestContainer, error) {
	parsed := manifestContainer{name: c.GetName(), networkOwner: c.GetNetworkOwner()}

	if parsed.name == "" {
		return parsed, rerrors.Wrap(errInvalidInspect, "container name is empty")
	}

	err := json.Unmarshal(c.GetInspectJson(), &parsed.inspect)
	if err != nil {
		return parsed, rerrors.Wrap(errInvalidInspect, parsed.name)
	}

	isComplete := parsed.inspect.ContainerJSONBase != nil &&
		parsed.inspect.HostConfig != nil &&
		parsed.inspect.Config != nil
	if !isComplete {
		return parsed, rerrors.Wrap(errInvalidInspect, parsed.name)
	}

	for _, m := range parsed.inspect.Mounts {
		if m.Type == mount.TypeBind {
			return parsed, rerrors.Wrap(errBindMountUnsupported, parsed.name)
		}
	}

	return parsed, nil
}

func validateContainers(manifest transferManifest) error {
	names := make(map[string]struct{}, len(manifest.containers))

	for _, c := range manifest.containers {
		_, isDuplicate := names[c.name]
		if isDuplicate {
			return rerrors.Wrap(errDuplicateContainer, c.name)
		}

		names[c.name] = struct{}{}
	}

	roots := make(map[string]struct{})

	for _, c := range manifest.roots() {
		roots[c.name] = struct{}{}
	}

	if len(roots) == 0 {
		return rerrors.Wrap(errNoRootContainer)
	}

	for _, c := range manifest.sidecars() {
		_, isRoot := roots[c.networkOwner]
		if !isRoot {
			return rerrors.Wrap(errUnknownNetworkOwner, c.name)
		}
	}

	return nil
}

func validateItems(items []manifestItem) error {
	names := make(map[string]struct{}, len(items))

	for _, it := range items {
		isKindKnown := it.kind == velez_api.TransferItemKind_TRANSFER_ITEM_KIND_IMAGE ||
			it.kind == velez_api.TransferItemKind_TRANSFER_ITEM_KIND_VOLUME
		if it.name == "" || it.totalBytes < 0 || !isKindKnown {
			return rerrors.Wrap(errInvalidItem, it.name)
		}

		_, isDuplicate := names[it.name]
		if isDuplicate {
			return rerrors.Wrap(errDuplicateItem, it.name)
		}

		names[it.name] = struct{}{}
	}

	return nil
}
