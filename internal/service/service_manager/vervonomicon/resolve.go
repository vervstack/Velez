package vervonomicon

import (
	"context"
	"maps"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
)

// ResolveRequest turns a fully merged Descriptor into the CreateSmerd.Request
// Velez launches from, per docs/features/vervonomicon.md's "Mapping onto
// CreateSmerd.Request" table. Only the primary app container is resolved
// here - resources[] provisioning and ingress are later waves' concern, not
// this one.
//
// deployEnvironment is the deploy's target environment (CreateDeploy_Request
// .environment) - it always wins, and always lands on the result's
// Environment field, never the Env map (that's app.env - container
// environment variables, a different thing entirely).
//
// requestImage is the image named on the deploy request itself (e.g.
// CreateDeploy_Request_FromVervonomicon.image). Per the spec, app.image is
// normally empty and exists only for the repo/pushed sources; a deploy
// request's image always wins when both are present.
func (r *BoxResolver) ResolveRequest(
	ctx context.Context, descriptor verv.Descriptor, deployEnvironment, requestImage string,
) (*velez_api.CreateSmerd_Request, error) {
	app := descriptor.Deployment.App

	sizing, err := r.ResolveApp(ctx, app, descriptor.Index.Box)
	if err != nil {
		return nil, err
	}

	request := &velez_api.CreateSmerd_Request{
		Name:                descriptor.Index.Service.Name,
		Repo:                optionalString(descriptor.Index.Service.Repo),
		ImageName:           resolveImage(app.Image, requestImage),
		Command:             optionalString(app.Command),
		Settings:            resolveSettings(app),
		Env:                 app.Env,
		Healthcheck:         resolveHealthcheck(app.Healthcheck),
		Labels:              resolveLabels(app.Labels, descriptor.Index.Service.Tags),
		UseImagePorts:       app.UseImagePorts,
		AutoUpgrade:         app.AutoUpgrade,
		Restart:             resolveRestart(app.Restart, app.RestartFailureCount),
		Hardware:            resolveHardware(sizing),
		IsDeclarativeDeploy: true,
		Environment:         deployEnvironment,
	}

	return request, nil
}

// resolveImage implements "a deploy request's image always wins": the
// request's image is used whenever it's set, and app.image (normally empty,
// present only for the repo/pushed sources) is the fallback.
func resolveImage(descriptorImage, requestImage string) string {
	if requestImage != "" {
		return requestImage
	}

	return descriptorImage
}

// resolveLabels merges app.labels with one verv.tag.<tag> entry per
// service.tags entry. Keys collide in app.labels' favor is not a concern the
// spec addresses; map iteration order means either could win on an actual
// key clash between a tag and a hand-written label, which is expected to
// never happen in practice (tags and labels are different vocabularies).
func resolveLabels(appLabels map[string]string, tags []string) map[string]string {
	if len(appLabels) == 0 && len(tags) == 0 {
		return nil
	}

	result := make(map[string]string, len(appLabels)+len(tags))
	maps.Copy(result, appLabels)

	for _, tag := range tags {
		result[labels.TagLabelPrefix+tag] = labels.TagLabelValue
	}

	return result
}

func resolveSettings(app verv.App) *velez_api.Container_Settings {
	ports := resolvePorts(app.Ports)
	volumes := resolveVolumes(app.Volumes)

	if len(ports) == 0 && len(volumes) == 0 {
		return nil
	}

	settings := &velez_api.Container_Settings{
		Ports:   ports,
		Volumes: volumes,
	}

	return settings
}

func resolvePorts(ports []verv.Port) []*velez_api.Port {
	if len(ports) == 0 {
		return nil
	}

	out := make([]*velez_api.Port, len(ports))

	for i, p := range ports {
		port := &velez_api.Port{
			ServicePortNumber: uint32(p.Port),
			Protocol:          resolveProtocol(p.Protocol),
		}

		if p.ExposeTo != 0 {
			exposedTo := uint32(p.ExposeTo)

			port.ExposedTo = &exposedTo
		}

		out[i] = port
	}

	return out
}

// resolveProtocol defaults an unset protocol to tcp - the descriptor's own
// examples always set it explicitly, but the field is optional and tcp is
// the sane default for a published port.
func resolveProtocol(protocol verv.Protocol) velez_api.Port_Protocol {
	switch protocol {
	case verv.ProtocolUdp:
		return velez_api.Port_udp
	case verv.ProtocolTcp:
		return velez_api.Port_tcp
	default:
		return velez_api.Port_tcp
	}
}

func resolveVolumes(volumes []verv.VolumeMount) []*velez_api.Volume {
	if len(volumes) == 0 {
		return nil
	}

	out := make([]*velez_api.Volume, len(volumes))

	for i, v := range volumes {
		out[i] = &velez_api.Volume{
			VolumeName:    v.Name,
			ContainerPath: v.Path,
		}
	}

	return out
}

func resolveHealthcheck(hc verv.Healthcheck) *velez_api.Container_Healthcheck {
	if hc == (verv.Healthcheck{}) {
		return nil
	}

	out := &velez_api.Container_Healthcheck{
		Command:        optionalString(hc.Command),
		IntervalSecond: uint32(hc.IntervalSecond),
		Retries:        uint32(hc.Retries),
	}

	if hc.TimeoutSecond != 0 {
		timeout := uint32(hc.TimeoutSecond)

		out.TimeoutSecond = &timeout
	}

	return out
}

// resolveRestart returns nil when the descriptor never set a restart policy
// at all, so CreateSmerd's own default applies unchanged.
func resolveRestart(restart verv.RestartPolicy, failureCount int) *velez_api.RestartPolicy {
	if restart == "" {
		return nil
	}

	out := &velez_api.RestartPolicy{
		Type: resolveRestartType(restart),
	}

	if restart == verv.RestartPolicyOnFailure && failureCount != 0 {
		count := uint32(failureCount)

		out.FailureCount = &count
	}

	return out
}

func resolveRestartType(restart verv.RestartPolicy) velez_api.RestartPolicyType {
	switch restart {
	case verv.RestartPolicyNo:
		return velez_api.RestartPolicyType_no
	case verv.RestartPolicyAlways:
		return velez_api.RestartPolicyType_always
	case verv.RestartPolicyOnFailure:
		return velez_api.RestartPolicyType_on_failure
	case verv.RestartPolicyUnlessStopped:
		return velez_api.RestartPolicyType_unless_stopped
	default:
		return velez_api.RestartPolicyType_unless_stopped
	}
}

func resolveHardware(sizing verv.Sizing) *velez_api.Container_Hardware {
	if sizing == (verv.Sizing{}) {
		return nil
	}

	out := &velez_api.Container_Hardware{}

	if sizing.Cpu != 0 {
		cpu := float32(sizing.Cpu)

		out.Cpu = &cpu
	}

	if sizing.RamMb != 0 {
		ramMb := uint32(sizing.RamMb)

		out.RamMb = &ramMb
	}

	if sizing.MemorySwapMb != 0 {
		swapMb := uint32(sizing.MemorySwapMb)

		out.MemorySwapMb = &swapMb
	}

	return out
}

func optionalString(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}
