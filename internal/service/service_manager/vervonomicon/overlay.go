package vervonomicon

import (
	"maps"
	"path"
	"strings"

	"go.redsock.ru/rerrors"
	"gopkg.in/yaml.v3"

	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
)

const (
	// nameKey is the field lists of objects merge by, per the spec's
	// "Environment overlays" section.
	nameKey = "name"
)

// MergeEnvironment applies the <environment>/ overlay over the base files
// under .verv/ and parses the result, per docs/features/vervonomicon.md's
// "Environment overlays" section:
//
//   - files under <environment>/ are the overlay, files at the root are the
//     base; every other <other-env>/ directory is dropped entirely
//   - YAML files deep-merge (mergeYamlValue); any other file (*.conf and
//     anything else) is replaced wholesale, never merged
//
// files is keyed exactly like Parse's input - path relative to .verv/. An
// empty environment means no overlay is applied at all.
func MergeEnvironment(files map[string][]byte, environment string) (verv.Descriptor, error) {
	merged, err := mergeFiles(files, environment)
	if err != nil {
		return verv.Descriptor{}, err
	}

	descriptor, err := Parse(merged)
	if err != nil {
		return verv.Descriptor{}, err
	}

	descriptor.Environment = environment

	return descriptor, nil
}

// mergeFiles splits files into base and <environment>/ overlay, drops every
// other top-level directory, and merges each overlaid file individually.
func mergeFiles(files map[string][]byte, environment string) (map[string][]byte, error) {
	base := make(map[string][]byte)
	overlay := make(map[string][]byte)

	for name, content := range files {
		dir, rel := splitTopDir(name)
		if dir == "" {
			base[name] = content

			continue
		}

		if dir == environment {
			overlay[rel] = content
		}

		// A different environment's overlay directory - dropped entirely.
	}

	merged := make(map[string][]byte, len(base)+len(overlay))
	maps.Copy(merged, base)

	for name, overlayContent := range overlay {
		baseContent, hasBase := base[name]
		if !hasBase {
			merged[name] = overlayContent

			continue
		}

		mergedContent, err := mergeFileContent(name, baseContent, overlayContent)
		if err != nil {
			return nil, err
		}

		merged[name] = mergedContent
	}

	return merged, nil
}

// splitTopDir splits a .verv/-relative path into its top-level directory (if
// any) and the remainder. A root file ("ingress.yaml") returns ("", name).
func splitTopDir(name string) (dir, rest string) {
	dir, rest, found := strings.Cut(name, "/")
	if !found {
		return "", name
	}

	return dir, rest
}

// mergeFileContent merges one overlaid file. Non-YAML files (*.conf and
// anything else) are replaced wholesale, verbatim - never merged line by
// line, and never round-tripped through a YAML encoder.
func mergeFileContent(name string, base, overlay []byte) ([]byte, error) {
	if !isYamlFile(name) {
		return overlay, nil
	}

	var baseVal any

	err := yaml.Unmarshal(base, &baseVal)
	if err != nil {
		return nil, rerrors.Wrap(err, "error parsing base yaml for "+name)
	}

	var overlayVal any

	err = yaml.Unmarshal(overlay, &overlayVal)
	if err != nil {
		return nil, rerrors.Wrap(err, "error parsing overlay yaml for "+name)
	}

	mergedVal := mergeYamlValue(baseVal, overlayVal)

	mergedBytes, err := yaml.Marshal(mergedVal)
	if err != nil {
		return nil, rerrors.Wrap(err, "error marshalling merged yaml for "+name)
	}

	return mergedBytes, nil
}

func isYamlFile(name string) bool {
	ext := path.Ext(name)

	return ext == ".yaml" || ext == ".yml"
}

// mergeYamlValue implements the merge semantics fixed by the spec, applied at
// the decoded map[string]any / []any level rather than on typed structs: a
// key the overlay never mentions is simply absent from its decoded map, so a
// `false` or `0` the overlay never set can never be confused with one it
// deliberately set - the ambiguity typed zero-values would create doesn't
// arise here.
//
//   - two maps deep-merge by key
//   - two lists whose every element is an object carrying a "name" field
//     merge by that name; an overlay entry with an unseen name is appended
//   - anything else (scalars, lists of scalars, mismatched types, or lists
//     that aren't uniformly named objects) - the overlay value replaces the
//     base value outright
func mergeYamlValue(base, overlay any) any {
	baseMap, baseIsMap := base.(map[string]any)
	overlayMap, overlayIsMap := overlay.(map[string]any)

	if baseIsMap && overlayIsMap {
		return mergeYamlMaps(baseMap, overlayMap)
	}

	baseList, baseIsList := base.([]any)
	overlayList, overlayIsList := overlay.([]any)

	if baseIsList && overlayIsList {
		return mergeYamlLists(baseList, overlayList)
	}

	return overlay
}

func mergeYamlMaps(base, overlay map[string]any) map[string]any {
	merged := make(map[string]any, len(base)+len(overlay))
	maps.Copy(merged, base)

	for k, overlayVal := range overlay {
		baseVal, exists := merged[k]
		if !exists {
			merged[k] = overlayVal

			continue
		}

		merged[k] = mergeYamlValue(baseVal, overlayVal)
	}

	return merged
}

func mergeYamlLists(base, overlay []any) []any {
	if !isNamedObjectList(base) || !isNamedObjectList(overlay) {
		return overlay
	}

	merged := make([]any, 0, len(base)+len(overlay))

	merged = append(merged, base...)

	index := make(map[string]int, len(base))
	for i, item := range base {
		index[objectName(item)] = i
	}

	for _, overlayItem := range overlay {
		name := objectName(overlayItem)

		pos, exists := index[name]
		if !exists {
			index[name] = len(merged)
			merged = append(merged, overlayItem)

			continue
		}

		merged[pos] = mergeYamlValue(merged[pos], overlayItem)
	}

	return merged
}

// isNamedObjectList reports whether every element of list is a map carrying
// a "name" key - the shape the spec's "merge by name" rule applies to. An
// empty list, a list of scalars, or a list with any non-conforming element
// is not, and falls back to wholesale replacement.
func isNamedObjectList(list []any) bool {
	if len(list) == 0 {
		return false
	}

	for _, item := range list {
		itemMap, ok := item.(map[string]any)
		if !ok {
			return false
		}

		_, hasName := itemMap[nameKey]
		if !hasName {
			return false
		}
	}

	return true
}

func objectName(item any) string {
	itemMap, ok := item.(map[string]any)
	if !ok {
		return ""
	}

	name, _ := itemMap[nameKey].(string)

	return name
}
