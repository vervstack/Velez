package vervonomicon

// Protocol is the closed vocabulary for a published port's transport.
type Protocol string

// RestartPolicy is the closed vocabulary for a container's restart policy.
type RestartPolicy string

const (
	ProtocolTcp Protocol = "tcp"
	ProtocolUdp Protocol = "udp"

	RestartPolicyUnlessStopped RestartPolicy = "unless_stopped"
	RestartPolicyNo            RestartPolicy = "no"
	RestartPolicyAlways        RestartPolicy = "always"
	RestartPolicyOnFailure     RestartPolicy = "on_failure"
)

// Deployment is the parsed content of deployment.yaml — the primary
// container. Every field maps 1:1 onto CreateSmerd.Request.
type Deployment struct {
	App App `yaml:"app"`
}

type App struct {
	// Image - normally omitted. Set only for the repo-url and client-push
	// sources; a deploy request's image always wins.
	Image string `yaml:"image,omitempty"`

	// Command - optional; overrides the image entrypoint args.
	Command string `yaml:"command,omitempty"`

	// UseImagePorts - publish every port the image EXPOSEs.
	UseImagePorts bool `yaml:"use_image_ports,omitempty"`

	Ports   []Port        `yaml:"ports,omitempty"`
	Volumes []VolumeMount `yaml:"volumes,omitempty"`

	Env    map[string]string `yaml:"env,omitempty"`
	Labels map[string]string `yaml:"labels,omitempty"`

	Healthcheck Healthcheck `yaml:"healthcheck,omitempty"`

	Restart RestartPolicy `yaml:"restart,omitempty"`

	// RestartFailureCount - on_failure restart policy only.
	RestartFailureCount int `yaml:"restart_failure_count,omitempty"`

	AutoUpgrade bool `yaml:"auto_upgrade,omitempty"`

	// Box - optional; overrides the root-level box.
	Box string `yaml:"box,omitempty"`

	// Resources - optional; exact values, win over any box.
	Resources Sizing `yaml:"resources,omitempty"`
}

type Port struct {
	Port     int      `yaml:"port"`
	Protocol Protocol `yaml:"protocol,omitempty"`

	// ExposeTo - optional host port; omit to keep it internal.
	ExposeTo int `yaml:"expose_to,omitempty"`
}

type VolumeMount struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

type Healthcheck struct {
	// Command - empty means just wait for the container to reach Running.
	Command        string `yaml:"command,omitempty"`
	IntervalSecond int    `yaml:"interval_second,omitempty"`
	TimeoutSecond  int    `yaml:"timeout_second,omitempty"`
	Retries        int    `yaml:"retries,omitempty"`
}

// Sizing is an exact resource allocation. It wins over any box when both are
// present.
type Sizing struct {
	Cpu          float64 `yaml:"cpu,omitempty"`
	RamMb        int64   `yaml:"ram_mb,omitempty"`
	MemorySwapMb int64   `yaml:"memory_swap_mb,omitempty"`
}
