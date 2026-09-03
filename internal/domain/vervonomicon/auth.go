package vervonomicon

// Auth is the parsed content of auth.yaml — access posture. Recorded and
// displayed today; nothing is enforced from this file yet.
type Auth struct {
	ForbidAll bool        `yaml:"forbid_all,omitempty"`
	Allow     []AllowRule `yaml:"allow,omitempty"`
}

type AllowRule struct {
	Service string   `yaml:"service"`
	Paths   []string `yaml:"paths"`
}
