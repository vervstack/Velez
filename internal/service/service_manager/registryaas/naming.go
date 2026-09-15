package registryaas

// registryaasUiServiceSuffix derives a registry instance's UI sidecar
// service name - mirrors internal/jobs's own copy of this suffix (the job
// deploys the sidecar under this name; DropRegistryInstance needs the same
// name to remove it). Kept as a small, self-contained duplicate rather than
// exported from internal/jobs - jobs already depends on the service package,
// so the reverse import would cycle.
const (
	registryaasUiServiceSuffix = "-ui"
)

func registryaasUiServiceName(instanceName string) string {
	return instanceName + registryaasUiServiceSuffix
}
