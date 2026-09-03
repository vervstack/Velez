package vervonomicon

// Ssl is the closed vocabulary for how a domain's TLS is terminated.
type Ssl string

const (
	SslEnabled  Ssl = "enabled"
	SslDisabled Ssl = "disabled"
	SslExternal Ssl = "external"
)

// Ingress is the parsed content of ingress.yaml — what domain must be
// obtained and how it is terminated. The referenced config file is the
// verbatim webserver config, mounted through Domain.Config.
type Ingress struct {
	Domains []Domain `yaml:"domains"`
}

type Domain struct {
	Host   string `yaml:"host"`
	Ssl    Ssl    `yaml:"ssl,omitempty"`
	Listen int    `yaml:"listen,omitempty"`

	Upstreams map[string]Upstream `yaml:"upstreams,omitempty"`

	// Config - path (relative to .verv/) to the verbatim webserver config
	// fragment for this domain.
	Config string `yaml:"config,omitempty"`
}

type Upstream struct {
	// App - true routes to the primary container.
	App bool `yaml:"app,omitempty"`

	// Resource - name of a container-backed resource to route to.
	Resource string `yaml:"resource,omitempty"`
}
