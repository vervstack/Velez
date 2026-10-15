package verv_services

import (
	"context"
	"net/url"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/proxyenv"
)

const (
	resourceTypeDind  = "dind"
	resourceTypeProxy = "proxy"
)

// runnerLink is a runner's DinD or proxy. Many runners can point at one, so a link
// is resolved per runner and never owned by it.
type runnerLink struct {
	Name         string
	ResourceType string
	Status       string
	IsManaged    bool
}

func (v *VervService) runnerLinks(ctx context.Context, serviceName string) ([]runnerLink, error) {
	req := &velez_api.ListSmerds_Request{
		Name: &serviceName,
	}

	resp, err := v.containerService.ListSmerds(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing smerds for runner links")
	}

	smerds := resp.GetSmerds()
	if len(smerds) == 0 {
		return nil, nil
	}

	smerd := smerds[0]

	_, isRunner := smerd.GetLabels()[labels.RunnerInstanceLabel]
	if !isRunner {
		return nil, nil
	}

	containers, err := v.listAllContainers(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers for runner links")
	}

	var links []runnerLink

	dindName := smerd.GetLabels()[labels.RunnerDindLabel]
	if dindName != "" {
		dind := runnerLink{
			Name:         dindName,
			ResourceType: resourceTypeDind,
			IsManaged:    true,
		}

		dindContainer, isFound := findManagedContainer(containers, dindName)
		if isFound {
			dind.Status = dindContainer.State
		}

		links = append(links, dind)
	}

	inspected, err := v.containerService.InspectSmerd(ctx, "", smerd.GetUuid())
	if err != nil {
		return nil, rerrors.Wrap(err, "error inspecting runner for proxy")
	}

	proxyUrl, _ := proxyenv.Parse(inspected.GetEnv())

	proxyHost := proxyHostOf(proxyUrl)
	if proxyHost != "" {
		proxy := runnerLink{
			Name:         proxyHost,
			ResourceType: resourceTypeProxy,
		}

		proxyContainer, isFound := findManagedContainer(containers, proxyHost)
		if isFound {
			proxy.Name = proxyContainer.Labels[labels.VervServiceLabel]
			proxy.Status = proxyContainer.State
			proxy.IsManaged = true
		}

		links = append(links, proxy)
	}

	return links, nil
}

// dindCallers are the runners whose DinD is serviceName.
func (v *VervService) dindCallers(ctx context.Context, serviceName string) ([]string, error) {
	containers, err := v.listAllContainers(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers for dind callers")
	}

	var runners []string

	for _, cont := range containers {
		if cont.Labels[labels.RunnerDindLabel] != serviceName {
			continue
		}

		runner := cont.Labels[labels.VervServiceLabel]
		if runner == "" {
			continue
		}

		runners = append(runners, runner)
	}

	return runners, nil
}

func (v *VervService) listAllContainers(ctx context.Context) ([]container.Summary, error) {
	runtime, err := v.runtimes.Runtime(ctx, "")
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving environment")
	}

	containers, err := runtime.ListAllContainers(ctx, listAllContainersNoLimit)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	return containers, nil
}

// findManagedContainer finds a Velez-managed (VERV_SERVICE-labelled, non-sidecar)
// container by its service label or its container name.
func findManagedContainer(containers []container.Summary, name string) (container.Summary, bool) {
	for _, cont := range containers {
		_, isSidecar := cont.Labels[labels.Sidecar]
		if isSidecar {
			continue
		}

		serviceLabel := cont.Labels[labels.VervServiceLabel]
		if serviceLabel == "" {
			continue
		}

		if serviceLabel == name || summaryName(cont) == name {
			return cont, true
		}
	}

	return container.Summary{}, false
}

func proxyHostOf(proxyUrl string) string {
	if proxyUrl == "" {
		return ""
	}

	parsed, err := url.Parse(proxyUrl)
	if err != nil {
		return ""
	}

	return parsed.Hostname()
}
