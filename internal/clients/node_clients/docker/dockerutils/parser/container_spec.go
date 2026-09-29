package parser

import (
	"time"

	"github.com/docker/docker/api/types/container"

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
			Ports:   ToPortsMapping(info.HostConfig.PortBindings),
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

// ToHealthcheck reverses FromHealthcheck's "CMD-SHELL, command" Test shape.
// nil, or an empty/inherited Test, means the container carries no healthcheck.
func ToHealthcheck(hc *container.HealthConfig) *velez_api.Container_Healthcheck {
	if hc == nil || len(hc.Test) == 0 {
		return nil
	}

	command := hc.Test[len(hc.Test)-1]
	timeoutSecond := uint32(hc.Timeout / time.Second)

	return &velez_api.Container_Healthcheck{
		Command:        &command,
		IntervalSecond: uint32(hc.Interval / time.Second),
		TimeoutSecond:  &timeoutSecond,
		Retries:        uint32(hc.Retries),
	}
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
