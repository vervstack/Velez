package parser

import (
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
)

// ToCreateRequest derives the CreateSmerd request that would recreate the
// inspected container: image, env, healthcheck, restart policy, ports and
// volumes. Labels and environment are the caller's to set. Settings without a
// CreateSmerd counterpart (caps, privileged, devices, ulimits, user,
// entrypoint, extra hosts, network aliases) are dropped.
func ToCreateRequest(name string, info container.InspectResponse) *velez_api.CreateSmerd_Request {
	req := &velez_api.CreateSmerd_Request{
		Name: name,
		Settings: &velez_api.Container_Settings{
			Ports:   ToPortsFromInspect(info),
			Volumes: ToVolume(info.Mounts),
		},
		Restart: ToRestartPolicy(info.HostConfig.RestartPolicy),
	}

	if info.Config != nil {
		req.ImageName = info.Config.Image
		req.Env = ToDockerEnv(info.Config.Env)
		req.Healthcheck = ToHealthcheck(info.Config.Healthcheck)
	}

	return req
}

// ToPortsFromInspect maps the inspected container's published ports, see
// ToPortsMapping.
func ToPortsFromInspect(info container.InspectResponse) []*velez_api.Port {
	var actual map[nat.Port][]nat.PortBinding

	if info.NetworkSettings != nil {
		actual = info.NetworkSettings.Ports
	}

	return ToPortsMapping(info.HostConfig.PortBindings, actual)
}

// ToHealthcheck reverses FromHealthcheck's "CMD-SHELL, command" and "CMD, exec..." Test shapes.
// nil, or an empty/inherited Test, means the container carries no healthcheck.
func ToHealthcheck(hc *container.HealthConfig) *velez_api.Container_Healthcheck {
	if hc == nil || len(hc.Test) == 0 {
		return nil
	}

	timeoutSecond := uint32(hc.Timeout / time.Second)

	out := &velez_api.Container_Healthcheck{
		IntervalSecond: uint32(hc.Interval / time.Second),
		TimeoutSecond:  &timeoutSecond,
		Retries:        uint32(hc.Retries),
	}

	if hc.Test[0] == healthcheckTestExec {
		out.Exec = hc.Test[1:]

		return out
	}

	command := hc.Test[len(hc.Test)-1]

	out.Command = &command

	return out
}

// ToRestartPolicy reverses FromRestart. always/on_failure/unless_stopped all
// collapse into container.RestartPolicyOnFailure on the way in, so that docker
// policy name can't be round-tripped back to which of the three was originally
// requested - on_failure is reported for it.
func ToRestartPolicy(rp container.RestartPolicy) *velez_api.RestartPolicy {
	policyType := velez_api.RestartPolicyType_unless_stopped

	switch rp.Name {
	case container.RestartPolicyDisabled, "":
		policyType = velez_api.RestartPolicyType_no
	case container.RestartPolicyOnFailure:
		policyType = velez_api.RestartPolicyType_on_failure
	case container.RestartPolicyAlways:
		policyType = velez_api.RestartPolicyType_always
	case container.RestartPolicyUnlessStopped:
		policyType = velez_api.RestartPolicyType_unless_stopped
	}

	result := &velez_api.RestartPolicy{Type: policyType}

	if rp.MaximumRetryCount > 0 {
		count := uint32(rp.MaximumRetryCount)

		result.FailureCount = &count
	}

	return result
}
