package local_storage

import (
	"context"
	"hash/fnv"
	"strings"
	"sync"

	"github.com/docker/docker/api/types/container"
	"go.redsock.ru/rerrors"

	pb "go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/clients/node_clients"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/user_errors"
)

type dockerServices struct {
	docker node_clients.Docker

	mu sync.Mutex
	// upserted tracks names UpsertService has recorded that GetByName can't
	// yet derive from a live container - a service is registered before its
	// container exists (CreateNewDeploy.GetByName runs before the deploy
	// watcher ever creates that container, see enable_registry.go's
	// deployRegistryJob and pgaas.CreatePgInstance, which both do
	// UpsertService then immediately CreateNewDeploy). This backend has no
	// persisted services table to fall back on, so without this overlay
	// GetByName reports ErrNotFound for a service that was just upserted,
	// breaking every vervonomicon deploy under single-node/dev mode. The
	// value is the display name UpsertService was called with (may be
	// empty - pendingService falls back to name).
	upserted map[string]string
}

func newServicesStorage(docker node_clients.Docker) *dockerServices {
	return &dockerServices{
		docker:   docker,
		upserted: make(map[string]string),
	}
}

func (s *dockerServices) GetByName(ctx context.Context, name string) (domain.Service, error) {
	// Matched by the VervServiceLabel, not the Docker container Name filter:
	// under a non-empty node.ContainerSuffix the real Docker container name is
	// "<name>_<suffix>" (labelSuffixResolver.ContainerName) while the label
	// stays the bare virtual name - see internal/domain/labels.VervServiceLabel's
	// doc comment and listDistinctServices below, which already resolves the
	// service list the same way. A Name-filtered lookup here used to disagree
	// with that list (a service could show up in List but 404 on GetByName),
	// which broke every runner/registry/pg-instance create flow that calls
	// GetByName right after its container exists (e.g.
	// internal/jobs/create_runner.go's registerRunnerRowJob).
	listReq := &pb.ListSmerds_Request{
		Label: map[string]string{labels.VervServiceLabel: name},
	}

	containers, err := s.docker.ListContainers(ctx, listReq, allEnvironments)
	if err != nil {
		return domain.Service{}, rerrors.Wrap(err, "error listing containers")
	}

	if len(containers) == 0 {
		if name == domain.VelezServiceName {
			return s.syntheticVelezService(ctx)
		}

		s.mu.Lock()

		displayName, ok := s.upserted[name]

		s.mu.Unlock()

		if ok {
			return s.pendingService(name, displayName), nil
		}

		return domain.Service{}, user_errors.ErrStorageNotFound
	}

	c := mainContainersFirst(containers)[0]
	svc := domain.Service{
		ServiceBaseInfo: domain.ServiceBaseInfo{
			Name:        name,
			DisplayName: displayNameOrFallback(c.Labels, name),
			ImageName:   c.Image,
			Status:      containerStateToString(c.State),
		},
		ID:     serviceIDFromName(name),
		Status: containerStateToDeploymentStatus(c.State),
	}

	return svc, nil
}

// mainContainersFirst moves sidecars behind the main containers, so the first
// container seen for a service is the one whose labels (display name) describe
// the service rather than its sidecar.
func mainContainersFirst(containers []container.Summary) []container.Summary {
	ordered := make([]container.Summary, 0, len(containers))
	sidecars := make([]container.Summary, 0)

	for _, c := range containers {
		_, isSidecar := c.Labels[labels.Sidecar]
		if isSidecar {
			sidecars = append(sidecars, c)

			continue
		}

		ordered = append(ordered, c)
	}

	return append(ordered, sidecars...)
}

// serviceIDFromName derives a stable, non-zero synthetic service id from a
// service name. Single-node/dev mode has no velez.services table and thus no
// real serial ids, but storage.PgInstancesStorage and pgaas's pg_instances
// <-> service join are keyed by int64 service id (see
// internal/service/service_manager/pgaas/list.go). A name-derived FNV-1a hash
// gives every caller the same id for the same service across restarts with no
// persisted counter. The single right shift keeps the result positive so it
// never collides with the zero value that signalled "no id" before.
func serviceIDFromName(name string) int64 {
	h := fnv.New64a()

	_, _ = h.Write([]byte(name))

	return int64(h.Sum64() >> 1)
}

func containerStateToDeploymentStatus(state string) pb.DeploymentStatus {
	switch state {
	case containerStateRunning:
		return pb.DeploymentStatus_RUNNING
	case containerStateDead:
		return pb.DeploymentStatus_FAILED
	case containerStateExited:
		return pb.DeploymentStatus_STOPPED
	default:
		return pb.DeploymentStatus_DEPLOYMENT_STATUS_UNKNOWN
	}
}

func containerStateToString(state string) string {
	switch state {
	case containerStateRunning:
		return containerStateRunning
	case containerStatePaused:
		return containerStateDegraded
	case containerStateExited, containerStateDead:
		return containerStateStopped
	default:
		return ""
	}
}

// UpsertService records name so GetByName can resolve it before any
// container backing it exists yet - see the upserted field's doc comment.
func (s *dockerServices) UpsertService(_ context.Context, name string, displayName string) error {
	s.mu.Lock()

	s.upserted[name] = displayName

	s.mu.Unlock()

	return nil
}

// displayNameOrFallback reads labels.DisplayNameLabel off a container's
// labels, falling back to name for a container created before that label
// existed.
func displayNameOrFallback(containerLabels map[string]string, name string) string {
	displayName := containerLabels[labels.DisplayNameLabel]
	if displayName == "" {
		return name
	}

	return displayName
}

// Delete drops name from the upserted overlay - the underlying service list
// otherwise derives purely from live Docker container labels (see List), so
// dropping the containers is what actually makes a service disappear; this
// only prevents a deleted-and-never-redeployed name from resolving stale via
// the overlay above.
func (s *dockerServices) Delete(_ context.Context, name string) error {
	s.mu.Lock()
	delete(s.upserted, name)
	s.mu.Unlock()

	return nil
}

func (s *dockerServices) List(ctx context.Context, req domain.ListServicesReq) (domain.ServiceList, error) {
	all, err := listDistinctServices(ctx, s.docker)
	if err != nil {
		return domain.ServiceList{}, err
	}

	if req.NamePattern.Valid {
		pattern := strings.ToLower(req.NamePattern.Value)

		filtered := all[:0]
		for _, svc := range all {
			if strings.Contains(strings.ToLower(svc.Name), pattern) {
				filtered = append(filtered, svc)
			}
		}

		all = filtered
	}

	if !req.IncludeInternal {
		visible := all[:0]
		for _, svc := range all {
			if domain.IsInternalLabels(svc.Labels) {
				continue
			}

			visible = append(visible, svc)
		}

		all = visible
	}

	total := uint64(len(all))

	if req.Paging.Offset < total {
		all = all[req.Paging.Offset:]
	} else {
		all = nil
	}

	if req.Paging.Limit > 0 && uint64(len(all)) > req.Paging.Limit {
		all = all[:req.Paging.Limit]
	}

	out := domain.ServiceList{
		Total:    total,
		Services: all,
	}

	return out, nil
}

// pendingService is what GetByName returns for a name UpsertService recorded
// but that has no container yet - status SCHEDULED_DEPLOYMENT mirrors the
// row postgres/services.go's UpsertService+GetByName pair would report for
// the same not-yet-deployed window in cluster mode.
func (s *dockerServices) pendingService(name string, displayName string) domain.Service {
	if displayName == "" {
		displayName = name
	}

	return domain.Service{
		ServiceBaseInfo: domain.ServiceBaseInfo{
			Name:        name,
			DisplayName: displayName,
			Status:      containerStateToString(""),
		},
		ID:     serviceIDFromName(name),
		Status: pb.DeploymentStatus_SCHEDULED_DEPLOYMENT,
	}
}

// syntheticVelezService mirrors the synthetic "velez" entry that
// listDistinctServices injects into the service list: in single-node mode
// Velez usually runs as a bare binary with no container of its own, so
// GetByName finds nothing to back the detail page. Returning ErrNotFound
// here would 500 a card the service list itself handed out. Reuses
// listDistinctServices so the derived labels (including the bound-resource
// scan) stay identical to the list entry.
func (s *dockerServices) syntheticVelezService(ctx context.Context) (domain.Service, error) {
	all, err := listDistinctServices(ctx, s.docker)
	if err != nil {
		return domain.Service{}, rerrors.Wrap(err, "error deriving synthetic velez service")
	}

	for _, info := range all {
		if info.Name != domain.VelezServiceName {
			continue
		}

		info.Status = containerStateRunning

		svc := domain.Service{
			ServiceBaseInfo: info,
			ID:              serviceIDFromName(domain.VelezServiceName),
			Status:          pb.DeploymentStatus_RUNNING,
		}

		return svc, nil
	}

	return domain.Service{}, user_errors.ErrStorageNotFound
}

// listDistinctServices derives the list of distinct Verv service names from
// live Docker container labels. It is shared between dockerServices.List
// (service listing) and nodes.List (running-services count for the Node
// Health panel) to avoid duplicating the label-filtering logic.
func listDistinctServices(ctx context.Context, docker node_clients.Docker) ([]domain.ServiceBaseInfo, error) {
	listReq := &pb.ListSmerds_Request{}

	containers, err := docker.ListContainers(ctx, listReq, allEnvironments)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	containerNames := make([]string, 0, len(containers))

	for _, c := range containers {
		if len(c.Names) == 0 {
			continue
		}

		containerNames = append(containerNames, strings.TrimPrefix(c.Names[0], "/"))
	}

	seen := make(map[string]bool)

	var all []domain.ServiceBaseInfo

	for _, c := range mainContainersFirst(containers) {
		serviceName := c.Labels[labels.VervServiceLabel]
		if serviceName == "" || seen[serviceName] {
			continue
		}

		seen[serviceName] = true

		isDind := c.Labels[labels.DindInstanceLabel] == boolLabelValue

		info := domain.ServiceBaseInfo{
			Name:        serviceName,
			DisplayName: displayNameOrFallback(c.Labels, serviceName),
			Labels:      classifyDockerService(serviceName, containerNames, isDind),
		}

		all = append(all, info)
	}

	if !seen[domain.VelezServiceName] {
		synthetic := domain.ServiceBaseInfo{
			Name:        domain.VelezServiceName,
			DisplayName: domain.VelezServiceName,
			Labels:      classifyDockerService(domain.VelezServiceName, containerNames, false),
		}

		all = append(all, synthetic)
	}

	return all, nil
}

// classifyDockerService derives a single-node service entry's labels. A
// service owns a bound resource when a sibling container is named
// "<serviceName>_<suffix>" (the same "<service>_<x>" prefix scan
// dockerServiceResourcesStorage.GetResources uses); the suffix is the
// resource type. Otherwise the entry is classified by name alone.
func classifyDockerService(serviceName string, containerNames []string, isDind bool) []string {
	prefix := serviceName + "_"

	for _, name := range containerNames {
		if !strings.HasPrefix(name, prefix) || domain.IsS3WebUiServiceNameOf(serviceName, name) {
			continue
		}

		resourceType := strings.TrimPrefix(name, prefix)

		return domain.ClassifyService(serviceName, resourceType, isDind)
	}

	return domain.ClassifyService(serviceName, "", isDind)
}

// countRunningServices returns the number of distinct running Verv services
// on this node, derived from the same Docker label logic as
// listDistinctServices.
func countRunningServices(ctx context.Context, docker node_clients.Docker) (uint64, error) {
	all, err := listDistinctServices(ctx, docker)
	if err != nil {
		return 0, rerrors.Wrap(err, "error counting running services")
	}

	return uint64(len(all)), nil
}
