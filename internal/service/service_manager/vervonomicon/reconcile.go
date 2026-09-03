package vervonomicon

import (
	"go.vervstack.ru/matreshka/pkg/matreshka"
	"go.vervstack.ru/matreshka/pkg/matreshka/resources"

	"go.vervstack.ru/Velez/internal/domain"
	verv "go.vervstack.ru/Velez/internal/domain/vervonomicon"
)

const (
	// dataSourcesKeyPrefix is how every resources.yaml `binds_to` path this
	// reconciliation cares about is rooted, per the spec: "matreshka
	// data_sources holds an entry for it".
	dataSourcesKeyPrefix = "data_sources."

	placeholderHostEmpty     = ""
	placeholderHostLocalhost = "localhost"
	placeholderHostLoopback  = "127.0.0.1"
	placeholderHostAllZeros  = "0.0.0.0"
)

// placeholderHosts are the values a repo's committed config.yaml carries for
// local development - never a real provisioned connection, per
// docs/features/vervonomicon.md's "Resource reconciliation".
var placeholderHosts = map[string]bool{
	placeholderHostEmpty:     true,
	placeholderHostLocalhost: true,
	placeholderHostLoopback:  true,
	placeholderHostAllZeros:  true,
}

// ReconcileResources decides, for every resources.yaml entry, whether the
// service already has a connection to it or still needs one provisioned -
// per docs/features/vervonomicon.md's "Resource reconciliation". Pure
// decision logic: bindings and dataSourceHosts are both already-fetched data,
// no I/O happens here.
//
// bindings is the service's existing velez.service_resources rows.
// dataSourceHosts maps a matreshka key path (e.g. "data_sources.postgres",
// matching a Resource.BindsTo value) to the host currently configured there -
// see DataSourceHosts. A path absent from the map is treated exactly like a
// placeholder host: no live connection found there.
func ReconcileResources(
	list []verv.Resource, bindings []domain.BoundResource, dataSourceHosts map[string]string,
) []domain.ResourceReconciliation {
	if len(list) == 0 {
		return nil
	}

	bound := make(map[string]bool, len(bindings))
	for _, binding := range bindings {
		bound[binding.Name] = true
	}

	decisions := make([]domain.ResourceReconciliation, len(list))
	for i, res := range list {
		decisions[i] = reconcileOne(res, bound, dataSourceHosts)
	}

	return decisions
}

func reconcileOne(
	res verv.Resource, bound map[string]bool, dataSourceHosts map[string]string,
) domain.ResourceReconciliation {
	status := domain.ResourceMustProvision

	if bound[res.Name] || hasLiveConnection(res.BindsTo, dataSourceHosts) {
		status = domain.ResourceAlreadyConnected
	}

	return domain.ResourceReconciliation{
		Name:   res.Name,
		Type:   res.Type,
		Status: status,
	}
}

func hasLiveConnection(bindsTo string, dataSourceHosts map[string]string) bool {
	if bindsTo == "" {
		return false
	}

	host, found := dataSourceHosts[bindsTo]
	if !found {
		return false
	}

	return !placeholderHosts[host]
}

// DataSourceHosts extracts the currently configured host for every
// matreshka data_sources entry a service's live config carries, keyed by the
// dotted key path Velez would write a resource's connection into (matches
// Resource.BindsTo, e.g. "data_sources.postgres"). Only data source kinds
// with a host concept (postgres, redis) contribute an entry - the others
// (grpc connection strings, sqlite paths, telegram tokens, ...) don't carry
// the "provisioned or not" question this reconciliation answers.
func DataSourceHosts(cfg matreshka.AppConfig) map[string]string {
	hosts := make(map[string]string, len(cfg.DataSources))

	for _, res := range cfg.DataSources {
		host, ok := hostOf(res)
		if !ok {
			continue
		}

		hosts[dataSourcesKeyPrefix+res.GetName()] = host
	}

	return hosts
}

func hostOf(res resources.Resource) (string, bool) {
	switch typed := res.(type) {
	case *resources.Postgres:
		return typed.Host, true
	case *resources.Redis:
		return typed.Host, true
	default:
		return "", false
	}
}
