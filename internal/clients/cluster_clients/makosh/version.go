package makosh

import (
	"runtime/debug"

	makoshConfig "go.vervstack.ru/makosh/config"
)

const (
	modulePath = "go.vervstack.ru/makosh"
)

// ModuleVersion is the Makosh release this binary was built against, i.e. the
// version pinned in go.mod. Makosh's embedded config version is not bumped on
// release (the version lives in git tags only), so it can't name the image.
func ModuleVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return makoshConfig.GetVersion()
	}

	for _, dep := range info.Deps {
		if dep.Path != modulePath {
			continue
		}

		if dep.Replace != nil || dep.Version == "" {
			break
		}

		return dep.Version
	}

	return makoshConfig.GetVersion()
}
