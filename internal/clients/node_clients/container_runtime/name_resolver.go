package container_runtime

// nameResolver is the pluggable seam between dockerRuntime's generic,
// backend-agnostic Docker-calling logic and the environment-scoping policy
// that differs between backends.
type nameResolver interface {
	ContainerName(name string) string
	NetworkName(name string) string
	VirtualContainerName(dockerName string) string
	CandidateNames(identifier string) []string
	Owns(containerLabels map[string]string) bool
	StampLabels(containerLabels map[string]string)
	ListFilterLabels(filterLabels map[string]string)
}
