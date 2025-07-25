package parser

import (
	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/toolbox"

	"go.vervstack.ru/Velez/pkg/velez_api"
)

const (
	// By default, container restarts maxRetryCountDefault times
	maxRetryCountDefault = 3
	// But not more that maxRetryCountBound times
	maxRetryCountBound = 10
)

func FromRestart(r *velez_api.RestartPolicy) container.RestartPolicy {
	rp := container.RestartPolicy{}

	maxRetryCount := maxRetryCountDefault
	if r != nil && r.FailureCount != nil {
		maxRetryCount = int(*r.FailureCount)
		maxRetryCount = min(maxRetryCount, maxRetryCountBound)
	}

	switch toolbox.FromPtr(r).Type {
	case velez_api.RestartPolicyType_no:
		rp.Name = container.RestartPolicyDisabled
		return container.RestartPolicy{}

	case velez_api.RestartPolicyType_always:
		rp.Name = container.RestartPolicyAlways
		rp.MaximumRetryCount = maxRetryCount

	case velez_api.RestartPolicyType_on_failure:
		rp.Name = container.RestartPolicyOnFailure
		rp.MaximumRetryCount = maxRetryCount

	case velez_api.RestartPolicyType_unless_stopped:
		rp.Name = container.RestartPolicyUnlessStopped
		rp.MaximumRetryCount = maxRetryCount

	}

	return rp
}
