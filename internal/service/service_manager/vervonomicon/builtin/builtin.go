// Package builtin embeds the .verv/ descriptors Velez ships for the
// containers it provisions on the user's behalf (postgres, registry) - see
// docs/features/pgaas_and_registry_plugin.md section 2. Read returns the same
// flat map keyed by path-relative-to-.verv/ that
// vervonomicon.ImageSource.Read returns, so
// internal/service/service_manager/vervonomicon.Parse and MergeEnvironment
// consume it unchanged.
package builtin

import (
	"embed"
	"io/fs"
	"strings"

	"go.redsock.ru/rerrors"
)

//go:embed postgres registry
var descriptors embed.FS

// Read returns the builtin descriptor files for name ("postgres" or
// "registry"), as a flat map keyed by path relative to .verv/ - e.g.
// "vervonomicon.yaml", "deployment.yaml".
func Read(name string) (map[string][]byte, error) {
	out := make(map[string][]byte)

	walkErr := fs.WalkDir(descriptors, name, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		content, readErr := fs.ReadFile(descriptors, path)
		if readErr != nil {
			return rerrors.Wrap(readErr, "error reading builtin descriptor file "+path)
		}

		relPath := strings.TrimPrefix(path, name+"/")

		out[relPath] = content

		return nil
	})
	if walkErr != nil {
		return nil, rerrors.Wrap(walkErr, "error reading builtin descriptor "+name)
	}

	return out, nil
}
