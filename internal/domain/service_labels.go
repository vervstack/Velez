package domain

import (
	"strings"
)

// Service classification labels. Derived, never persisted — the storage layer
// attaches them to ServiceBaseInfo.Labels and the transport layer forwards them.
const (
	LabelServiceCore = "service-core"
	LabelServiceApp  = "service-app"

	labelResourcePrefix = "resource-"
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
// resourceType is non-empty when the entry is bound as a resource somewhere.
func ClassifyService(name string, resourceType string) []string {
	if resourceType != "" {
		return []string{ResourceLabel(resourceType)}
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
