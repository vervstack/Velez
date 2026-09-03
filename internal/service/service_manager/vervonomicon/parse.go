// Package vervonomicon implements parsing, environment-overlay merging, box
// resolution and image-sourced retrieval for the .verv/ deployment descriptor
// described in docs/features/vervonomicon.md. internal/domain/vervonomicon
// holds the pure data shapes this package fills in.
package vervonomicon

import (
	"strings"

	"go.redsock.ru/rerrors"
	"gopkg.in/yaml.v3"

	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
)

const (
	indexFileName = "vervonomicon.yaml"

	defaultDeploymentFile = "deployment.yaml"
	defaultResourcesFile  = "resources.yaml"
	defaultIngressFile    = "ingress.yaml"
	defaultAuthFile       = "auth.yaml"

	// recognisedMajorVersion is the only vervonomicon.yaml `version` this
	// Velez understands. Anything else is rejected outright rather than
	// partially applied.
	recognisedMajorVersion = "1"
)

// ErrNoDescriptor reports that a service simply has no vervonomicon
// descriptor. Per docs/features/vervonomicon.md: "Absence of a descriptor is
// never an error - the service simply deploys the way it does today." Callers
// must be able to tell this apart from a descriptor that exists but is
// broken - check with errors.Is.
var ErrNoDescriptor = rerrors.New("no vervonomicon descriptor found")

// Parse turns the flat file map read out of .verv/ - keyed by path relative
// to .verv/, exactly what dockerutils.ReadDirFromContainer returns - into a
// Descriptor.
//
// vervonomicon.yaml is required; its absence is reported as ErrNoDescriptor,
// not a hard error. The four path-override fields (deployment/resources/
// ingress/auth) resolve against their conventional defaults; a referenced
// file that is missing is simply absent, but one that is present and
// malformed is an error. Parse never partially applies an unrecognised
// descriptor version - it rejects before populating anything else.
func Parse(files map[string][]byte) (verv.Descriptor, error) {
	indexRaw, ok := files[indexFileName]
	if !ok {
		return verv.Descriptor{}, ErrNoDescriptor
	}

	var index verv.Index

	err := yaml.Unmarshal(indexRaw, &index)
	if err != nil {
		return verv.Descriptor{}, rerrors.Wrap(err, "error parsing "+indexFileName)
	}

	err = checkVersion(index.Version)
	if err != nil {
		return verv.Descriptor{}, err
	}

	descriptor := verv.Descriptor{
		Index: index,
		Raw:   files,
	}

	deploymentFile := resolvePath(index.Deployment, defaultDeploymentFile)

	descriptor.Deployment, err = parseOptionalYaml[verv.Deployment](files, deploymentFile)
	if err != nil {
		return verv.Descriptor{}, err
	}

	resourcesFile := resolvePath(index.Resources, defaultResourcesFile)

	descriptor.Resources, err = parseOptionalYaml[[]verv.Resource](files, resourcesFile)
	if err != nil {
		return verv.Descriptor{}, err
	}

	descriptor.Ingress, err = parseOptionalYaml[verv.Ingress](files, resolvePath(index.Ingress, defaultIngressFile))
	if err != nil {
		return verv.Descriptor{}, err
	}

	descriptor.Auth, err = parseOptionalYaml[verv.Auth](files, resolvePath(index.Auth, defaultAuthFile))
	if err != nil {
		return verv.Descriptor{}, err
	}

	return descriptor, nil
}

// checkVersion rejects any vervonomicon.yaml whose major version this Velez
// does not recognise. version is documented as a bare major ("1"), but a
// dotted form is tolerated by comparing only the part before the first ".".
func checkVersion(version string) error {
	major, _, _ := strings.Cut(version, ".")

	if major != recognisedMajorVersion {
		return rerrors.New("unrecognised vervonomicon version '" + version + "', only major version " +
			recognisedMajorVersion + " is supported")
	}

	return nil
}

// resolvePath applies a path-override field's conventional default when it
// was left empty.
func resolvePath(override, conventional string) string {
	if override == "" {
		return conventional
	}

	return override
}

// parseOptionalYaml decodes filename out of files into T. A missing file is
// not an error - it returns the zero value. A present-but-malformed file is.
func parseOptionalYaml[T any](files map[string][]byte, filename string) (T, error) {
	var zero T

	raw, ok := files[filename]
	if !ok {
		return zero, nil
	}

	var out T

	err := yaml.Unmarshal(raw, &out)
	if err != nil {
		return zero, rerrors.Wrap(err, "error parsing "+filename)
	}

	return out, nil
}
