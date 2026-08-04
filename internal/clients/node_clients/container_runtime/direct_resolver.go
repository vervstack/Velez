package container_runtime

// directResolver is a true no-op nameResolver: for a backend where each
// environment already has its own dedicated Docker daemon (tier 2), there is
// no shared-daemon namespace to disambiguate, so every name/label decision is
// the identity/always-true/no-op case.
type directResolver struct{}

// ContainerName is the identity - no suffixing needed on a dedicated daemon.
func (d *directResolver) ContainerName(name string) string {
	return name
}

// NetworkName is the identity - no suffixing needed on a dedicated daemon.
func (d *directResolver) NetworkName(name string) string {
	return name
}

// VirtualContainerName is the identity - there is no suffix to strip.
func (d *directResolver) VirtualContainerName(dockerName string) string {
	return dockerName
}

// CandidateNames has exactly one candidate form: identity naming means the
// suffixed and bare forms are always the same.
func (d *directResolver) CandidateNames(identifier string) []string {
	return []string{identifier}
}

// Owns always reports true: a dedicated daemon has no shared-daemon
// disambiguation problem to solve, so every container found belongs to this
// environment.
func (d *directResolver) Owns(_ map[string]string) bool {
	return true
}

// StampLabels is a no-op: labels.CreatedWithVelezLabel is already stamped
// unconditionally by dockerRuntime.ContainerCreate itself before calling
// this, regardless of resolver, so nothing is lost.
func (d *directResolver) StampLabels(_ map[string]string) {
}

// ListFilterLabels is a no-op - a dedicated daemon needs no suffix filter.
func (d *directResolver) ListFilterLabels(_ map[string]string) {
}
