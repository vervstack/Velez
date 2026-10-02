package version

const (
	devVersion = "dev"
)

// version is set at release build time: -ldflags "-X go.vervstack.ru/Velez/internal/version.version=<tag>".
//
//nolint:gochecknoglobals // link-time injected release tag, cannot be a constant
var version string

// Get returns the release tag baked in at build time, or "dev" for any other build.
func Get() string {
	if version == "" {
		return devVersion
	}

	return version
}
