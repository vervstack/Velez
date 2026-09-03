// Package vervonomicon holds the pure data shapes of the .verv/ deployment
// descriptor described in docs/features/vervonomicon.md. Parsing, merging and
// resolution live in later waves; this package only mirrors the YAML shape.
package vervonomicon

// SourceKind is the closed vocabulary for where a descriptor was found.
type SourceKind string

const (
	SourceKindImage  SourceKind = "image"
	SourceKindRepo   SourceKind = "repo"
	SourceKindPushed SourceKind = "pushed"
)

// Index is the parsed content of vervonomicon.yaml — the required index and
// identity file.
type Index struct {
	Version string  `yaml:"version"`
	Service Service `yaml:"service"`

	// Box - default sizing tier for the app container and every resource
	// that does not override it.
	Box string `yaml:"box,omitempty"`

	Config Config `yaml:"config,omitempty"`

	// Path overrides. All optional — an omitted field resolves to the
	// conventional filename, and a missing file is not an error.
	Deployment string `yaml:"deployment,omitempty"`
	Resources  string `yaml:"resources,omitempty"`
	Ingress    string `yaml:"ingress,omitempty"`
	Auth       string `yaml:"auth,omitempty"`
}

type Service struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Repo        string   `yaml:"repo,omitempty"`
	Tags        []string `yaml:"tags,omitempty"`
}

type Config struct {
	// Sensitive - matreshka key paths whose values are masked in the UI and
	// excluded from deployment snapshots.
	Sensitive []string `yaml:"sensitive,omitempty"`
}

// Descriptor is the fully assembled result of locating, merging and
// resolving every file under .verv/ for one deploy.
type Descriptor struct {
	Index      Index
	Deployment Deployment
	Resources  []Resource
	Ingress    Ingress
	Auth       Auth

	// Environment - which overlay was merged in. Empty for the default
	// environment.
	Environment string

	// Source - which of the three mechanisms (image, repo, pushed) this
	// descriptor was located through.
	Source SourceKind

	// Raw - every file found under .verv/, as found, keyed by path relative
	// to .verv/ (e.g. "vervonomicon.yaml", "prod/ingress.conf").
	Raw map[string][]byte
}
