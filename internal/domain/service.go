package domain

import (
	"time"

	rtb "go.redsock.ru/toolbox"
	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
)

type CreateServiceReq struct {
	Name string
}

type AboutService struct {
	Description  string
	OriginalName string
	Env          string
	ServiceType  string
	Team         string
	Repo         string
	Port         string
}

type Service struct {
	ServiceBaseInfo

	ID int64

	CurrentDeploymentId *uint64
	Status              velez_api.DeploymentStatus
	About               AboutService

	Sidecars []ServiceSidecar

	// ProxyUrl and ProxyBypassHosts are read from the running container's env;
	// empty when no proxy is configured.
	ProxyUrl         string
	ProxyBypassHosts []string
}

type ServiceSidecar struct {
	ContainerId   string
	ContainerName string
	ImageName     string
	Status        velez_api.Smerd_Status
}

type ServiceBaseInfo struct {
	Name string
	// DisplayName - the human-facing service name. Falls back to Name when no
	// clean name was recorded (see labels.DisplayNameLabel).
	DisplayName    string
	LastDeployedAt *time.Time
	ImageName      string
	Status         string
	Env            string
	Repo           string
	// Labels - derived, never stored. One of "service-core", "service-app",
	// "resource-<type>". Populated by the storage layer via ClassifyService.
	Labels []string
}

type GetServiceReq struct {
	Name string
}

type RemoveServiceReq struct {
	Name                 string
	DropRunningInstances bool

	// Environment - the isolated namespace this operation targets. Empty
	// means the default/PROD environment (see storage/environments.Resolve).
	Environment string
}

type CreateDeployReq struct {
	LaunchSmerd

	ServiceName string

	// DisplayName - the human-facing service name to persist alongside
	// ServiceName. Empty means the storage layer falls back to ServiceName.
	DisplayName string

	// VervDescriptor - set only for vervonomicon-driven deploys (see
	// CreateDeployFromVervonomicon); persisted into
	// deployment_specifications.verv_descriptor alongside the resolved
	// CreateSmerd.Request, so the deployment stays reproducible and
	// auditable.
	VervDescriptor *verv.Descriptor
}

// SetServiceProxyReq recreates a running service's container with the proxy
// env; an empty ProxyUrl removes it.
type SetServiceProxyReq struct {
	ServiceName      string
	ProxyUrl         string
	ProxyBypassHosts []string
}

type UpgradeDeployReq struct {
	ServiceName  string
	DeploymentId uint64

	NewImage *string

	// EnvOverrides overlays onto the current spec's env before the upgrade -
	// a value of "" deletes that key instead of setting it to empty, since
	// no env var this pipeline manages is ever meaningfully "set to empty"
	// (e.g. DOCKER_HOST). Nil/absent keys are left untouched.
	EnvOverrides map[string]string

	// ExtraNetworks are added to the current spec's networks before the
	// upgrade; a network already in the spec (by name) is left untouched.
	ExtraNetworks []*velez_api.NetworkBind
}

type ListServicesReq struct {
	Paging      Paging
	NamePattern rtb.Optional[string]

	// IncludeInternal - when false, Verv-internal entries (core services and
	// anything bound as a resource) are excluded from the result.
	IncludeInternal bool
}

type ServiceList struct {
	Total    uint64
	Services []ServiceBaseInfo
}
