package vervonomicon

// Isolation is the closed vocabulary for how a database-type resource is
// hosted.
type Isolation string

const (
	IsolationSharedPool       Isolation = "shared_pool"
	IsolationSeparateInstance Isolation = "separate_instance"
)

// Resource is one entry of resources.yaml — something Velez must provision
// or bind. A resource entry never carries host, port, user or pwd; those are
// resolved by Velez and written into matreshka.
type Resource struct {
	// Name - unique within the service; the service_resources binding key.
	Name string `yaml:"name"`

	// Type - "postgres", "redis", "local_volume", "nginx", ...
	Type string `yaml:"type"`

	// Isolation - database types only.
	Isolation Isolation `yaml:"isolation,omitempty"`

	// Box - sizing tier; Resources wins over Box when both are present,
	// exactly as in deployment.yaml.
	Box string `yaml:"box,omitempty"`

	Resources ResourceSizing `yaml:"resources,omitempty"`

	// BindsTo - matreshka key path the resolved connection is written to.
	BindsTo string `yaml:"binds_to,omitempty"`

	// ContainerPath - local_volume only: mount point inside the app
	// container.
	ContainerPath string `yaml:"container_path,omitempty"`

	// Sticky - local_volume only: pins the service to the node holding the
	// volume.
	Sticky bool `yaml:"sticky,omitempty"`

	// Config - container-backed resources only, each entry shaped
	// "<file in .verv>:<path in container>[:<mode>]".
	Config []string `yaml:"config,omitempty"`

	// Volumes - container-backed resources only: volumes mounted into the
	// resource's own container.
	Volumes []VolumeMount `yaml:"volumes,omitempty"`
}

// ResourceSizing is Sizing plus the disk allocation a resource's own
// container may need.
type ResourceSizing struct {
	Sizing `yaml:",inline"`

	DiskMb int64 `yaml:"disk_mb,omitempty"`
}
