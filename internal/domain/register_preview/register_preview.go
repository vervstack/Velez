package register_preview

import (
	"maps"
	"sort"
	"strings"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

const (
	labelTrueValue = "true"

	NetworkModeOfPrefix = "container:"
)

type RegisterPreviewInput struct {
	// Members are the docker inspect results of the root and every container sharing its network namespace.
	Members     []container.InspectResponse
	ServiceName string
	// RootId and ServiceId are full container ids.
	RootId    string
	ServiceId string
	// PreviousRootId is the container that owns the network namespace today when RootId replaces it; empty when the
	// root stays.
	PreviousRootId string
	Pattern        velez_api.ServicePattern
}

// BuildRegisterPreview returns the spec each member of an onboarded group gets, every value
// tagged by whether it already sits on the container or onboarding stamps it.
func BuildRegisterPreview(input RegisterPreviewInput) ([]*velez_api.ContainerPreview, error) {
	root, isRootFound := findMember(input.Members, input.RootId)
	_, isServiceFound := findMember(input.Members, input.ServiceId)

	if !isRootFound || !isServiceFound {
		return nil, rerrors.Wrap(user_errors.ErrPreviewRoleNotInGroup)
	}

	rootName := strings.TrimPrefix(root.Name, "/")

	previousRoot := root

	if input.PreviousRootId != "" {
		var isPreviousFound bool

		previousRoot, isPreviousFound = findMember(input.Members, input.PreviousRootId)
		if !isPreviousFound {
			return nil, rerrors.Wrap(user_errors.ErrPreviewRoleNotInGroup)
		}
	}

	previews := make([]*velez_api.ContainerPreview, 0, len(input.Members))

	for _, member := range input.Members {
		isRoot := member.ID == input.RootId
		isService := member.ID == input.ServiceId

		preview := &velez_api.ContainerPreview{
			ContainerId: member.ID,
			Name:        strings.TrimPrefix(member.Name, "/"),
			IsRoot:      isRoot,
			IsService:   isService,
			Env:         previewEnv(member),
			Labels:      previewLabels(member, input, isRoot, isService),
			Mounts:      previewMounts(member),
			NetworkMode: previewNetworkMode(member, root, previousRoot, rootName, isRoot),
		}

		if member.Config != nil {
			preview.ImageName = member.Config.Image
		}

		if isRoot {
			preview.Ports = previewPorts(previousRoot)
		}

		previews = append(previews, preview)
	}

	return previews, nil
}

func findMember(members []container.InspectResponse, id string) (container.InspectResponse, bool) {
	for _, member := range members {
		if member.ID == id {
			return member, true
		}
	}

	return container.InspectResponse{}, false
}

func previewEnv(member container.InspectResponse) []*velez_api.PreviewValue {
	if member.Config == nil {
		return nil
	}

	values := make([]*velez_api.PreviewValue, 0, len(member.Config.Env))

	for _, entry := range member.Config.Env {
		key, value, _ := strings.Cut(entry, "=")

		values = append(values, transferredValue(key, value))
	}

	return sortedValues(values)
}

func previewLabels(
	member container.InspectResponse, input RegisterPreviewInput, isRoot, isService bool,
) []*velez_api.PreviewValue {
	var original map[string]string

	if member.Config != nil {
		original = member.Config.Labels
	}

	target := make(map[string]string, len(original))
	maps.Copy(target, original)

	if isService {
		maps.Copy(target, RegisteredLabels(input.ServiceName, input.Pattern))
	} else {
		maps.Copy(target, SidecarLabels(input.ServiceName))

		delete(target, labels.DisplayNameLabel)
	}

	// Only the root is recreated by an upgrade that records the container it replaces. A new root replaces the
	// root it takes the network namespace from.
	if isRoot {
		target[labels.OnboardedFromLabel] = replacedRootId(member, input)
	}

	values := make([]*velez_api.PreviewValue, 0, len(target))

	for key, value := range target {
		values = append(values, originValue(key, value, original))
	}

	return sortedValues(values)
}

func replacedRootId(root container.InspectResponse, input RegisterPreviewInput) string {
	if input.PreviousRootId != "" {
		return input.PreviousRootId
	}

	return root.ID
}

func previewMounts(member container.InspectResponse) []*velez_api.PreviewValue {
	values := make([]*velez_api.PreviewValue, 0, len(member.Mounts))

	for _, mount := range member.Mounts {
		values = append(values, transferredValue(mount.Destination, mount.Source))
	}

	return sortedValues(values)
}

func previewPorts(member container.InspectResponse) []*velez_api.PreviewValue {
	if member.HostConfig == nil {
		return nil
	}

	values := make([]*velez_api.PreviewValue, 0, len(member.HostConfig.PortBindings))

	for port, bindings := range member.HostConfig.PortBindings {
		for _, binding := range bindings {
			values = append(values, transferredValue(string(port), binding.HostIP+":"+binding.HostPort))
		}
	}

	return sortedValues(values)
}

func previewNetworkMode(
	member, root, previousRoot container.InspectResponse, rootName string, isRoot bool,
) *velez_api.PreviewValue {
	currentMode := networkModeOf(member)

	if isRoot {
		value := transferredValue("network_mode", networkModeOf(previousRoot))

		// A new root takes over the network mode of the root it replaces.
		if value.GetValue() != currentMode {
			value.Origin = velez_api.ValueOrigin_ADDED
		}

		return value
	}

	target := NetworkModeOfPrefix + rootName

	value := transferredValue("network_mode", target)

	if !isModeOfRoot(currentMode, root, rootName) {
		value.Origin = velez_api.ValueOrigin_ADDED
	}

	return value
}

func networkModeOf(member container.InspectResponse) string {
	if member.HostConfig == nil {
		return ""
	}

	return string(member.HostConfig.NetworkMode)
}

func isModeOfRoot(mode string, root container.InspectResponse, rootName string) bool {
	ref, isShared := strings.CutPrefix(mode, NetworkModeOfPrefix)
	if !isShared || ref == "" {
		return false
	}

	return ref == rootName || strings.HasPrefix(root.ID, ref)
}

func originValue(key, value string, original map[string]string) *velez_api.PreviewValue {
	originalValue, isPresent := original[key]
	if isPresent && originalValue == value {
		return transferredValue(key, value)
	}

	return &velez_api.PreviewValue{Key: key, Value: value, Origin: velez_api.ValueOrigin_ADDED}
}

func transferredValue(key, value string) *velez_api.PreviewValue {
	return &velez_api.PreviewValue{Key: key, Value: value, Origin: velez_api.ValueOrigin_TRANSFERRED}
}

func sortedValues(values []*velez_api.PreviewValue) []*velez_api.PreviewValue {
	sort.SliceStable(values, func(i, j int) bool {
		if values[i].GetKey() != values[j].GetKey() {
			return values[i].GetKey() < values[j].GetKey()
		}

		return values[i].GetValue() < values[j].GetValue()
	})

	return values
}

func RegisteredServiceLabels(serviceName string) map[string]string {
	return map[string]string{
		labels.VervServiceLabel: serviceName,
		labels.DisplayNameLabel: serviceName,
	}
}

func RegisteredPatternLabels(pattern velez_api.ServicePattern) map[string]string {
	if pattern != velez_api.ServicePattern_SERVICE_PATTERN_POSTGRES {
		return map[string]string{}
	}

	return map[string]string{labels.PgaasInstanceLabel: labelTrueValue}
}

func RegisteredLabels(serviceName string, pattern velez_api.ServicePattern) map[string]string {
	merged := RegisteredServiceLabels(serviceName)

	maps.Copy(merged, RegisteredPatternLabels(pattern))

	return merged
}

// SidecarLabels are the labels a recreated sidecar carries: the service link
// plus the sidecar marker, which keeps it off the service pages. It carries no
// display name - that belongs to the service's main container.
func SidecarLabels(serviceName string) map[string]string {
	merged := make(map[string]string)

	if serviceName != "" {
		merged[labels.VervServiceLabel] = serviceName
	}

	merged[labels.Sidecar] = labelTrueValue

	return merged
}
