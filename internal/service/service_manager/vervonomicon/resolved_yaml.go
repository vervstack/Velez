package vervonomicon

import (
	"strings"

	"go.redsock.ru/rerrors"
	"gopkg.in/yaml.v3"

	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
)

const (
	// maskedValue replaces a sensitive value in resolved_yaml, per
	// docs/features/vervonomicon.md's "config.sensitive values are masked in
	// the UI and excluded from deployment snapshots".
	maskedValue = "***"

	// keysPerSensitivePath - sensitiveLeafKeys keeps both the full dotted
	// path and its last segment for every sensitive path.
	keysPerSensitivePath = 2
)

// resolvedView is the merged descriptor's on-the-wire YAML shape: index and
// deployment/resources/ingress/auth flattened into one document, the way an
// operator would actually author .verv/ before it's split across files.
type resolvedView struct {
	Version     string          `yaml:"version"`
	Service     verv.Service    `yaml:"service"`
	Box         string          `yaml:"box,omitempty"`
	Config      verv.Config     `yaml:"config,omitempty"`
	Environment string          `yaml:"environment,omitempty"`
	Deployment  verv.Deployment `yaml:"deployment,omitempty"`
	Resources   []verv.Resource `yaml:"resources,omitempty"`
	Ingress     verv.Ingress    `yaml:"ingress,omitempty"`
	Auth        verv.Auth       `yaml:"auth,omitempty"`
}

// MarshalResolvedYaml re-marshals a merged Descriptor to YAML for the
// GetVervonomicon RPC's resolved_yaml field, masking every value whose key
// matches one of config.sensitive's matreshka key paths.
//
// The descriptor never actually mirrors matreshka's own nested document
// shape (matreshka values are never copied into the descriptor - see the
// spec's "matreshka values never live in the descriptor itself"), so a
// sensitive path can't be resolved against this document by walking it
// segment by segment. Masking instead matches each sensitive path's last
// dotted segment against any map key anywhere in the resolved document -
// a defensive, forward-compatible best effort (e.g. it catches an app.env
// entry literally named after a sensitive key) rather than a guarantee tied
// to a structural correspondence that doesn't exist today. Flagged as an
// interpretation of an underspecified part of the spec.
func MarshalResolvedYaml(descriptor verv.Descriptor) ([]byte, error) {
	view := resolvedView{
		Version:     descriptor.Index.Version,
		Service:     descriptor.Index.Service,
		Box:         descriptor.Index.Box,
		Config:      descriptor.Index.Config,
		Environment: descriptor.Environment,
		Deployment:  descriptor.Deployment,
		Resources:   descriptor.Resources,
		Ingress:     descriptor.Ingress,
		Auth:        descriptor.Auth,
	}

	raw, err := yaml.Marshal(view)
	if err != nil {
		return nil, rerrors.Wrap(err, "error marshalling resolved vervonomicon yaml")
	}

	return redactSensitiveValues(raw, descriptor.Index.Config.Sensitive)
}

func redactSensitiveValues(raw []byte, sensitive []string) ([]byte, error) {
	if len(sensitive) == 0 {
		return raw, nil
	}

	var doc any

	err := yaml.Unmarshal(raw, &doc)
	if err != nil {
		return nil, rerrors.Wrap(err, "error parsing resolved yaml for redaction")
	}

	keys := sensitiveLeafKeys(sensitive)
	maskSensitiveValues(doc, keys)

	redacted, err := yaml.Marshal(doc)
	if err != nil {
		return nil, rerrors.Wrap(err, "error re-marshalling redacted yaml")
	}

	return redacted, nil
}

// sensitiveLeafKeys reduces every dotted matreshka key path to its last
// segment, and keeps the full path too - either can match a document key.
func sensitiveLeafKeys(sensitive []string) map[string]bool {
	keys := make(map[string]bool, len(sensitive)*keysPerSensitivePath)

	for _, path := range sensitive {
		keys[path] = true
		keys[lastSegment(path)] = true
	}

	return keys
}

func lastSegment(path string) string {
	idx := strings.LastIndex(path, ".")
	if idx < 0 {
		return path
	}

	return path[idx+1:]
}

func maskSensitiveValues(node any, keys map[string]bool) {
	switch v := node.(type) {
	case map[string]any:
		for k, val := range v {
			if keys[k] {
				v[k] = maskedValue

				continue
			}

			maskSensitiveValues(val, keys)
		}
	case []any:
		for _, item := range v {
			maskSensitiveValues(item, keys)
		}
	}
}
