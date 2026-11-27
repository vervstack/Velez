package haproxyaas

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/rs/zerolog/log"
	"go.redsock.ru/rerrors"

	"go.vervstack.ru/Velez/internal/api/server/velez_api"
	"go.vervstack.ru/Velez/internal/domain"
	"go.vervstack.ru/Velez/internal/domain/labels"
	"go.vervstack.ru/Velez/internal/service/user_errors"
)

type instanceRef struct {
	name        string
	environment string
	summary     container.Summary
}

func (s *HaproxyServiceImpl) discoverInstances(ctx context.Context, name string) ([]instanceRef, error) {
	environments, err := s.dataStorage.Environments().ListEnvironments(ctx)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing environments")
	}

	var refs []instanceRef

	for _, environment := range environments {
		found, discoverErr := s.discoverInEnvironment(ctx, environment.Name, name)
		if discoverErr != nil {
			return nil, rerrors.Wrap(discoverErr)
		}

		refs = append(refs, found...)
	}

	return refs, nil
}

func (s *HaproxyServiceImpl) discoverInEnvironment(
	ctx context.Context,
	environment, name string,
) ([]instanceRef, error) {
	runtime, err := s.runtimes.Runtime(ctx, environment)
	if err != nil {
		return nil, rerrors.Wrap(err, "error resolving container runtime")
	}

	req := &velez_api.ListSmerds_Request{Environment: environment}
	if name != "" {
		req.Label = map[string]string{labels.HaproxyInstanceLabel: name}
	}

	list, err := runtime.ListContainers(ctx, req)
	if err != nil {
		return nil, rerrors.Wrap(err, "error listing containers")
	}

	var refs []instanceRef

	for _, summary := range list {
		instanceName := summary.Labels[labels.HaproxyInstanceLabel]
		if instanceName == "" || summary.Labels[labels.VervServiceLabel] != domain.HaproxyServiceName(instanceName) {
			continue
		}

		ref := instanceRef{name: instanceName, environment: environment, summary: summary}

		refs = append(refs, ref)
	}

	return refs, nil
}

func (s *HaproxyServiceImpl) findInstance(ctx context.Context, name string) (instanceRef, error) {
	if name == "" {
		return instanceRef{}, rerrors.Wrap(user_errors.ErrHaproxyInstanceNotFound)
	}

	refs, err := s.discoverInstances(ctx, name)
	if err != nil {
		return instanceRef{}, rerrors.Wrap(err)
	}

	for _, ref := range refs {
		if ref.name == name {
			return ref, nil
		}
	}

	return instanceRef{}, rerrors.Wrap(user_errors.ErrHaproxyInstanceNotFound)
}

func (s *HaproxyServiceImpl) ResolveMasterSelection(
	ctx context.Context, name string,
) (domain.HaproxyInstance, string, error) {
	ref, err := s.findInstance(ctx, name)
	if err != nil {
		return domain.HaproxyInstance{}, "", rerrors.Wrap(err)
	}

	if ref.summary.State != "running" {
		return domain.HaproxyInstance{}, "", rerrors.Wrap(user_errors.ErrHaproxyInstanceStopped)
	}

	runtime, err := s.runtimes.Runtime(ctx, ref.environment)
	if err != nil {
		return domain.HaproxyInstance{}, "", rerrors.Wrap(err)
	}

	content, err := runtime.CopyFromContainer(ctx, ref.summary.ID, domain.HaproxyConfigPath)
	if err != nil {
		return domain.HaproxyInstance{}, "", rerrors.Wrap(err, "error reading live HAProxy config")
	}

	master, port := parseConfig(content)
	if master == "" {
		return domain.HaproxyInstance{}, "", rerrors.Wrap(user_errors.ErrHaproxyInvalidMaster)
	}

	if port == 0 {
		port = publishedPort(ref.summary)
	}

	if port == 0 {
		return domain.HaproxyInstance{}, "", rerrors.Wrap(user_errors.ErrHaproxyInvalidPort)
	}

	instance := domain.HaproxyInstance{
		Name: ref.name, Environment: ref.environment, Port: port, MasterPgInstance: master,
	}

	return instance, ref.summary.ID, nil
}

func (s *HaproxyServiceImpl) ListInstances(
	ctx context.Context,
	paging domain.Paging,
) ([]domain.HaproxyInstance, uint64, error) {
	refs, err := s.discoverInstances(ctx, "")
	if err != nil {
		return nil, 0, rerrors.Wrap(err)
	}

	slices.SortFunc(refs, func(a, b instanceRef) int { return strings.Compare(a.name, b.name) })

	total := uint64(len(refs))
	if paging.Offset >= total {
		return nil, total, nil
	}

	refs = refs[paging.Offset:]
	if paging.Limit > 0 && paging.Limit < uint64(len(refs)) {
		refs = refs[:paging.Limit]
	}

	instances := make([]domain.HaproxyInstance, 0, len(refs))
	for _, ref := range refs {
		instances = append(instances, s.describe(ctx, ref))
	}

	return instances, total, nil
}

func (s *HaproxyServiceImpl) describe(ctx context.Context, ref instanceRef) domain.HaproxyInstance {
	instance := domain.HaproxyInstance{
		Name:        ref.name,
		Port:        publishedPort(ref.summary),
		Environment: ref.environment,
		Status:      ref.summary.State,
		CreatedAt:   time.Unix(ref.summary.Created, 0),
	}

	serviceName := domain.HaproxyServiceName(ref.name)

	content, err := s.configResolver.ReadFile(ctx, serviceName, ref.environment, domain.HaproxyConfigPath)
	if err != nil {
		log.Ctx(ctx).Warn().Str("instance", ref.name).Err(err).Msg("error reading HAProxy instance config")

		return instance
	}

	master, port := parseConfig(content)

	instance.MasterPgInstance = master

	if instance.Port == 0 {
		instance.Port = port
	}

	return instance
}

func publishedPort(summary container.Summary) uint32 {
	for _, port := range summary.Ports {
		if port.Type == "tcp" && port.PublicPort != 0 && port.PublicPort == port.PrivatePort {
			return uint32(port.PublicPort)
		}
	}

	return 0
}

func parseConfig(content []byte) (master string, port uint32) {
	for line := range strings.SplitSeq(string(content), "\n") {
		fields := strings.Fields(line)
		switch {
		case len(fields) >= 3 && fields[0] == "server":
			master = fields[1]
		case len(fields) == 2 && fields[0] == "bind":
			value, err := strconv.ParseUint(strings.TrimPrefix(fields[1], "*:"), 10, 16)
			if err == nil {
				port = uint32(value)
			}
		}
	}

	return master, port
}
