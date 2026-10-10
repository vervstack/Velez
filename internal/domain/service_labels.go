package domain

import (
	"strings"

	"go.vervstack.ru/Velez/internal/domain/labels"
)

// Service classification labels. Derived, never persisted — the storage layer
// attaches them to ServiceBaseInfo.Labels and the transport layer forwards them.
const (
	LabelServiceCore         = "service-core"
	LabelServiceApp          = "service-app"
	LabelServiceRunnerGitlab = "service-runner-gitlab"
	LabelServiceRunnerGithub = "service-runner-github"
	LabelServicePgaas        = "service-pgaas"
	LabelServiceS3           = "service-s3"
	LabelServiceDind         = "service-dind"

	labelResourcePrefix = "resource-"

	// VelezServiceName is the name of the node manager's own service entry.
	VelezServiceName = "velez"
)

// CoreServiceNames - Verv infra services that are hidden from the default
// service list. Membership also drives the "service-core" label.
var CoreServiceNames = map[string]struct{}{
	"velez":     {},
	"matreshka": {},
	"makosh":    {},
	"headscale": {},
	"portainer": {},
	"angie":     {},
}

// IsCoreServiceName reports whether name is a known Verv infra service.
func IsCoreServiceName(name string) bool {
	_, ok := CoreServiceNames[name]

	return ok
}

// ResourceLabel builds the "resource-<type>" label for a bound resource.
func ResourceLabel(resourceType string) string {
	return labelResourcePrefix + resourceType
}

// IsResourceLabel reports whether label classifies an entry as a bound resource.
func IsResourceLabel(label string) bool {
	return strings.HasPrefix(label, labelResourcePrefix)
}

// ClassifyService returns the derived labels for a service list entry.
// resourceType is non-empty when the entry is bound as a resource somewhere;
// isDind marks a Docker-in-Docker instance.
func ClassifyService(name string, resourceType string, isDind bool) []string {
	if resourceType != "" {
		return []string{ResourceLabel(resourceType)}
	}

	if strings.HasPrefix(name, labels.GitlabRunnerNamePrefix) {
		return []string{LabelServiceRunnerGitlab}
	}

	if strings.HasPrefix(name, labels.GithubRunnerNamePrefix) {
		return []string{LabelServiceRunnerGithub}
	}

	if strings.HasPrefix(name, labels.PgaasNamePrefix) {
		return []string{LabelServicePgaas}
	}

	if strings.HasPrefix(name, labels.S3NamePrefix) {
		return []string{LabelServiceS3}
	}

	if isDind {
		return []string{LabelServiceDind}
	}

	if IsCoreServiceName(name) {
		return []string{LabelServiceCore}
	}

	return []string{LabelServiceApp}
}

// IsInternalLabels reports whether a set of derived labels marks the entry as
// Verv-internal (core service or bound resource) — i.e. hidden by default.
func IsInternalLabels(lbls []string) bool {
	for _, l := range lbls {
		if l == LabelServiceCore || IsResourceLabel(l) {
			return true
		}
	}

	return false
}
