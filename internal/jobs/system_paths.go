package jobs

import (
	"path"
	"strings"
)

const (
	socketSuffix = ".sock"
	runRoot      = "/run"
)

var systemPathRoots = []string{"/dev", "/proc", "/sys", runRoot, "/var/run"}

// isSystemPath reports whether a host path is a kernel or runtime path that a
// named volume cannot stand in for: a device, a pseudo filesystem, a runtime
// directory or a unix socket.
func isSystemPath(p string) bool {
	cleaned := path.Clean(p)

	for _, root := range systemPathRoots {
		if cleaned == root || strings.HasPrefix(cleaned, root+"/") {
			return true
		}
	}

	return strings.HasSuffix(path.Base(cleaned), socketSuffix)
}
