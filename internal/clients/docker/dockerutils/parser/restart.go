package parser

import (
	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/pkg/velez_api"
)

func FromRestart(r *velez_api.RestartPolicy) container.RestartPolicy {
	rp := container.RestartPolicy{}

	switch r.Type {
	case velez_api.RestartPolicyType_no:
		rp.Name = container.RestartPolicyDisabled
		return container.RestartPolicy{}

	case velez_api.RestartPolicyType_always:
		rp.Name = container.RestartPolicyAlways

	case velez_api.RestartPolicyType_on_failure:
		rp.Name = container.RestartPolicyOnFailure
		rp.MaximumRetryCount = int(toolbox.Coalesce(toolbox.FromPtr(r.FailureCount), uint32(3)))

	case velez_api.RestartPolicyType_unless_stopped:
		rp.Name = container.RestartPolicyUnlessStopped

	}

	return rp
}
