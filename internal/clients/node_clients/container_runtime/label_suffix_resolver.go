package container_runtime

import (
	"strings"

	"go.vervstack.ru/Velez/internal/domain/labels"
)

const (
	// nameSuffixSeparator joins a smerd's logical name and its environment
	// suffix. It reproduces, byte for byte, the convention the
	// pre-environments pipeliner used (the deleted
	// internal/pipelines/do_smerd_launch.go's
	// `req.Name = req.GetName() + "_" + p.suffix`).
	nameSuffixSeparator = "_"
)

// labelSuffixResolver is tier 1's nameResolver, the default: every
// environment on the node shares one Docker daemon and is kept apart by
//
//   - the labels.SuffixLabel stamped on each container (what ListSmerds and
//     friends filter on), and
//   - the actual Docker container NAME, which carries the same suffix so two
//     environments deploying the same logical smerd name don't collide on the
//     daemon's globally-unique container namespace.
//
// An empty suffix - today's default/PROD on a node that never configured
// ContainerSuffix - means "unsuffixed", preserving pre-multi-environment
// container naming exactly.
type labelSuffixResolver struct {
	suffix string
}

// ContainerName resolves the logical smerd name into the actual Docker
// container name for this environment: the bare name when the suffix is empty,
// "<name>_<suffix>" otherwise.
func (n *labelSuffixResolver) ContainerName(name string) string {
	if n.suffix == "" {
		return name
	}

	return name + nameSuffixSeparator + n.suffix
}

// NetworkName resolves a logical network name into the actual Docker network
// name for this environment - byte-for-byte the same suffixing rule
// ContainerName applies to container names (see nameSuffixSeparator): the
// bare name when the suffix is empty, "<name>_<suffix>" otherwise. Kept as
// its own named method (rather than callers using ContainerName directly)
// so a future divergence between container- and network-naming rules doesn't
// require re-auditing every call site.
func (n *labelSuffixResolver) NetworkName(name string) string {
	return n.ContainerName(name)
}

// VirtualContainerName reverses ContainerName: strips this resolver's suffix
// from a real Docker name, recovering the virtual/logical name
// ContainerCreate was called with. Delegates to the package-level
// StripEnvironmentSuffix so the "<name>_<suffix>" convention is defined in
// exactly one place - see that function's doc comment for why it's exported
// even though this is its only caller today.
func (n *labelSuffixResolver) VirtualContainerName(dockerName string) string {
	return StripEnvironmentSuffix(dockerName, n.suffix)
}

// CandidateNames returns the identifier forms resolveOwnedContainer should
// try, in order: the suffixed logical name first, then identifier itself.
// When n.suffix is empty, ContainerName(identifier) already equals
// identifier, so the second attempt would be redundant - it's omitted.
func (n *labelSuffixResolver) CandidateNames(identifier string) []string {
	suffixed := n.ContainerName(identifier)
	if suffixed == identifier {
		return []string{identifier}
	}

	return []string{suffixed, identifier}
}

// Owns reports whether containerLabels belongs to this resolver's
// environment - an exact match of labels.SuffixLabel against n.suffix,
// including when n.suffix is "" (empty is a real value to match, not a
// wildcard). A nil containerLabels map reads as the zero value ("") for any
// key, so this is naturally correct for a not-yet-labeled/missing-config
// container without any extra nil-guard.
func (n *labelSuffixResolver) Owns(containerLabels map[string]string) bool {
	return containerLabels[labels.SuffixLabel] == n.suffix
}

// StampLabels stamps this resolver's suffix onto containerLabels
// unconditionally - even an empty suffix is a real, meaningful value being
// stamped ("this environment's containers, whose suffix happens to be
// empty"), not "skip stamping."
func (n *labelSuffixResolver) StampLabels(containerLabels map[string]string) {
	containerLabels[labels.SuffixLabel] = n.suffix
}

// ListFilterLabels sets this resolver's suffix as a list filter
// unconditionally - same "empty is a real value, not a wildcard" rule as
// StampLabels. See docs/container_runtimes/interface_design.md.
func (n *labelSuffixResolver) ListFilterLabels(filterLabels map[string]string) {
	filterLabels[labels.SuffixLabel] = n.suffix
}

// StripEnvironmentSuffix reverses the "<name>_<suffix>" naming convention
// labelSuffixResolver.ContainerName applies (byte for byte - see
// nameSuffixSeparator): given a real Docker name and the suffix it was
// created with, it returns the virtual/logical name. An empty suffix or a
// name that doesn't end in "_<suffix>" is returned unchanged.
//
// Exported (rather than folded into VirtualContainerName) because,
// historically, not every reader of a Docker container name went through a
// resolved ContainerRuntime - container_manager.InspectSmerd used to call
// client.APIClient.ContainerInspect directly and recover the suffix from the
// container's own labels.SuffixLabel. InspectSmerd itself has since been
// rewired onto ContainerRuntime.Inspect (which calls this via
// VirtualContainerName), so VirtualContainerName is this function's only
// caller today - kept exported since other packages/tests may still
// reference it directly (see docs/container_runtimes/roadmap.md).
func StripEnvironmentSuffix(dockerName, suffix string) string {
	if suffix == "" {
		return dockerName
	}

	return strings.TrimSuffix(dockerName, nameSuffixSeparator+suffix)
}
